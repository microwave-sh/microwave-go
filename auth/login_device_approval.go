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
	DeviceCode string `json:"device_code"`
	UserCode   string `json:"user_code"`
	// VerificationURI is the static console page the operator opens to approve.
	// It carries no secret; the operator types UserCode there by hand. The
	// device_code stays here in the client and is only ever sent back to the
	// poll endpoint.
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// deviceApprovalPollResult is the response from POST /auth/device/token.
type deviceApprovalPollResult struct {
	Status    string `json:"status"`
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

// loginDeviceApproval drives the management device-approval flow: request a
// device code, tell the operator to open the static verification page and type
// the user code there, then poll until they approve it. Authorization happens
// against the operator's console session (which carries their per-operator
// permissions), so the flow needs no client-id and no PKCE. The device_code
// secret never leaves this client — the operator only ever handles the short
// user code, which is the RFC 8628 anti-phishing step.
func loginDeviceApproval(ctx context.Context, cfg LoginConfig, httpClient *http.Client) (*Credentials, error) {
	base := strings.TrimRight(cfg.DeviceApprovalURL, "/")
	out := output(cfg)

	// Empty body selects the server's SYSTEM CLI exchange; a product minting
	// through its own exchange names it so the server resolves the right one.
	reqBody := map[string]any{}
	if strings.TrimSpace(cfg.TrustExchangeID) != "" {
		reqBody["trust_exchange_id"] = cfg.TrustExchangeID
	}

	reportBegin(cfg, "Starting device authorization")
	var da deviceApprovalRequestResult
	if err := postJSONInto(ctx, httpClient, base+"/auth/device", reqBody, &da); err != nil {
		reportFail(cfg, "Could not start device authorization")
		return nil, err
	}
	if da.DeviceCode == "" || da.VerificationURI == "" || da.UserCode == "" {
		reportFail(cfg, "Could not start device authorization")
		return nil, fmt.Errorf("microwave/auth: device request missing device_code/verification_uri/user_code")
	}
	reportSucceed(cfg, "Device authorization started")

	fmt.Fprintf(out, "\n  To sign in, open:\n  %s\n\n  and enter the code:  %s\n\n", da.VerificationURI, da.UserCode)
	if cfg.OpenBrowser != nil {
		_ = cfg.OpenBrowser(da.VerificationURI)
	} else {
		_ = openBrowser(da.VerificationURI)
	}

	interval := time.Duration(da.Interval) * time.Second
	if interval < time.Second {
		interval = 2 * time.Second
	}
	deadline := time.Now().Add(time.Duration(maxInt(da.ExpiresIn, 60)) * time.Second)

	reportBegin(cfg, "Waiting for approval")
	for {
		select {
		case <-ctx.Done():
			reportFail(cfg, "Login cancelled")
			return nil, ctx.Err()
		case <-time.After(interval):
		}
		if time.Now().After(deadline) {
			reportFail(cfg, "Login timed out")
			return nil, &OAuthError{Code: "expired_token", Description: "device code expired before approval"}
		}
		var p deviceApprovalPollResult
		if err := postJSONInto(ctx, httpClient, base+"/auth/device/token", map[string]string{"device_code": da.DeviceCode}, &p); err != nil {
			reportFail(cfg, "Login failed")
			return nil, err
		}
		switch p.Status {
		case "approved":
			if p.Token == "" {
				reportFail(cfg, "Login failed")
				return nil, fmt.Errorf("microwave/auth: approval returned no token")
			}
			creds := &Credentials{AccessToken: p.Token, TokenType: "Bearer"}
			if p.ExpiresIn > 0 {
				creds.ExpiresAt = time.Now().Add(time.Duration(p.ExpiresIn) * time.Second)
			}
			reportSucceed(cfg, "Approved")
			return creds, nil
		case "pending":
			continue
		case "denied":
			reportFail(cfg, "Login denied")
			return nil, &OAuthError{Code: "access_denied", Description: "device login was denied"}
		case "expired":
			reportFail(cfg, "Login expired")
			return nil, &OAuthError{Code: "expired_token", Description: "device code expired before approval"}
		default:
			reportFail(cfg, "Login failed")
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
