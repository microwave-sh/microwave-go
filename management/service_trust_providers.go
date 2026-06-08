package management

import (
	"context"
	"net/http"
)

// TrustProvidersService is the Management API surface for trust providers.
type TrustProvidersService struct {
	client *Client
}

// Create posts a new trust provider. The CEL Policy is validated server-side
// against the assertion shape; a malformed policy returns 400 with the
// compilation error in the response body.
func (s *TrustProvidersService) Create(ctx context.Context, input *TrustProviderInput) (*TrustProvider, error) {
	var out TrustProvider
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/trust-providers", nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a trust provider by ID.
func (s *TrustProvidersService) Get(ctx context.Context, id string) (*TrustProvider, error) {
	var out TrustProvider
	if err := s.client.doRequest(ctx, http.MethodGet, "/api/trust-providers/"+id, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update replaces a trust provider's mutable fields. Type and client/output
// key spec bindings are immutable post-create; PATCH requests that change
// them are rejected.
func (s *TrustProvidersService) Update(ctx context.Context, id string, input *TrustProviderInput) (*TrustProvider, error) {
	var out TrustProvider
	if err := s.client.doRequest(ctx, http.MethodPatch, "/api/trust-providers/"+id, nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a trust provider. In-flight mints against this provider
// complete; subsequent mint attempts return 404.
func (s *TrustProvidersService) Delete(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, http.MethodDelete, "/api/trust-providers/"+id, nil, nil, nil)
}

// List returns every trust provider in the workspace.
func (s *TrustProvidersService) List(ctx context.Context) (*TrustProviderList, error) {
	var out TrustProviderList
	if err := s.client.doRequest(ctx, http.MethodGet, "/api/trust-providers", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
