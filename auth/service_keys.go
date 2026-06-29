package auth

import (
	"context"
	"net/http"
	"time"
)

// VerifyResult is the decoded response from verifying an opaque key. Valid==false
// is a denial (Code names the reason), not a transport error. Mirrors the
// server's dto.VerifyKeyResult.
type VerifyResult struct {
	Valid     bool           `json:"valid"`
	Code      string         `json:"code,omitempty"`
	KeyID     string         `json:"key_id,omitempty"`
	Subject   string         `json:"subject,omitempty"`
	Scopes    []string       `json:"scopes,omitempty"`
	Claims    map[string]any `json:"claims,omitempty"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	JWT       string         `json:"jwt,omitempty"`
}

// KeysService is the auth-plane surface for verifying opaque keys.
type KeysService struct {
	client *Client
}

// Verify checks an opaque key against Microwave. A denied key returns a
// result with Valid==false and a non-nil error only on transport failure.
func (s *KeysService) Verify(ctx context.Context, key string) (*VerifyResult, error) {
	body := struct {
		Key string `json:"key"`
	}{Key: key}
	var out VerifyResult
	if err := s.client.doRequest(ctx, http.MethodPost, "/verify", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
