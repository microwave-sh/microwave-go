package management_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/microwave-sh/microwave-go/management"
)

func TestKeySpecs_IssueKey_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	var sawBody management.IssueKeyInput
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"key_1","key":"ap_live_abc","subject":"ws_42","spec_id":"spec_9","scopes":["a"],"created_at":"2026-06-29T00:00:00Z"}`))
	}))

	out, err := client.KeySpecs.IssueKey(context.Background(), "spec_9", &management.IssueKeyInput{
		Subject: "ws_42",
		Name:    "prod key",
	})
	if err != nil {
		t.Fatalf("IssueKey: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", sawMethod)
	}
	if sawPath != "/api/key-specs/spec_9/keys" {
		t.Errorf("path: got %q", sawPath)
	}
	if sawBody.Subject != "ws_42" || sawBody.Name != "prod key" {
		t.Errorf("body: got %+v", sawBody)
	}
	if out.ID != "key_1" || out.Key != "ap_live_abc" || out.Subject != "ws_42" {
		t.Errorf("response: got %+v", out)
	}
}

func TestKeySpecs_SearchIssuedKeys_FiltersBySubject(t *testing.T) {
	var sawPath string
	var sawBody management.SearchRequest
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"key_1","spec_id":"spec_9","subject":"ws_42","name":"prod","status":"active","created_at":"2026-06-29T00:00:00Z"}],"has_more":false,"limit":25}`))
	}))

	out, err := client.KeySpecs.SearchIssuedKeys(context.Background(), "spec_9", &management.SearchRequest{
		Filter: map[string]map[string]any{"subject": {"eq": "ws_42"}},
	})
	if err != nil {
		t.Fatalf("SearchIssuedKeys: %v", err)
	}
	if sawPath != "/api/key-specs/spec_9/keys/search" {
		t.Errorf("path: got %q", sawPath)
	}
	if sawBody.Filter["subject"]["eq"] != "ws_42" {
		t.Errorf("filter: got %+v", sawBody.Filter)
	}
	if len(out.Data) != 1 || out.Data[0].Subject != "ws_42" || out.Data[0].Name != "prod" {
		t.Errorf("data: got %+v", out.Data)
	}
}

func TestKeySpecs_RevokeKey_RoundTrip(t *testing.T) {
	var sawMethod, sawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMethod = r.Method
		sawPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := client.KeySpecs.RevokeKey(context.Background(), "spec_9", "key_1"); err != nil {
		t.Fatalf("RevokeKey: %v", err)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", sawMethod)
	}
	if sawPath != "/api/key-specs/spec_9/keys/key_1/revoke" {
		t.Errorf("path: got %q", sawPath)
	}
}
