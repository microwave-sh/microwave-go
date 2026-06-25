package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

// TokenSource yields a currently-valid access token, refreshing (and
// persisting) the credentials when they reach expiry. Pass its Token method to
// management.WithBearerSource so the Management client auto-refreshes an
// interactive login's session without the caller threading tokens by hand:
//
//	creds, _ := auth.Login(ctx, cfg)
//	src := auth.NewTokenSource(creds, cfg.Store, nil)
//	mgmt, _ := management.NewClient(management.WithBearerSource(src.Token))
type TokenSource struct {
	mu         sync.Mutex
	creds      *Credentials
	store      TokenStore
	httpClient *http.Client
}

// NewTokenSource wraps credentials with optional persistence. A nil httpClient
// uses a default client for refreshes.
func NewTokenSource(creds *Credentials, store TokenStore, httpClient *http.Client) *TokenSource {
	if httpClient == nil {
		httpClient = defaultHTTPClient()
	}
	return &TokenSource{creds: creds, store: store, httpClient: httpClient}
}

// Token returns a valid access token, refreshing in place (and saving to the
// store, if set) when the cached token is at or near expiry. Safe for
// concurrent use.
func (s *TokenSource) Token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.creds == nil {
		return "", fmt.Errorf("microwave/auth: no credentials; run login first")
	}
	if s.creds.Expired() {
		if err := s.creds.Refresh(ctx, s.httpClient); err != nil {
			return "", err
		}
		if s.store != nil {
			if err := s.store.Save(s.creds); err != nil {
				return "", err
			}
		}
	}
	return s.creds.AccessToken, nil
}
