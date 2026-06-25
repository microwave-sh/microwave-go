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

func TestTrustFederations_Create_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	var sawBody management.TrustFederationInput
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(management.TrustFederation{
			ID:             "tf_abc",
			WorkspaceID:    "ws_42",
			Key:            sawBody.Key,
			Label:          sawBody.Label,
			Description:    sawBody.Description,
			IdentityFields: sawBody.IdentityFields,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		})
	}))

	in := &management.TrustFederationInput{
		Key:            management.FederationKey("terraform_cloud"),
		Label:          "Terraform Cloud",
		Description:    "Bind Terraform Cloud workload identity assertions.",
		IdentityFields: []string{"terraform_organization_name", "terraform_workspace_name"},
	}
	out, err := client.TrustFederations.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", sawMethod)
	}
	if sawPath != "/api/trust-federations" {
		t.Errorf("path: got %q, want /api/trust-federations", sawPath)
	}
	if sawBody.Key != management.FederationKey("terraform_cloud") {
		t.Errorf("body key: got %q", sawBody.Key)
	}
	if out.ID != "tf_abc" || out.WorkspaceID != "ws_42" {
		t.Errorf("response: got %+v", out)
	}
}

func TestTrustFederations_Get_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.TrustFederation{
			ID:          "tf_abc",
			WorkspaceID: "ws_42",
			Key:         management.FederationKey("terraform_cloud"),
			Label:       "Terraform Cloud",
		})
	}))

	out, err := client.TrustFederations.Get(context.Background(), "tf_abc")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if sawMethod != http.MethodGet {
		t.Errorf("method: got %q, want GET", sawMethod)
	}
	if sawPath != "/api/trust-federations/tf_abc" {
		t.Errorf("path: got %q, want /api/trust-federations/tf_abc", sawPath)
	}
	if out.ID != "tf_abc" {
		t.Errorf("id: got %q", out.ID)
	}
}

func TestTrustFederations_Update_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	var sawBody management.TrustFederationUpdateInput
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		w.Header().Set("Content-Type", "application/json")
		lbl := "Updated Label"
		if sawBody.Label != "" {
			lbl = sawBody.Label
		}
		_ = json.NewEncoder(w).Encode(management.TrustFederation{
			ID:          "tf_abc",
			WorkspaceID: "ws_42",
			Key:         management.FederationKey("terraform_cloud"),
			Label:       lbl,
			UpdatedAt:   time.Now().UTC(),
		})
	}))

	out, err := client.TrustFederations.Update(context.Background(), "tf_abc", &management.TrustFederationUpdateInput{
		Label: "Updated Label",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if sawMethod != http.MethodPatch {
		t.Errorf("method: got %q, want PATCH", sawMethod)
	}
	if sawPath != "/api/trust-federations/tf_abc" {
		t.Errorf("path: got %q, want /api/trust-federations/tf_abc", sawPath)
	}
	if out.Label != "Updated Label" {
		t.Errorf("label: got %q", out.Label)
	}
}

func TestTrustFederations_Delete_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := client.TrustFederations.Delete(context.Background(), "tf_abc"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if sawMethod != http.MethodDelete {
		t.Errorf("method: got %q, want DELETE", sawMethod)
	}
	if sawPath != "/api/trust-federations/tf_abc" {
		t.Errorf("path: got %q, want /api/trust-federations/tf_abc", sawPath)
	}
}

func TestTrustFederations_Search_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.SearchResponse[management.TrustFederation]{
			Data: []management.TrustFederation{
				{
					ID:          "tf_1",
					WorkspaceID: "ws_42",
					Key:         management.FederationKey("terraform_cloud"),
					Label:       "Terraform Cloud",
					IdentityFields: []string{
						"terraform_organization_name",
						"terraform_workspace_name",
					},
				},
				{
					ID:          "tf_2",
					WorkspaceID: "ws_42",
					Key:         management.FederationKey("github_actions"),
					Label:       "GitHub Actions",
					IdentityFields: []string{
						"repository",
						"ref",
					},
				},
			},
		})
	}))

	limit := 25
	out, err := client.TrustFederations.Search(context.Background(), &management.SearchRequest{Limit: &limit})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", sawMethod)
	}
	if sawPath != "/api/trust-federations/search" {
		t.Errorf("path: got %q", sawPath)
	}
	if len(out.Data) != 2 {
		t.Fatalf("data length: got %d, want 2", len(out.Data))
	}
	if out.Data[0].Key != management.FederationKey("terraform_cloud") {
		t.Errorf("data[0] key: got %q", out.Data[0].Key)
	}
	if out.Data[1].Key != management.FederationKey("github_actions") {
		t.Errorf("data[1] key: got %q", out.Data[1].Key)
	}
}

func TestTrustFederations_Redeem_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	var sawBody management.RedeemFederationInput
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.RedeemFederationResult{
			Token:     "mw_tok_xyz",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
			Scopes:    []string{"deploy:create"},
			Claims:    map[string]any{"sub": "repo:sandbar-cloud/example"},
		})
	}))

	out, err := client.TrustFederations.Redeem(context.Background(), "tf_abc", &management.RedeemFederationInput{
		Token: "eyJhbGciOiJSUzI1NiJ9.test",
	})
	if err != nil {
		t.Fatalf("Redeem: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", sawMethod)
	}
	if sawPath != "/api/trust-federations/tf_abc/redeem" {
		t.Errorf("path: got %q, want /api/trust-federations/tf_abc/redeem", sawPath)
	}
	if sawBody.Token != "eyJhbGciOiJSUzI1NiJ9.test" {
		t.Errorf("body token: got %q", sawBody.Token)
	}
	if out.Token != "mw_tok_xyz" {
		t.Errorf("response token: got %q", out.Token)
	}
	if len(out.Scopes) != 1 || out.Scopes[0] != "deploy:create" {
		t.Errorf("response scopes: got %v", out.Scopes)
	}
	if out.Claims["sub"] != "repo:sandbar-cloud/example" {
		t.Errorf("response claims: got %v", out.Claims)
	}
}

func TestTrustFederations_Create_APIError_Surfaces(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"title":"bad request","detail":"identity_fields is required"}`))
	}))

	in := &management.TrustFederationInput{
		Key:   management.FederationKey("terraform_cloud"),
		Label: "Terraform Cloud",
	}
	_, err := client.TrustFederations.Create(context.Background(), in)
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

func TestTrustFederations_Get_NotFound(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"title":"not found","detail":"no such trust federation"}`))
	}))
	_, err := client.TrustFederations.Get(context.Background(), "tf_missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !management.IsNotFound(err) {
		t.Errorf("IsNotFound: got false, want true (err=%v)", err)
	}
}
