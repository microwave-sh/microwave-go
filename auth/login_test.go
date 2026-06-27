package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockAS is a minimal brokered authorization server for tests: it implements
// RFC 8414 metadata, an /authorize that "instantly authenticates" and redirects
// to the loopback with a code, a /token that verifies PKCE and mints, and the
// device endpoints.
type mockAS struct {
	*httptest.Server
	mu          sync.Mutex
	challenges  map[string]string // authz code -> code_challenge
	issParam    bool
	devicePolls int
}

func newMockAS(t *testing.T, issParam bool) *mockAS {
	t.Helper()
	m := &mockAS{challenges: map[string]string{}, issParam: issParam}
	mux := http.NewServeMux()

	mux.HandleFunc("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(ASMetadata{
			Issuer:                        m.URL,
			AuthorizationEndpoint:         m.URL + "/authorize",
			TokenEndpoint:                 m.URL + "/token",
			DeviceAuthorizationEndpoint:   m.URL + "/device_authorization",
			GrantTypesSupported:           []string{"authorization_code", deviceGrantType},
			CodeChallengeMethodsSupported: []string{"S256"},
			IssParameterSupported:         m.issParam,
		})
	})

	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("response_type") != "code" || q.Get("client_id") == "" {
			http.Error(w, "bad authorize", http.StatusBadRequest)
			return
		}
		if q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") == "" {
			http.Error(w, "pkce required", http.StatusBadRequest)
			return
		}
		code := "code_" + q.Get("state")
		m.mu.Lock()
		m.challenges[code] = q.Get("code_challenge")
		m.mu.Unlock()
		redirect := q.Get("redirect_uri") + "?code=" + code + "&state=" + q.Get("state")
		if m.issParam {
			redirect += "&iss=" + m.URL
		}
		http.Redirect(w, r, redirect, http.StatusFound)
	})

	mux.HandleFunc("/device_authorization", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(deviceAuthResponse{
			DeviceCode:              "dev_abc",
			UserCode:                "WXYZ-1234",
			VerificationURI:         m.URL + "/activate",
			VerificationURIComplete: m.URL + "/activate?user_code=WXYZ-1234",
			ExpiresIn:               300,
			Interval:                1,
		})
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		switch r.Form.Get("grant_type") {
		case "authorization_code":
			m.mu.Lock()
			challenge, ok := m.challenges[r.Form.Get("code")]
			m.mu.Unlock()
			if !ok {
				writeTokenErr(w, "invalid_grant", "unknown code")
				return
			}
			sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
			if base64.RawURLEncoding.EncodeToString(sum[:]) != challenge {
				writeTokenErr(w, "invalid_grant", "pkce mismatch")
				return
			}
			writeToken(w, "sess-jwt", "refresh-1", 900)
		case deviceGrantType:
			m.mu.Lock()
			m.devicePolls++
			n := m.devicePolls
			m.mu.Unlock()
			if n < 2 {
				writeTokenErr(w, "authorization_pending", "")
				return
			}
			writeToken(w, "dev-sess-jwt", "", 900)
		case "refresh_token":
			if r.Form.Get("refresh_token") != "refresh-1" {
				writeTokenErr(w, "invalid_grant", "bad refresh")
				return
			}
			writeToken(w, "sess-jwt-2", "refresh-2", 900)
		default:
			writeTokenErr(w, "unsupported_grant_type", "")
		}
	})

	m.Server = httptest.NewServer(mux)
	t.Cleanup(m.Close)
	return m
}

func writeToken(w http.ResponseWriter, access, refresh string, expiresIn int) {
	w.Header().Set("Content-Type", "application/json")
	out := map[string]any{"access_token": access, "token_type": "Bearer", "expires_in": expiresIn}
	if refresh != "" {
		out["refresh_token"] = refresh
	}
	_ = json.NewEncoder(w).Encode(out)
}

func writeTokenErr(w http.ResponseWriter, code, desc string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "error_description": desc})
}

// browserToCallback returns an OpenBrowser that follows the authorize redirect
// to the loopback callback, simulating a user approving in the browser.
func browserToCallback() func(string) error {
	return func(u string) error {
		go func() {
			resp, err := http.Get(u) //nolint:noctx
			if err == nil {
				_ = resp.Body.Close()
			}
		}()
		return nil
	}
}

func TestPKCE_S256(t *testing.T) {
	p, err := newPKCE()
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(p.verifier))
	if base64.RawURLEncoding.EncodeToString(sum[:]) != p.challenge {
		t.Fatal("challenge is not S256(verifier)")
	}
}

