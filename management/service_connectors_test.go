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

func TestTrustBindings_Create_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	var sawBody management.TrustBindingInput
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(management.TrustBinding{
			ID:           "tb_abc",
			WorkspaceID:  "ws_42",
			BindingType:  sawBody.BindingType,
			Identity:     sawBody.Identity,
			OutputClaims: sawBody.OutputClaims,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		})
	}))

	in := &management.TrustBindingInput{
		BindingType: management.TrustBindingType("terraform_cloud"),
		Identity: map[string]any{
			"terraform_organization_name": "mataki",
			"terraform_workspace_name":    "sandbar-microwave",
		},
		OutputClaims: map[string]any{"environment": "prod"},
	}
	out, err := client.TrustBindings.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", sawMethod)
	}
	if sawPath != "/api/trust-bindings" {
		t.Errorf("path: got %q, want /api/trust-bindings", sawPath)
	}
	if sawBody.BindingType != management.TrustBindingType("terraform_cloud") {
		t.Errorf("body binding_type: got %q", sawBody.BindingType)
	}
	if sawBody.Identity["terraform_organization_name"] != "mataki" {
		t.Errorf("body identity: got %+v", sawBody.Identity)
	}
	if out.ID != "tb_abc" || out.WorkspaceID != "ws_42" {
		t.Errorf("response: got %+v", out)
	}
	if out.OutputClaims["environment"] != "prod" {
		t.Errorf("response output_claims: got %+v", out.OutputClaims)
	}
}

func TestTrustBindings_Get_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.TrustBinding{
			ID:          "tb_abc",
			WorkspaceID: "ws_42",
			BindingType: management.TrustBindingType("terraform_cloud"),
			Identity: map[string]any{
				"terraform_organization_name": "mataki",
				"terraform_workspace_name":    "sandbar-microwave",
			},
		})
	}))

	out, err := client.TrustBindings.Get(context.Background(), "tb_abc")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if sawMethod != http.MethodGet {
		t.Errorf("method: got %q, want GET", sawMethod)
	}
	if sawPath != "/api/trust-bindings/tb_abc" {
		t.Errorf("path: got %q, want /api/trust-bindings/tb_abc", sawPath)
	}
	if out.ID != "tb_abc" {
		t.Errorf("id: got %q", out.ID)
	}
}

func TestTrustBindings_Search_ReturnsForWorkspace(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.SearchResponse[management.TrustBinding]{
			Data: []management.TrustBinding{
				{
					ID:          "tb_1",
					WorkspaceID: "ws_42",
					BindingType: management.TrustBindingType("terraform_cloud"),
					Identity: map[string]any{
						"terraform_organization_name": "mataki",
						"terraform_workspace_name":    "one",
					},
				},
				{
					ID:          "tb_2",
					WorkspaceID: "ws_42",
					BindingType: management.TrustBindingType("github_actions"),
					Identity: map[string]any{
						"repository": "sandbar-cloud/example",
						"workflow":   "deploy.yml",
					},
				},
			},
		})
	}))

	limit := 25
	out, err := client.TrustBindings.Search(context.Background(), &management.SearchRequest{Limit: &limit})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", sawMethod)
	}
	if sawPath != "/api/trust-bindings/search" {
		t.Errorf("path: got %q", sawPath)
	}
	if len(out.Data) != 2 {
		t.Fatalf("data length: got %d, want 2", len(out.Data))
	}
	if out.Data[0].BindingType != management.TrustBindingType("terraform_cloud") {
		t.Errorf("data[0] binding_type: got %q", out.Data[0].BindingType)
	}
	if out.Data[1].BindingType != management.TrustBindingType("github_actions") {
		t.Errorf("data[1] binding_type: got %q", out.Data[1].BindingType)
	}
}

func TestTrustBindingTypes_Search(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.SearchResponse[management.TrustBindingTypeDefinition]{
			Data: []management.TrustBindingTypeDefinition{
				{
					Key:                    management.TrustBindingType("terraform_cloud"),
					DisplayName:            "Terraform Cloud",
					Description:            "Bind Terraform Cloud workload identity assertions.",
					LogoURL:                "https://assets.microwave.sh/trust-binding-types/terraform-cloud.svg",
					DocsURL:                "https://microwave.sh/docs/trust-bindings/terraform-cloud",
					RequiredIdentityClaims: []string{"terraform_organization_name", "terraform_workspace_name"},
				},
			},
		})
	}))

	out, err := client.TrustBindingTypes.Search(context.Background(), nil)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", sawMethod)
	}
	if sawPath != "/api/trust-binding-types/search" {
		t.Errorf("path: got %q", sawPath)
	}
	if len(out.Data) != 1 {
		t.Fatalf("data length: got %d, want 1", len(out.Data))
	}
	if out.Data[0].LogoURL == "" {
		t.Fatalf("logo_url should be populated: %+v", out.Data[0])
	}
}

func TestTrustBindings_Delete_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := client.TrustBindings.Delete(context.Background(), "tb_abc"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if sawMethod != http.MethodDelete {
		t.Errorf("method: got %q, want DELETE", sawMethod)
	}
	if sawPath != "/api/trust-bindings/tb_abc" {
		t.Errorf("path: got %q, want /api/trust-bindings/tb_abc", sawPath)
	}
}

func TestTrustBindings_Create_APIError_Surfaces(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"title":"bad request","detail":"identity is required"}`))
	}))

	in := &management.TrustBindingInput{
		BindingType: management.TrustBindingType("terraform_cloud"),
	}
	_, err := client.TrustBindings.Create(context.Background(), in)
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

func TestTrustBindings_Get_NotFound(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"title":"not found","detail":"no such trust binding"}`))
	}))
	_, err := client.TrustBindings.Get(context.Background(), "tb_missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !management.IsNotFound(err) {
		t.Errorf("IsNotFound: got false, want true (err=%v)", err)
	}
}
