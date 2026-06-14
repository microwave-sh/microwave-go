package management

import (
	"context"
	"net/http"
)

// TrustFederationsService is the Management API surface for trust federations.
type TrustFederationsService struct {
	client *Client
}

// Create posts a new trust federation. Key must reference a federation key
// present in the catalog; an unknown key returns 400.
func (s *TrustFederationsService) Create(ctx context.Context, input *TrustFederationInput) (*TrustFederation, error) {
	var out TrustFederation
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/trust-federations", nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a trust federation by ID.
func (s *TrustFederationsService) Get(ctx context.Context, federationID string) (*TrustFederation, error) {
	var out TrustFederation
	if err := s.client.doRequest(ctx, http.MethodGet, "/api/trust-federations/"+federationID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update patches a trust federation's mutable fields. Key is immutable; PATCH
// requests that attempt to change it are rejected by the server.
func (s *TrustFederationsService) Update(ctx context.Context, federationID string, input *TrustFederationUpdateInput) (*TrustFederation, error) {
	var out TrustFederation
	if err := s.client.doRequest(ctx, http.MethodPatch, "/api/trust-federations/"+federationID, nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a trust federation and all of its bindings. In-flight
// redemption requests against this federation complete; subsequent attempts
// return 404.
func (s *TrustFederationsService) Delete(ctx context.Context, federationID string) error {
	return s.client.doRequest(ctx, http.MethodDelete, "/api/trust-federations/"+federationID, nil, nil, nil)
}

// Search returns trust federations matching the request. Pass nil for
// default pagination (server-side defaults apply).
func (s *TrustFederationsService) Search(ctx context.Context, req *SearchRequest) (*SearchResponse[TrustFederation], error) {
	var out SearchResponse[TrustFederation]
	if req == nil {
		req = &SearchRequest{}
	}
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/trust-federations/search", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Redeem exchanges an OIDC JWT for a Microwave token via federation redemption.
// The endpoint is unauthenticated at the Microwave layer — the OIDC token in
// input.Token IS the authentication credential. Callers should construct the
// Client without a management key (or use a zero-value key) when calling
// Redeem, since the Bearer header is not used by this endpoint. If the client
// has a key configured, the server will ignore the Authorization header rather
// than reject it.
func (s *TrustFederationsService) Redeem(ctx context.Context, federationID string, in *RedeemFederationInput) (*RedeemFederationResult, error) {
	var out RedeemFederationResult
	if err := s.client.doRequest(ctx, http.MethodPost, "/api/trust-federations/"+federationID+"/redeem", nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
