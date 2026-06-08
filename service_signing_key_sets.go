package microwave

import (
	"context"
	"net/http"
	"net/url"
)

// SigningKeySetsService is the Management API surface for signing key sets.
// The composite (kind, name) primary key means Get/Update/Delete take two
// segments — that's why these methods accept both.
type SigningKeySetsService struct {
	client *Client
}

// Create posts a new signing key set. Algorithm + Kind are immutable once set;
// to "change" them, Delete and re-Create.
func (s *SigningKeySetsService) Create(ctx context.Context, input *SigningKeySetInput) (*SigningKeySet, error) {
	var out SigningKeySet
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/signing-key-sets", nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a signing key set by (kind, name).
func (s *SigningKeySetsService) Get(ctx context.Context, kind SigningKeySetKind, name string) (*SigningKeySet, error) {
	var out SigningKeySet
	path := "/api/signing-key-sets/" + url.PathEscape(string(kind)) + "/" + url.PathEscape(name)
	if err := s.client.doRequest(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete soft-deletes a signing key set. The DeletedAt timestamp on the
// returned record is non-nil after this call; signed tokens still verify
// against keys in the set until those keys are individually revoked.
func (s *SigningKeySetsService) Delete(ctx context.Context, kind SigningKeySetKind, name string) error {
	path := "/api/signing-key-sets/" + url.PathEscape(string(kind)) + "/" + url.PathEscape(name)
	return s.client.doRequest(ctx, http.MethodDelete, path, nil, nil, nil)
}

// List returns every signing key set in the workspace.
func (s *SigningKeySetsService) List(ctx context.Context) (*SigningKeySetList, error) {
	var out SigningKeySetList
	if err := s.client.doRequest(ctx, http.MethodGet, "/api/signing-key-sets", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
