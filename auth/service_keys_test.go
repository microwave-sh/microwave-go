package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestKeys_Verify_Valid(t *testing.T) {
	var sawPath string
	var sawBody struct {
		Key string `json:"key"`
	}
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&sawBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"valid":true,"key_id":"key_1","subject":"ws_42","scopes":["api:read"]}`))
	}))

	out, err := client.Keys.Verify(context.Background(), "ap_live_abc")
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if sawPath != "/verify" {
		t.Errorf("path: got %q, want /verify", sawPath)
	}
	if sawBody.Key != "ap_live_abc" {
		t.Errorf("body key: got %q", sawBody.Key)
	}
	if !out.Valid || out.Subject != "ws_42" || len(out.Scopes) != 1 {
		t.Errorf("result: got %+v", out)
	}
}

func TestKeys_Verify_InvalidIsNotAnError(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"valid":false,"code":"invalid_key"}`))
	}))
	out, err := client.Keys.Verify(context.Background(), "ap_live_bad")
	if err != nil {
		t.Fatalf("Verify returned transport error for a denial: %v", err)
	}
	if out.Valid || out.Code != "invalid_key" {
		t.Errorf("result: got %+v", out)
	}
}
