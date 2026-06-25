package management_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/microwave-sh/microwave-go/management"
)

// newTestServer constructs an httptest.Server that records requests and lets
// individual tests supply per-path handlers. Returned alongside the client so
// each test gets a fully isolated client+server pair.
func newTestServer(t *testing.T, handler http.Handler) (*httptest.Server, *management.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := management.NewClient(
		management.WithEndpoint(srv.URL),
		management.WithManagementKey("mw_live_test"),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return srv, client
}

func TestNewClientRequiresManagementKey(t *testing.T) {
	t.Setenv("MICROWAVE_MANAGEMENT_KEY", "")
	_, err := management.NewClient(management.WithEndpoint("http://localhost"))
	if err == nil {
		t.Fatal("expected error when management key is absent")
	}
}

func TestNewClientReadsKeyFromEnv(t *testing.T) {
	t.Setenv("MICROWAVE_MANAGEMENT_KEY", "mw_live_from_env")
	c, err := management.NewClient(management.WithEndpoint("http://localhost"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestRequestHeadersIncludeAuthAndVersion(t *testing.T) {
	var sawAuth, sawAPIVersion, sawUserAgent string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAPIVersion = r.Header.Get("API-Version")
		sawUserAgent = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.PermissionSet{ID: "ps_abc", Name: "viewer"})
	}))
	if _, err := client.PermissionSets.Get(context.Background(), "ps_abc"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if sawAuth != "Bearer mw_live_test" {
		t.Errorf("Authorization header: got %q, want Bearer mw_live_test", sawAuth)
	}
	if sawAPIVersion != management.APIVersion {
		t.Errorf("API-Version header: got %q, want %q", sawAPIVersion, management.APIVersion)
	}
	if !strings.HasPrefix(sawUserAgent, "microwave-go-management/") {
		t.Errorf("User-Agent: got %q, want prefix microwave-go-management/", sawUserAgent)
	}
}

func TestWorkspaceIDHeaderOptIn(t *testing.T) {
	var saw string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		saw = r.Header.Get("X-Microwave-Workspace")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.SearchResponse[management.PermissionSet]{})
	}))
	t.Cleanup(srv.Close)
	client, _ := management.NewClient(
		management.WithEndpoint(srv.URL),
		management.WithManagementKey("mw_live_test"),
		management.WithWorkspaceID("ws_42"),
	)
	if _, err := client.PermissionSets.Search(context.Background(), nil); err != nil {
		t.Fatalf("Search: %v", err)
	}
	if saw != "ws_42" {
		t.Errorf("X-Microwave-Workspace: got %q, want ws_42", saw)
	}
}

