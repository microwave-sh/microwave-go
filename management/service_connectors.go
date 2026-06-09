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

func (s *TrustBindingsService) List(ctx context.Context, workspaceID string) (*TrustBindingList, error) {
	var out TrustBindingList
	if err := s.client.doRequest(ctx, http.MethodGet, "/workspaces/"+workspaceID+"/trust-bindings", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type TrustBindingTypesService struct {
	client *Client
}

func (s *TrustBindingTypesService) List(ctx context.Context) (*TrustBindingTypeList, error) {
	var out TrustBindingTypeList
	if err := s.client.doRequest(ctx, http.MethodGet, "/trust-binding-types", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
