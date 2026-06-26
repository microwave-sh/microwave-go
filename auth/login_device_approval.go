package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// deviceApprovalRequestResult is the response from POST /auth/device.
type deviceApprovalRequestResult struct {
	DeviceCode   string `json:"device_code"`
	UserCode     string `json:"user_code"`
	AuthorizeURL string `json:"authorize_url"`
	ExpiresIn    int    `json:"expires_in"`
	Interval     int    `json:"interval"`
}

// deviceApprovalPollResult is the response from POST /auth/device/token.
type deviceApprovalPollResult struct {
	Status    string `json:"status"`
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

// loginDeviceApproval drives the management device-approval flow: request a
// device code, surface the console authorize URL + user code, then poll until
// the operator approves it in the console. Authorization happens there against
// the operator's session (which carries their per-operator permissions), so the
// flow needs no client-id and no PKCE.
func loginDeviceApproval(ctx context.Context, cfg LoginConfig, httpClient *http.Client) (*Credentials, error) {
	base := strings.TrimRight(cfg.DeviceApprovalURL, "/")
	out := output(cfg)

	var da deviceApprovalRequestResult
	if err := postJSONInto(ctx, httpClient, base+"/auth/device", map[string]any{}, &da); err != nil {
		return nil, err
	}
	if da.DeviceCode == "" || da.AuthorizeURL == "" {
		return nil, fmt.Errorf("microwave/auth: device request missing device_code/authorize_url")
	}

	fmt.Fprintf(out, "\n  To approve this login, visit:\n  %s\n", da.AuthorizeURL)
	if da.UserCode != "" {
		fmt.Fprintf(out, "\n  and enter the code:  %s\n\n", da.UserCode)
	}
	if cfg.OpenBrowser != nil {
		_ = cfg.OpenBrowser(da.AuthorizeURL)
	} else {
		_ = openBrowser(da.AuthorizeURL)
	}

	interval := time.Duration(da.Interval) * time.Second
	if interval < time.Second {
		interval = 2 * time.Second
	}
	deadline := time.Now().Add(time.Duration(maxInt(da.ExpiresIn, 60)) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
		if time.Now().After(deadline) {
			return nil, &OAuthError{Code: "expired_token", Description: "device code expired before approval"}
		}
		var p deviceApprovalPollResult
		if err := postJSONInto(ctx, httpClient, base+"/auth/device/token", map[string]string{"device_code": da.DeviceCode}, &p); err != nil {
			return nil, err
		}
		switch p.Status {
		case "approved":
			if p.Token == "" {
				return nil, fmt.Errorf("microwave/auth: approval returned no token")
			}
			creds := &Credentials{AccessToken: p.Token, TokenType: "Bearer"}
			if p.ExpiresIn > 0 {
				creds.ExpiresAt = time.Now().Add(time.Duration(p.ExpiresIn) * time.Second)
			}
			return creds, nil
		case "pending":
			continue
		case "denied":
			return nil, &OAuthError{Code: "access_denied", Description: "device login was denied"}
		case "expired":
			return nil, &OAuthError{Code: "expired_token", Description: "device code expired before approval"}
		default:
			return nil, fmt.Errorf("microwave/auth: unexpected device status %q", p.Status)
		}
	}
}

// postJSONInto POSTs body as JSON and decodes the JSON response into dst; a
// non-2xx status is parsed as an OAuthError.
func postJSONInto(ctx context.Context, httpClient *http.Client, endpoint string, body, dst any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("microwave/auth: marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("microwave/auth: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("microwave/auth: request: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseOAuthError(resp.StatusCode, data)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("microwave/auth: decode response: %w", err)
	}
	return nil
}
