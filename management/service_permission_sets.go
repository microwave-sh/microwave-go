package management

import (
	"context"
	"net/http"
)

// PermissionSetsService is the Management API surface for permission sets.
type PermissionSetsService struct {
	client *Client
}

// Create posts a new permission set. Name must be unique within the workspace.
func (s *PermissionSetsService) Create(ctx context.Context, input *PermissionSetInput) (*PermissionSet, error) {
	var out PermissionSet
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/permission-sets", nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a permission set by ID.
func (s *PermissionSetsService) Get(ctx context.Context, id string) (*PermissionSet, error) {
	var out PermissionSet
	if err := s.client.doRequest(ctx, http.MethodGet, "/api/permission-sets/"+id, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update replaces a permission set's fields and permission list. Partial
// updates are not supported by the API; pass the full desired state.
func (s *PermissionSetsService) Update(ctx context.Context, id string, input *PermissionSetInput) (*PermissionSet, error) {
	var out PermissionSet
	if err := s.client.doRequest(ctx, http.MethodPatch, "/api/permission-sets/"+id, nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a permission set. Returns nil on success or *Error on failure
// (use IsNotFound to detect an already-deleted case for idempotency).
func (s *PermissionSetsService) Delete(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, http.MethodDelete, "/api/permission-sets/"+id, nil, nil, nil)
}

// Search returns permission sets matching the request. Pass nil for default
// pagination (server-side defaults apply).
func (s *PermissionSetsService) Search(ctx context.Context, req *SearchRequest) (*SearchResponse[PermissionSet], error) {
	var out SearchResponse[PermissionSet]
	if req == nil {
		req = &SearchRequest{}
	}
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/permission-sets/search", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
