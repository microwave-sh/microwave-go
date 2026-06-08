// Package management provides a Go client for the Microwave Management API
// (https://api.microwave.sh). It covers the workspace-admin surface: permission
// sets, signing key sets, key specifications, and trust exchanges.
//
//	client, err := management.NewClient(
//	    management.WithManagementKey("mw_live_..."),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	ps, err := client.PermissionSets.Create(ctx, &management.PermissionSetInput{
//	    Name: "deployer",
//	    Permissions: []management.PermissionInput{
//	        {Resource: "deploys", Action: "create"},
//	        {Resource: "blobs", Action: "upload"},
//	    },
//	})
package management

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultEndpoint is the production Management API base URL.
const DefaultEndpoint = "https://api.microwave.sh"

// Client is the Microwave Management API client. Field references like
// client.PermissionSets give scoped per-resource service surfaces.
type Client struct {
	cfg *clientConfig

	PermissionSets  *PermissionSetsService
	SigningKeySets  *SigningKeySetsService
	KeySpecs        *KeySpecsService
	TrustExchanges  *TrustExchangesService
}

// NewClient creates a new Management API client. A management key must be
// supplied via WithManagementKey or the MICROWAVE_MANAGEMENT_KEY environment
// variable. Without one, NewClient returns an error rather than producing a
// client that will 401 on first call.
func NewClient(opts ...Option) (*Client, error) {
	cfg, err := resolveConfig(opts)
	if err != nil {
		return nil, err
	}
	c := &Client{cfg: cfg}
	c.PermissionSets = &PermissionSetsService{client: c}
	c.SigningKeySets = &SigningKeySetsService{client: c}
	c.KeySpecs = &KeySpecsService{client: c}
	c.TrustExchanges = &TrustExchangesService{client: c}
	return c, nil
}

// doRequest performs an authenticated HTTP request and decodes the JSON
// response into result (or discards the body if result is nil). All four
// services route through this single transport so headers, error handling,
// and JSON contracts are uniform.
func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body, result any) error {
	endpoint := strings.TrimRight(c.cfg.endpoint, "/") + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("microwave: marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("microwave: build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.managementKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("API-Version", APIVersion)
	req.Header.Set("User-Agent", "microwave-go-management/"+Version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.workspaceID != "" {
		req.Header.Set("X-Microwave-Workspace", c.cfg.workspaceID)
	}

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("microwave: request failed: %w", err)
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
		return fmt.Errorf("microwave: decode response: %w", err)
	}
	return nil
}

// Time is exported solely so service files can format request times without
// pulling time across files; keeps the dependency graph readable.
type Time = time.Time
