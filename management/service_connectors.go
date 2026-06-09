package management

import (
	"context"
	"net/http"
)

// TrustBindingsService is the Management API surface for Trust Bindings.
type TrustBindingsService struct {
	client *Client
}

func (s *TrustBindingsService) Create(ctx context.Context, workspaceID string, input *TrustBindingInput) (*TrustBinding, error) {
	var out TrustBinding
	if err := s.client.doRequest(ctx, http.MethodPost, "/workspaces/"+workspaceID+"/trust-bindings", nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *TrustBindingsService) Get(ctx context.Context, workspaceID, bindingID string) (*TrustBinding, error) {
	var out TrustBinding
	if err := s.client.doRequest(ctx, http.MethodGet, "/workspaces/"+workspaceID+"/trust-bindings/"+bindingID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *TrustBindingsService) Delete(ctx context.Context, workspaceID, bindingID string) error {
	return s.client.doRequest(ctx, http.MethodDelete, "/workspaces/"+workspaceID+"/trust-bindings/"+bindingID, nil, nil, nil)
}

func (s *TrustBindingsService) Search(ctx context.Context, workspaceID string, req *SearchRequest) (*SearchResponse[TrustBinding], error) {
	var out SearchResponse[TrustBinding]
	if req == nil {
		req = &SearchRequest{}
	}
	if err := s.client.doRequest(ctx, http.MethodPost, "/workspaces/"+workspaceID+"/trust-bindings/search", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type TrustBindingTypesService struct {
	client *Client
}

func (s *TrustBindingTypesService) Search(ctx context.Context, req *SearchRequest) (*SearchResponse[TrustBindingTypeDefinition], error) {
	var out SearchResponse[TrustBindingTypeDefinition]
	if req == nil {
		req = &SearchRequest{}
	}
	if err := s.client.doRequest(ctx, http.MethodPost, "/trust-binding-types/search", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
