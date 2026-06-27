package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// refreshSkew is how long before exp a token is treated as already expired, so
// callers refresh proactively rather than racing a request against the clock.
const refreshSkew = 60 * time.Second

// Credentials is a minted Microwave session plus everything needed to refresh
// it without another interactive login. It is JSON-serialisable for storage.
type Credentials struct {
	AccessToken   string    `json:"access_token"`
	TokenType     string    `json:"token_type,omitempty"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
	RefreshToken  string    `json:"refresh_token,omitempty"`
	TokenEndpoint string    `json:"token_endpoint"`
	ClientID      string    `json:"client_id"`
}

// Expired reports whether the access token is at or near expiry (within the
// refresh skew). A zero ExpiresAt means "no known expiry" and is never expired.
func (c *Credentials) Expired() bool { return c.expiredAt(time.Now()) }

func (c *Credentials) expiredAt(now time.Time) bool {
	if c.ExpiresAt.IsZero() {
		return false
	}
	return !now.Before(c.ExpiresAt.Add(-refreshSkew))
}

// tokenResponse is the RFC 6749 §5.1 token-endpoint success body.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func (t tokenResponse) credentials(tokenEndpoint, clientID string, now time.Time) *Credentials {
	c := &Credentials{
		AccessToken:   t.AccessToken,
		TokenType:     t.TokenType,
		RefreshToken:  t.RefreshToken,
		TokenEndpoint: tokenEndpoint,
		ClientID:      clientID,
	}
	if t.ExpiresIn > 0 {
		c.ExpiresAt = now.Add(time.Duration(t.ExpiresIn) * time.Second)
	}
	return c
}

// OAuthError is an RFC 6749 §5.2 token-endpoint error. The device-flow poller
// inspects Code for `authorization_pending`, `slow_down`, `access_denied`, and
// `expired_token`.
type OAuthError struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
	Status      int    `json:"-"`
}

func (e *OAuthError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("microwave/auth: %s: %s", e.Code, e.Description)
	}
	return "microwave/auth: " + e.Code
}

func parseOAuthError(status int, body []byte) error {
	var oe OAuthError
	if err := json.Unmarshal(body, &oe); err == nil && oe.Code != "" {
		oe.Status = status
		return &oe
	}
	return fmt.Errorf("microwave/auth: token endpoint status %d: %s", status, strings.TrimSpace(string(body)))
}

// postToken POSTs a form-encoded grant to the token endpoint and decodes the
// RFC 6749 §5.1 success body, returning a typed *OAuthError on §5.2 failures.
// Shared by the loopback, device, and refresh paths.
func postToken(ctx context.Context, httpClient *http.Client, tokenEndpoint string, form url.Values) (tokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return tokenResponse{}, fmt.Errorf("microwave/auth: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("microwave/auth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return tokenResponse{}, parseOAuthError(resp.StatusCode, body)
	}
	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return tokenResponse{}, fmt.Errorf("microwave/auth: decode token response: %w", err)
	}
	if tok.AccessToken == "" {
		return tokenResponse{}, fmt.Errorf("microwave/auth: token response missing access_token")
	}
	return tok, nil
}

// Refresh exchanges the stored refresh token for a fresh access token
// (RFC 6749 §6), updating c in place. Returns an error (re-login required) when
// there is no refresh token.
func (c *Credentials) Refresh(ctx context.Context, httpClient *http.Client) error {
	return c.refreshAt(ctx, httpClient, time.Now())
}

func (c *Credentials) refreshAt(ctx context.Context, httpClient *http.Client, now time.Time) error {
	if c.RefreshToken == "" {
		return fmt.Errorf("microwave/auth: session expired and no refresh token; run login again")
	}
	if httpClient == nil {
		httpClient = defaultHTTPClient()
	}
	tok, err := postToken(ctx, httpClient, c.TokenEndpoint, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {c.RefreshToken},
		"client_id":     {c.ClientID},
	})
	if err != nil {
		return err
	}
	next := tok.credentials(c.TokenEndpoint, c.ClientID, now)
	if next.RefreshToken == "" {
		next.RefreshToken = c.RefreshToken // server didn't rotate; keep the old one
	}
	*c = *next
	return nil
}

// TokenStore persists credentials between CLI invocations.
type TokenStore interface {
	Load() (*Credentials, error)
	Save(*Credentials) error
	Clear() error
}

// FileStore is the default TokenStore: a 0600 JSON file under the user's config
// directory. A nil/zero return from Load means "no stored credentials".
type FileStore struct{ Path string }

// DefaultFileStore returns a FileStore at <user-config-dir>/<app>/credentials.json.
func DefaultFileStore(app string) (FileStore, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return FileStore{}, fmt.Errorf("microwave/auth: resolve config dir: %w", err)
	}
	return FileStore{Path: filepath.Join(dir, app, "credentials.json")}, nil
}

func (s FileStore) Load() (*Credentials, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("microwave/auth: read credentials: %w", err)
	}
	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("microwave/auth: parse credentials: %w", err)
	}
	return &c, nil
}

func (s FileStore) Save(c *Credentials) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return fmt.Errorf("microwave/auth: create config dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("microwave/auth: encode credentials: %w", err)
	}
	if err := os.WriteFile(s.Path, data, 0o600); err != nil {
		return fmt.Errorf("microwave/auth: write credentials: %w", err)
	}
	return nil
}

func (s FileStore) Clear() error {
	if err := os.Remove(s.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("microwave/auth: clear credentials: %w", err)
	}
	return nil
}
