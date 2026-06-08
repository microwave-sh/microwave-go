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

// List returns every permission set in the workspace. The Management API
// returns the full set today; pagination metadata will land in a future SDK
// version when the server-side paging contract stabilises.
func (s *PermissionSetsService) List(ctx context.Context) (*PermissionSetList, error) {
	var out PermissionSetList
	if err := s.client.doRequest(ctx, http.MethodGet, "/api/permission-sets", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
