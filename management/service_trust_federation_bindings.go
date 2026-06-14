package management

import (
	"context"
	"net/http"
)

// TrustFederationBindingsService is the Management API surface for trust
// federation bindings.
type TrustFederationBindingsService struct {
	client *Client
}

// Create posts a new trust federation binding. FederationKey must reference an
// existing federation in the workspace; duplicate identity tuples within the
// same federation return 409.
func (s *TrustFederationBindingsService) Create(ctx context.Context, input *TrustFederationBindingInput) (*TrustFederationBinding, error) {
	var out TrustFederationBinding
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/trust-federation-bindings", nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a trust federation binding by ID.
func (s *TrustFederationBindingsService) Get(ctx context.Context, bindingID string) (*TrustFederationBinding, error) {
	var out TrustFederationBinding
	if err := s.client.doRequest(ctx, http.MethodGet, "/api/trust-federation-bindings/"+bindingID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a trust federation binding. In-flight redemption requests
// that already matched this binding complete; subsequent attempts will fail to
// match and return an authorization error.
func (s *TrustFederationBindingsService) Delete(ctx context.Context, bindingID string) error {
	return s.client.doRequest(ctx, http.MethodDelete, "/api/trust-federation-bindings/"+bindingID, nil, nil, nil)
}

// Search returns trust federation bindings matching the request. Pass nil for
// default pagination (server-side defaults apply).
func (s *TrustFederationBindingsService) Search(ctx context.Context, req *SearchRequest) (*SearchResponse[TrustFederationBinding], error) {
	var out SearchResponse[TrustFederationBinding]
	if req == nil {
		req = &SearchRequest{}
	}
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/trust-federation-bindings/search", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
