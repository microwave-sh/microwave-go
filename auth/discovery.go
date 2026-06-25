package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ASMetadata is the subset of RFC 8414 authorization-server metadata the login
// flow consumes. The full document may carry more fields; unknown keys are
// ignored.
type ASMetadata struct {
	Issuer                        string   `json:"issuer"`
	AuthorizationEndpoint         string   `json:"authorization_endpoint"`
	TokenEndpoint                 string   `json:"token_endpoint"`
	DeviceAuthorizationEndpoint   string   `json:"device_authorization_endpoint"`
	GrantTypesSupported           []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
	IssParameterSupported         bool     `json:"authorization_response_iss_parameter_supported"`
}

// supportsDeviceGrant reports whether the AS advertises the device flow.
func (m *ASMetadata) supportsDeviceGrant() bool {
	return m.DeviceAuthorizationEndpoint != ""
}

// fetchMetadata loads RFC 8414 authorization-server metadata from metadataURL.
func fetchMetadata(ctx context.Context, httpClient *http.Client, metadataURL string) (*ASMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("microwave/auth: build metadata request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("microwave/auth: fetch metadata: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("microwave/auth: metadata %s: status %d", metadataURL, resp.StatusCode)
	}
	var md ASMetadata
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&md); err != nil {
		return nil, fmt.Errorf("microwave/auth: decode metadata: %w", err)
	}
	if md.TokenEndpoint == "" {
		return nil, fmt.Errorf("microwave/auth: metadata %s has no token_endpoint", metadataURL)
	}
	return &md, nil
}
