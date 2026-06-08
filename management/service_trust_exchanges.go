package management

import (
	"context"
	"net/http"
)

// TrustExchangesService is the Management API surface for trust exchanges.
type TrustExchangesService struct {
	client *Client
}

// Create posts a new trust exchange. The CEL Policy is validated server-side
// against the assertion shape implied by Provider; a malformed policy returns
// 400 with a compilation error in the response body.
func (s *TrustExchangesService) Create(ctx context.Context, input *TrustExchangeInput) (*TrustExchange, error) {
	var out TrustExchange
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/trust-exchanges", nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a trust exchange by ID.
func (s *TrustExchangesService) Get(ctx context.Context, id string) (*TrustExchange, error) {
	var out TrustExchange
	if err := s.client.doRequest(ctx, http.MethodGet, "/api/trust-exchanges/"+id, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update replaces a trust exchange's mutable fields. Type and Provider are
// immutable; PATCH requests that change them are rejected.
func (s *TrustExchangesService) Update(ctx context.Context, id string, input *TrustExchangeInput) (*TrustExchange, error) {
	var out TrustExchange
	if err := s.client.doRequest(ctx, http.MethodPatch, "/api/trust-exchanges/"+id, nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a trust exchange. In-flight token mints against this exchange
// complete; subsequent attempts return 404.
func (s *TrustExchangesService) Delete(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, http.MethodDelete, "/api/trust-exchanges/"+id, nil, nil, nil)
}

// List returns every trust exchange in the workspace.
func (s *TrustExchangesService) List(ctx context.Context) (*TrustExchangeList, error) {
	var out TrustExchangeList
	if err := s.client.doRequest(ctx, http.MethodGet, "/api/trust-exchanges", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
