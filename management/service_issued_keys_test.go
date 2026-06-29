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