func TestLogin_LoopbackHappyPath(t *testing.T) {
	m := newMockAS(t, true)
	creds, err := Login(context.Background(), LoginConfig{
		MetadataURL: m.URL + "/.well-known/oauth-authorization-server",
		ClientID:    "spec_cli",
		Mode:        LoginLoopback,
		OpenBrowser: browserToCallback(),
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if creds.AccessToken != "sess-jwt" {
		t.Fatalf("access_token = %q", creds.AccessToken)
	}
	if creds.RefreshToken != "refresh-1" {
		t.Fatalf("refresh_token = %q", creds.RefreshToken)
	}
	if creds.Expired() {
		t.Fatal("fresh token reports expired")
	}
}

func TestLogin_StateMismatchRejected(t *testing.T) {
	m := newMockAS(t, true)
	// Opener hits the loopback callback directly with a forged state (CSRF).
	opener := func(u string) error {
		parsed, err := url.Parse(u)
		if err != nil {
			return err
		}
		redirectURI := parsed.Query().Get("redirect_uri")
		go func() {
			resp, gerr := http.Get(redirectURI + "?code=code_x&state=WRONG") //nolint:noctx
			if gerr == nil {
				_ = resp.Body.Close()
			}
		}()
		return nil
	}
	_, err := Login(context.Background(), LoginConfig{
		MetadataURL: m.URL + "/.well-known/oauth-authorization-server",
		ClientID:    "spec_cli",
		Mode:        LoginLoopback,
		OpenBrowser: opener,
	})
	if err == nil || !strings.Contains(err.Error(), "state mismatch") {
		t.Fatalf("expected state mismatch rejection, got %v", err)
	}
}

func TestLogin_IssMismatchRejected(t *testing.T) {
	m := newMockAS(t, true)
	// Force an iss mismatch by pointing the metadata issuer elsewhere via a
	// wrapper server that rewrites the metadata.
	wrap := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(ASMetadata{
			Issuer:                "https://evil.example", // != the iss the AS returns
			AuthorizationEndpoint: m.URL + "/authorize",
			TokenEndpoint:         m.URL + "/token",
			IssParameterSupported: true,
		})
	}))
	defer wrap.Close()

	_, err := Login(context.Background(), LoginConfig{
		MetadataURL: wrap.URL,
		ClientID:    "spec_cli",
		Mode:        LoginLoopback,
		OpenBrowser: browserToCallback(),
	})
	if err == nil {
		t.Fatal("expected issuer-mismatch rejection")
	}
}

func TestLogin_DeviceFallbackAndPending(t *testing.T) {
	m := newMockAS(t, false)
	creds, err := Login(context.Background(), LoginConfig{
		MetadataURL: m.URL + "/.well-known/oauth-authorization-server",
		ClientID:    "spec_cli",
		Mode:        LoginDevice,
		OpenBrowser: func(string) error { return nil },
	})
	if err != nil {
		t.Fatalf("device Login: %v", err)
	}
	if creds.AccessToken != "dev-sess-jwt" {
		t.Fatalf("access_token = %q", creds.AccessToken)
	}
	if m.devicePolls < 2 {
		t.Fatalf("expected to poll through authorization_pending, polls=%d", m.devicePolls)
	}
}

func TestCredentials_Refresh(t *testing.T) {
	m := newMockAS(t, true)
	c := &Credentials{
		AccessToken:   "sess-jwt",
		RefreshToken:  "refresh-1",
		TokenEndpoint: m.URL + "/token",
		ClientID:      "spec_cli",
		ExpiresAt:     time.Now().Add(-time.Minute), // already expired
	}
	if !c.Expired() {
		t.Fatal("expected expired")
	}
	if err := c.Refresh(context.Background(), nil); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if c.AccessToken != "sess-jwt-2" || c.RefreshToken != "refresh-2" {
		t.Fatalf("after refresh: %q / %q", c.AccessToken, c.RefreshToken)
	}
	if c.Expired() {
		t.Fatal("refreshed token reports expired")
	}
}

func TestTokenSource_RefreshesAndPersists(t *testing.T) {
	m := newMockAS(t, true)
	store := FileStore{Path: filepath.Join(t.TempDir(), "creds.json")}
	creds := &Credentials{
		AccessToken:   "old",
		RefreshToken:  "refresh-1",
		TokenEndpoint: m.URL + "/token",
		ClientID:      "spec_cli",
		ExpiresAt:     time.Now().Add(-time.Minute), // expired → forces refresh
	}
	src := NewTokenSource(creds, store, nil)
	tok, err := src.Token(context.Background())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if tok != "sess-jwt-2" {
		t.Fatalf("token = %q, want refreshed sess-jwt-2", tok)
	}
	saved, _ := store.Load()
	if saved == nil || saved.AccessToken != "sess-jwt-2" {
		t.Fatalf("refreshed token not persisted: %#v", saved)
	}
}

func TestFileStore_RoundTrip(t *testing.T) {
	s := FileStore{Path: filepath.Join(t.TempDir(), "nested", "credentials.json")}
	if c, err := s.Load(); err != nil || c != nil {
		t.Fatalf("empty load = %v, %v", c, err)
	}
	in := &Credentials{AccessToken: "a", RefreshToken: "r", ClientID: "spec", TokenEndpoint: "https://x/token"}
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "a" || got.RefreshToken != "r" {
		t.Fatalf("round-trip = %#v", got)
	}
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	if c, _ := s.Load(); c != nil {
		t.Fatal("expected nil after Clear")
	}
}
