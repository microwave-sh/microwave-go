package management

import (
	"context"
	"net/http"
)

// KeySpecsService is the Management API surface for key specifications.
type KeySpecsService struct {
	client *Client
}

// Create posts a new key spec. Format (opaque or jwt) is the axis that drives
// which of input.Opaque or input.JWT is read; the unused config block is
// ignored by the server.
func (s *KeySpecsService) Create(ctx context.Context, input *KeySpecInput) (*KeySpec, error) {
	var out KeySpec
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/key-specs", nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a key spec by ID. The returned KeySpec includes the resolved
// PermissionSet (not just its ID) when one is bound.
func (s *KeySpecsService) Get(ctx context.Context, id string) (*KeySpec, error) {
	var out KeySpec
	if err := s.client.doRequest(ctx, http.MethodGet, "/api/key-specs/"+id, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update replaces a key spec's mutable fields. Format is immutable; PATCH
// requests that change it are rejected by the server.
func (s *KeySpecsService) Update(ctx context.Context, id string, input *KeySpecInput) (*KeySpec, error) {
	var out KeySpec
	if err := s.client.doRequest(ctx, http.MethodPatch, "/api/key-specs/"+id, nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a key spec. Issued keys against this spec are revoked
// transitively.
func (s *KeySpecsService) Delete(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, http.MethodDelete, "/api/key-specs/"+id, nil, nil, nil)
}

// Search returns key specs matching the request. Pass nil for default
// pagination (server-side defaults apply).
func (s *KeySpecsService) Search(ctx context.Context, req *SearchRequest) (*SearchResponse[KeySpec], error) {
	var out SearchResponse[KeySpec]
	if req == nil {
		req = &SearchRequest{}
	}
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/key-specs/search", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
