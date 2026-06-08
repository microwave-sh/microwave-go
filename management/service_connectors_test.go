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

func TestConnectors_Create_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	var sawBody management.ConnectorInput
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(management.Connector{
			ID:             "wfb_abc",
			WorkspaceID:    "ws_42",
			Provider:       sawBody.Provider,
			TerraformCloud: sawBody.TerraformCloud,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		})
	}))

	in := &management.ConnectorInput{
		Provider: management.ConnectorProviderTerraformCloud,
		TerraformCloud: &management.TerraformCloudClaims{
			Organization: "mataki",
			Workspace:    "sandbar-microwave",
		},
	}
	out, err := client.Connectors.Create(context.Background(), "ws_42", in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", sawMethod)
	}
	if sawPath != "/workspaces/ws_42/connectors" {
		t.Errorf("path: got %q, want /workspaces/ws_42/connectors", sawPath)
	}
	if sawBody.Provider != management.ConnectorProviderTerraformCloud {
		t.Errorf("body provider: got %q", sawBody.Provider)
	}
	if sawBody.TerraformCloud == nil || sawBody.TerraformCloud.Organization != "mataki" {
		t.Errorf("body terraform_cloud: got %+v", sawBody.TerraformCloud)
	}
	if out.ID != "wfb_abc" || out.WorkspaceID != "ws_42" {
		t.Errorf("response: got %+v", out)
	}
	if out.TerraformCloud == nil || out.TerraformCloud.Workspace != "sandbar-microwave" {
		t.Errorf("response terraform_cloud: got %+v", out.TerraformCloud)
	}
}

func TestConnectors_Create_GitHubActions(t *testing.T) {
	var sawBody management.ConnectorInput
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(management.Connector{
			ID:            "wfb_gh1",
			WorkspaceID:   "ws_42",
			Provider:      sawBody.Provider,
			GitHubActions: sawBody.GitHubActions,
		})
	}))

	in := &management.ConnectorInput{
		Provider: management.ConnectorProviderGitHubActions,
		GitHubActions: &management.GitHubActionsClaims{
			Repository: "sandbar-cloud/example",
			Workflow:   "deploy.yml",
		},
	}
	out, err := client.Connectors.Create(context.Background(), "ws_42", in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if out.Provider != management.ConnectorProviderGitHubActions {
		t.Errorf("provider: got %q", out.Provider)
	}
	if out.GitHubActions == nil || out.GitHubActions.Repository != "sandbar-cloud/example" {
		t.Errorf("github_actions: got %+v", out.GitHubActions)
	}
	if out.TerraformCloud != nil {
		t.Errorf("terraform_cloud should be nil for github_actions provider, got %+v", out.TerraformCloud)
	}
}

func TestConnectors_Get_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.Connector{
			ID:          "wfb_abc",
			WorkspaceID: "ws_42",
			Provider:    management.ConnectorProviderTerraformCloud,
			TerraformCloud: &management.TerraformCloudClaims{
				Organization: "mataki",
				Workspace:    "sandbar-microwave",
			},
		})
	}))

	out, err := client.Connectors.Get(context.Background(), "ws_42", "wfb_abc")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if sawMethod != http.MethodGet {
		t.Errorf("method: got %q, want GET", sawMethod)
	}
	if sawPath != "/workspaces/ws_42/connectors/wfb_abc" {
		t.Errorf("path: got %q, want /workspaces/ws_42/connectors/wfb_abc", sawPath)
	}
	if out.ID != "wfb_abc" {
		t.Errorf("id: got %q", out.ID)
	}
}

func TestConnectors_List_ReturnsForWorkspace(t *testing.T) {
	var sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.ConnectorList{
			Data: []management.Connector{
				{
					ID:          "wfb_1",
					WorkspaceID: "ws_42",
					Provider:    management.ConnectorProviderTerraformCloud,
					TerraformCloud: &management.TerraformCloudClaims{
						Organization: "mataki",
						Workspace:    "one",
					},
				},
				{
					ID:          "wfb_2",
					WorkspaceID: "ws_42",
					Provider:    management.ConnectorProviderGitHubActions,
					GitHubActions: &management.GitHubActionsClaims{
						Repository: "sandbar-cloud/example",
						Workflow:   "deploy.yml",
					},
				},
			},
		})
	}))

	out, err := client.Connectors.List(context.Background(), "ws_42")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if sawPath != "/workspaces/ws_42/connectors" {
		t.Errorf("path: got %q", sawPath)
	}
	if len(out.Data) != 2 {
		t.Fatalf("data length: got %d, want 2", len(out.Data))
	}
	if out.Data[0].Provider != management.ConnectorProviderTerraformCloud {
		t.Errorf("data[0] provider: got %q", out.Data[0].Provider)
	}
	if out.Data[1].Provider != management.ConnectorProviderGitHubActions {
		t.Errorf("data[1] provider: got %q", out.Data[1].Provider)
	}
}

func TestConnectors_Delete_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := client.Connectors.Delete(context.Background(), "ws_42", "wfb_abc"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if sawMethod != http.MethodDelete {
		t.Errorf("method: got %q, want DELETE", sawMethod)
	}
	if sawPath != "/workspaces/ws_42/connectors/wfb_abc" {
		t.Errorf("path: got %q, want /workspaces/ws_42/connectors/wfb_abc", sawPath)
	}
}

func TestConnectors_Create_APIError_Surfaces(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"title":"bad request","detail":"terraform_cloud is required when provider is \"terraform_cloud\""}`))
	}))

	in := &management.ConnectorInput{
		Provider: management.ConnectorProviderTerraformCloud,
		// TerraformCloud intentionally missing to trigger server-side 400.
	}
	_, err := client.Connectors.Create(context.Background(), "ws_42", in)
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

func TestConnectors_Get_NotFound(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"title":"not found","detail":"no such connector"}`))
	}))
	_, err := client.Connectors.Get(context.Background(), "ws_42", "wfb_missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !management.IsNotFound(err) {
		t.Errorf("IsNotFound: got false, want true (err=%v)", err)
	}
}
