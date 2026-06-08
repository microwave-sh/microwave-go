package auth

import (
	"fmt"
	"net/http"
	"time"
)

// Option configures an auth.Client. Options compose left-to-right.
type Option func(*clientConfig)

type clientConfig struct {
	endpoint   string
	httpClient *http.Client
}

// WithEndpoint overrides the Auth plane base URL.
func WithEndpoint(endpoint string) Option {
	return func(c *clientConfig) {
		c.endpoint = endpoint
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
		endpoint:   DefaultEndpoint,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.endpoint == "" {
		return nil, fmt.Errorf("microwave/auth: endpoint is required")
	}
	if cfg.httpClient == nil {
		cfg.httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return cfg, nil
}
