package management

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Option configures a Client. Options compose: later options override earlier
// ones, env-derived defaults apply when no option sets a value.
type Option func(*clientConfig)

type clientConfig struct {
	endpoint      string
	managementKey string
	bearerSource  func(context.Context) (string, error)
	workspaceID   string
	httpClient    *http.Client
}

// WithEndpoint overrides the API base URL (default https://api.microwave.sh).
// Use for self-hosted deployments or local development against a compose stack.
func WithEndpoint(endpoint string) Option {
	return func(c *clientConfig) {
		c.endpoint = endpoint
	}
}

// WithManagementKey sets the bearer credential. Required either via this option
// or the MICROWAVE_MANAGEMENT_KEY environment variable.
func WithManagementKey(key string) Option {
	return func(c *clientConfig) {
		c.managementKey = key
	}
}

// WithBearerSource supplies the bearer credential dynamically, called once per
// request. Use it with auth.TokenSource so the client transparently refreshes
// an interactive login's session token. Mutually exclusive with
// WithManagementKey / MICROWAVE_MANAGEMENT_KEY.
func WithBearerSource(src func(context.Context) (string, error)) Option {
	return func(c *clientConfig) {
		c.bearerSource = src
	}
}

// WithWorkspaceID pins requests to a specific workspace. If unset, the
// management key's owning workspace is used. Set this when a single key has
// access to multiple workspaces (uncommon — most keys are workspace-scoped).
func WithWorkspaceID(workspaceID string) Option {
	return func(c *clientConfig) {
		c.workspaceID = workspaceID
	}
}

// WithHTTPClient supplies a custom *http.Client (e.g. one with a shared
// transport pool, an OpenTelemetry round-tripper, or a stricter timeout).
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = httpClient
	}
}

func resolveConfig(opts []Option) (*clientConfig, error) {
	cfg := &clientConfig{
		endpoint:      DefaultEndpoint,
		managementKey: os.Getenv("MICROWAVE_MANAGEMENT_KEY"),
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.managementKey == "" && cfg.bearerSource == nil {
		return nil, fmt.Errorf("microwave: a credential is required (WithManagementKey, MICROWAVE_MANAGEMENT_KEY, or WithBearerSource)")
	}
	if cfg.endpoint == "" {
		return nil, fmt.Errorf("microwave: endpoint is required")
	}
	if cfg.httpClient == nil {
		cfg.httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return cfg, nil
}
