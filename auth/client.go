// Package auth provides a Go client for the Microwave Auth plane
// (https://auth.microwave.sh) — the public, unauthenticated surface that
// validates inbound OIDC assertions and mints Microwave session JWTs.
//
// The typical consumer is a federated workload (a Terraform Cloud run, a
// GitHub Actions job, an internal service with workload identity) that
// holds an OIDC token from its native issuer and needs to redeem it for a
// Microwave session before calling the Management API.
//
//	authClient, err := auth.NewClient()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	result, err := authClient.TokenExchange.Redeem(ctx, "ex_tfc_admin", tfcToken)
//	if err != nil || !result.Valid {
//	    log.Fatalf("exchange failed: %v code=%s", err, result.Code)
//	}
//
//	mgmt, _ := microwave.NewClient(microwave.WithManagementKey(result.JWT))
//
// The returned JWT is short-lived; long-running consumers should re-redeem
// on 401 or before the token's exp claim passes.
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DefaultEndpoint is the production Auth plane base URL.
const DefaultEndpoint = "https://auth.microwave.sh"

// Version mirrors the parent microwave-go module version. Bumped together
// because the auth and management surfaces ship as one release artefact.
const Version = "0.12.0"

// Client is the Auth plane client. The Auth plane is unauthenticated at the
// HTTP layer — the OIDC assertion is the only credential — so this client
// holds no API key.
type Client struct {
	cfg *clientConfig

	TokenExchange *TokenExchangeService
}

// NewClient creates a new Auth plane client. The endpoint defaults to
// https://auth.microwave.sh; override with WithEndpoint for self-hosted
// deployments or local development.
func NewClient(opts ...Option) (*Client, error) {
	cfg, err := resolveConfig(opts)
	if err != nil {
		return nil, err
	}
	c := &Client{cfg: cfg}
	c.TokenExchange = &TokenExchangeService{client: c}
	return c, nil
}

// doRequest is the shared transport. Both the JSON marshal and a best-effort
// error decode live here so service methods stay focused on the URL + body
// pair they need to send.
func (c *Client) doRequest(ctx context.Context, method, path string, body, result any) error {
	endpoint := strings.TrimRight(c.cfg.endpoint, "/") + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("microwave/auth: marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("microwave/auth: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "microwave-go-auth/"+Version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("microwave/auth: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return decodeError(resp)
	}
	if resp.StatusCode == http.StatusNoContent || result == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("microwave/auth: decode response: %w", err)
	}
	return nil
}
