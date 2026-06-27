package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// deviceGrantType is the RFC 8628 device-code grant URN.
const deviceGrantType = "urn:ietf:params:oauth:grant-type:device_code"

// deviceAuthResponse is the RFC 8628 §3.2 device-authorization response.
type deviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// loginDevice runs the RFC 8628 device authorization grant: request a code,
// show the user where to approve, then poll the token endpoint.
func loginDevice(ctx context.Context, cfg LoginConfig, md *ASMetadata, httpClient *http.Client) (*Credentials, error) {
	da, err := requestDeviceCode(ctx, httpClient, md.DeviceAuthorizationEndpoint, cfg.ClientID, cfg.Scopes)
	if err != nil {
		return nil, err
	}

	out := output(cfg)
	verifyURL := da.VerificationURIComplete
	if verifyURL == "" {
		verifyURL = da.VerificationURI
	}
	_, _ = fmt.Fprintf(out, "\n  To sign in, visit:\n  %s\n\n  and enter the code:  %s\n\n", da.VerificationURI, da.UserCode)
	if cfg.OpenBrowser != nil {
		_ = cfg.OpenBrowser(verifyURL)
	} else {
		_ = openBrowser(verifyURL)
	}

	interval := time.Duration(da.Interval) * time.Second
	if interval < time.Second {
		interval = 5 * time.Second // RFC 8628 §3.5 default
	}
	deadline := time.Now().Add(time.Duration(maxInt(da.ExpiresIn, 1)) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
		if time.Now().After(deadline) {
			return nil, &OAuthError{Code: "expired_token", Description: "device code expired before approval"}
		}

		tok, err := postToken(ctx, httpClient, md.TokenEndpoint, url.Values{
			"grant_type":  {deviceGrantType},
			"device_code": {da.DeviceCode},
			"client_id":   {cfg.ClientID},
		})
		if err == nil {
			return tok.credentials(md.TokenEndpoint, cfg.ClientID, time.Now()), nil
		}

		var oe *OAuthError
		if !asOAuthError(err, &oe) {
			return nil, err
		}
		switch oe.Code {
		case "authorization_pending":
			// keep polling at the current interval
		case "slow_down":
			interval += 5 * time.Second // RFC 8628 §3.5
		default: // access_denied, expired_token, anything else terminal
			return nil, oe
		}
	}
}

func requestDeviceCode(ctx context.Context, httpClient *http.Client, endpoint, clientID string, scopes []string) (*deviceAuthResponse, error) {
	form := url.Values{"client_id": {clientID}}
	if len(scopes) > 0 {
		form.Set("scope", strings.Join(scopes, " "))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("microwave/auth: build device request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("microwave/auth: device request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, parseOAuthError(resp.StatusCode, body)
	}
	var da deviceAuthResponse
	if err := json.Unmarshal(body, &da); err != nil {
		return nil, fmt.Errorf("microwave/auth: decode device response: %w", err)
	}
	if da.DeviceCode == "" || da.UserCode == "" {
		return nil, fmt.Errorf("microwave/auth: device response missing device_code/user_code")
	}
	return &da, nil
}

func asOAuthError(err error, target **OAuthError) bool {
	for e := err; e != nil; {
		if oe, ok := e.(*OAuthError); ok {
			*target = oe
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := e.(unwrapper)
		if !ok {
			return false
		}
		e = u.Unwrap()
	}
	return false
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
