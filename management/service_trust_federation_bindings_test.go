package management_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/microwave-sh/microwave-go/management"
)

func TestTrustFederationBindings_Create_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	var sawBody management.TrustFederationBindingInput
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(management.TrustFederationBinding{
			ID:            "tfb_abc",
			WorkspaceID:   "ws_42",
			FederationID:  "tf_001",
			FederationKey: sawBody.FederationKey,
			Identity:      sawBody.Identity,
			OutputClaims:  sawBody.OutputClaims,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		})
	}))

	in := &management.TrustFederationBindingInput{
		FederationKey: management.FederationKey("terraform_cloud"),
		Identity: map[string]any{
			"terraform_organization_name": "mataki",
			"terraform_workspace_name":    "sandbar-microwave",
		},
		OutputClaims: map[string]any{"environment": "prod"},
	}
	out, err := client.TrustFederationBindings.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", sawMethod)
	}
	if sawPath != "/api/trust-federation-bindings" {
		t.Errorf("path: got %q, want /api/trust-federation-bindings", sawPath)
	}
	if sawBody.FederationKey != management.FederationKey("terraform_cloud") {
		t.Errorf("body federation_key: got %q", sawBody.FederationKey)
	}
	if sawBody.Identity["terraform_organization_name"] != "mataki" {
		t.Errorf("body identity: got %+v", sawBody.Identity)
	}
	if out.ID != "tfb_abc" || out.WorkspaceID != "ws_42" {
		t.Errorf("response: got %+v", out)
	}
	if out.OutputClaims["environment"] != "prod" {
		t.Errorf("response output_claims: got %+v", out.OutputClaims)
	}
}

func TestTrustFederationBindings_Get_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.TrustFederationBinding{
			ID:            "tfb_abc",
			WorkspaceID:   "ws_42",
			FederationID:  "tf_001",
			FederationKey: management.FederationKey("terraform_cloud"),
			Identity: map[string]any{
				"terraform_organization_name": "mataki",
				"terraform_workspace_name":    "sandbar-microwave",
			},
		})
	}))

	out, err := client.TrustFederationBindings.Get(context.Background(), "tfb_abc")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if sawMethod != http.MethodGet {
		t.Errorf("method: got %q, want GET", sawMethod)
	}
	if sawPath != "/api/trust-federation-bindings/tfb_abc" {
		t.Errorf("path: got %q, want /api/trust-federation-bindings/tfb_abc", sawPath)
	}
	if out.ID != "tfb_abc" {
		t.Errorf("id: got %q", out.ID)
	}
}

func TestTrustFederationBindings_Search_ReturnsForWorkspace(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.SearchResponse[management.TrustFederationBinding]{
			Data: []management.TrustFederationBinding{
				{
					ID:            "tfb_1",
					WorkspaceID:   "ws_42",
					FederationID:  "tf_001",
					FederationKey: management.FederationKey("terraform_cloud"),
					Identity: map[string]any{
						"terraform_organization_name": "mataki",
						"terraform_workspace_name":    "one",
					},
				},
				{
					ID:            "tfb_2",
					WorkspaceID:   "ws_42",
					FederationID:  "tf_002",
					FederationKey: management.FederationKey("github_actions"),
					Identity: map[string]any{
						"repository": "sandbar-cloud/example",
						"workflow":   "deploy.yml",
					},
				},
			},
		})
	}))

	limit := 25
	out, err := client.TrustFederationBindings.Search(context.Background(), &management.SearchRequest{Limit: &limit})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", sawMethod)
	}
	if sawPath != "/api/trust-federation-bindings/search" {
		t.Errorf("path: got %q", sawPath)
	}
	if len(out.Data) != 2 {
		t.Fatalf("data length: got %d, want 2", len(out.Data))
	}
	if out.Data[0].FederationKey != management.FederationKey("terraform_cloud") {
		t.Errorf("data[0] federation_key: got %q", out.Data[0].FederationKey)
	}
	if out.Data[1].FederationKey != management.FederationKey("github_actions") {
		t.Errorf("data[1] federation_key: got %q", out.Data[1].FederationKey)
	}
}

func TestTrustFederationBindings_Delete_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := client.TrustFederationBindings.Delete(context.Background(), "tfb_abc"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if sawMethod != http.MethodDelete {
		t.Errorf("method: got %q, want DELETE", sawMethod)
	}
	if sawPath != "/api/trust-federation-bindings/tfb_abc" {
		t.Errorf("path: got %q, want /api/trust-federation-bindings/tfb_abc", sawPath)
	}
}

func TestTrustFederationBindings_Create_APIError_Surfaces(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"title":"bad request","detail":"identity is required"}`))
	}))

	in := &management.TrustFederationBindingInput{
		FederationKey: management.FederationKey("terraform_cloud"),
	}
	_, err := client.TrustFederationBindings.Create(context.Background(), in)
	if err == nil {
		t.Fatal("expected error from 400")
	}
	var apiErr *management.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *management.Error, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", apiErr.StatusCode)
	}
	if apiErr.Detail == "" {
		t.Errorf("detail should be populated: %+v", apiErr)
	}
}

func TestTrustFederationBindings_Get_NotFound(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"title":"not found","detail":"no such trust federation binding"}`))
	}))
	_, err := client.TrustFederationBindings.Get(context.Background(), "tfb_missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !management.IsNotFound(err) {
		t.Errorf("IsNotFound: got false, want true (err=%v)", err)
	}
}
