package management

import (
	"context"
	"net/http"
)

// ConnectorsService is the Management API surface for workspace federation
// connectors — the customer-facing rows the SYSTEM federation Trust Exchanges
// resolve at policy-evaluation time.
type ConnectorsService struct {
	client *Client
}

// Create posts a new federation connector under the given workspace. The
// server validates that exactly one provider-shaped sub-object is populated
// and that it matches Provider; mismatches return 400.
func (s *ConnectorsService) Create(ctx context.Context, workspaceID string, input *ConnectorInput) (*Connector, error) {
	var out Connector
	if err := s.client.doRequest(ctx, http.MethodPost, "/workspaces/"+workspaceID+"/connectors", nil, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a federation connector by ID under the given workspace. The
// path is workspace-scoped so a stolen connector ID alone can't address a
// connector under a different workspace.
func (s *ConnectorsService) Get(ctx context.Context, workspaceID, connectorID string) (*Connector, error) {
	var out Connector
	if err := s.client.doRequest(ctx, http.MethodGet, "/workspaces/"+workspaceID+"/connectors/"+connectorID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a federation connector. Subsequent policy evaluations that
// reference this binding fall through to no-match and the underlying token
// exchange fails closed.
func (s *ConnectorsService) Delete(ctx context.Context, workspaceID, connectorID string) error {
	return s.client.doRequest(ctx, http.MethodDelete, "/workspaces/"+workspaceID+"/connectors/"+connectorID, nil, nil, nil)
}

// List returns every federation connector in the given workspace.
func (s *ConnectorsService) List(ctx context.Context, workspaceID string) (*ConnectorList, error) {
	var out ConnectorList
	if err := s.client.doRequest(ctx, http.MethodGet, "/workspaces/"+workspaceID+"/connectors", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
