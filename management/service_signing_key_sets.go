package management

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

// Get fetches a signing key set by (kind, name). The response carries the set
// plus its individual keys.
func (s *SigningKeySetsService) Get(ctx context.Context, kind SigningKeySetKind, name string) (*SigningKeySetDetail, error) {
	var out SigningKeySetDetail
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

// Search returns signing key sets matching the request. Pass nil for default
// pagination (server-side defaults apply).
func (s *SigningKeySetsService) Search(ctx context.Context, req *SearchRequest) (*SearchResponse[SigningKeySet], error) {
	var out SearchResponse[SigningKeySet]
	if req == nil {
		req = &SearchRequest{}
	}
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/signing-key-sets/search", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