func TestErrorDecode(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"title":"not found","detail":"no such permission set"}`))
	}))
	_, err := client.PermissionSets.Get(context.Background(), "ps_missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !management.IsNotFound(err) {
		t.Errorf("IsNotFound: got false, want true (err=%v)", err)
	}
}

func TestPermissionSetCreateGetUpdateDelete(t *testing.T) {
	// Records every method+path combination so the test can assert routing.
	var calls []string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/permission-sets":
			_ = json.NewEncoder(w).Encode(management.PermissionSet{ID: "ps_new", Name: "deployer"})
		case "/api/permission-sets/ps_new":
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			_ = json.NewEncoder(w).Encode(management.PermissionSet{ID: "ps_new", Name: "deployer"})
		}
	}))

	ctx := context.Background()
	if _, err := client.PermissionSets.Create(ctx, &management.PermissionSetInput{Name: "deployer"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := client.PermissionSets.Get(ctx, "ps_new"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if _, err := client.PermissionSets.Update(ctx, "ps_new", &management.PermissionSetInput{Name: "deployer"}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := client.PermissionSets.Delete(ctx, "ps_new"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	want := []string{
		"POST /api/permission-sets",
		"GET /api/permission-sets/ps_new",
		"PATCH /api/permission-sets/ps_new",
		"DELETE /api/permission-sets/ps_new",
	}
	if len(calls) != len(want) {
		t.Fatalf("call count: got %d, want %d (%v)", len(calls), len(want), calls)
	}
	for i, c := range calls {
		if c != want[i] {
			t.Errorf("call %d: got %q, want %q", i, c, want[i])
		}
	}
}

func TestSigningKeySetCompositeKeyRouting(t *testing.T) {
	var sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.SigningKeySetDetail{
			Set: management.SigningKeySet{Name: "sandbar-cli-sessions", Kind: management.SigningKeySetKindAsymmetric},
		})
	}))
	if _, err := client.SigningKeySets.Get(context.Background(), management.SigningKeySetKindAsymmetric, "sandbar-cli-sessions"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	want := "/api/signing-key-sets/asymmetric/sandbar-cli-sessions"
	if sawPath != want {
		t.Errorf("path: got %q, want %q", sawPath, want)
	}
}

func TestKeySpecCreateRoundTrip(t *testing.T) {
	var sawBody management.KeySpecInput
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.KeySpec{
			ID:     "spec_abc",
			Name:   sawBody.Name,
			Format: sawBody.Format,
			Opaque: sawBody.Opaque,
		})
	}))
	in := &management.KeySpecInput{
		Name:   "sandbar-management",
		Format: management.KeyFormatOpaque,
		Opaque: management.OpaqueConfig{Prefix: "sbr_live_"},
	}
	out, err := client.KeySpecs.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if out.ID != "spec_abc" || out.Format != management.KeyFormatOpaque || out.Opaque.Prefix != "sbr_live_" {
		t.Errorf("round-trip: got %+v", out)
	}
}

func TestTrustExchangePolicyRoundTrip(t *testing.T) {
	var sawBody management.TrustExchangeInput
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.TrustExchange{
			ID:       "ex_abc",
			Name:     sawBody.Name,
			Provider: sawBody.Provider,
			Policy:   sawBody.Policy,
		})
	}))
	in := &management.TrustExchangeInput{
		Name:     "sandbar-github-actions-exchange",
		Type:     "oidc",
		Provider: management.TrustExchangeProviderCustomOIDC,
		Issuer:   "https://token.actions.githubusercontent.com",
		Policy:   `assertion.repository == "sandbar-cloud/example"`,
	}
	out, err := client.TrustExchanges.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if out.Policy != in.Policy {
		t.Errorf("policy round-trip: got %q, want %q", out.Policy, in.Policy)
	}
}

func TestTrustProviderCreateGetUpdateDelete(t *testing.T) {
	// Records (method, path) pairs so the test can assert routing without
	// duplicating handler logic per verb.
	var calls []string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/trust-providers":
			_ = json.NewEncoder(w).Encode(management.TrustProvider{
				ID:     "tp_new",
				Name:   "ci-deploy-mint",
				Type:   "oidc",
				Active: true,
			})
		case "/api/trust-providers/tp_new":
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			_ = json.NewEncoder(w).Encode(management.TrustProvider{
				ID:     "tp_new",
				Name:   "ci-deploy-mint",
				Type:   "oidc",
				Active: true,
			})
		}
	}))

	ctx := context.Background()
	in := &management.TrustProviderInput{
		Name:            "ci-deploy-mint",
		Type:            "oidc",
		ClientKeySpecID: "spec_client",
		OutputKeySpecID: "spec_output",
		Policy:          `client.workspace_id == output.workspace_id`,
	}
	if _, err := client.TrustProviders.Create(ctx, in); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := client.TrustProviders.Get(ctx, "tp_new"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if _, err := client.TrustProviders.Update(ctx, "tp_new", in); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := client.TrustProviders.Delete(ctx, "tp_new"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	want := []string{
		"POST /api/trust-providers",
		"GET /api/trust-providers/tp_new",
		"PATCH /api/trust-providers/tp_new",
		"DELETE /api/trust-providers/tp_new",
	}
	for i, c := range calls {
		if i >= len(want) || c != want[i] {
			t.Errorf("call %d: got %q, want %q", i, c, want[i])
		}
	}
}
