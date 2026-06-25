package auth_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/microwave-sh/microwave-go/auth"
)

func TestRedeemTokenExchange_Success(t *testing.T) {
	var gotForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotForm = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mw_jwt","token_type":"Bearer","expires_in":3600,"scope":"deploy"}`))
	}))
	defer srv.Close()

	out, err := auth.RedeemTokenExchange(context.Background(), srv.Client(), srv.URL, "https://api.example/fed_ci", "gh_oidc_token")
	if err != nil {
		t.Fatalf("RedeemTokenExchange: %v", err)
	}
	if out.AccessToken != "mw_jwt" || out.ExpiresIn != 3600 || out.Scope != "deploy" {
		t.Fatalf("unexpected result: %+v", out)
	}
	// The RFC 8693 grant + resource indicator + subject token must be sent.
	if gotForm.Get("grant_type") != "urn:ietf:params:oauth:grant-type:token-exchange" {
		t.Errorf("grant_type: got %q", gotForm.Get("grant_type"))
	}
	if gotForm.Get("resource") != "https://api.example/fed_ci" {
		t.Errorf("resource: got %q", gotForm.Get("resource"))
	}
	if gotForm.Get("subject_token") != "gh_oidc_token" {
		t.Errorf("subject_token: got %q", gotForm.Get("subject_token"))
	}
}

// TestRedeemTokenExchange_SurfacesOAuthError is the regression guard: a server
// RFC 6749 §5.2 error must reach the caller as a typed *OAuthError carrying the
// code + description, NOT collapse to a bare "HTTP 400".
func TestRedeemTokenExchange_SurfacesOAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"policy denied: assertion.repository did not match"}`))
	}))
	defer srv.Close()

	_, err := auth.RedeemTokenExchange(context.Background(), srv.Client(), srv.URL, "https://api.example/fed_ci", "gh_oidc_token")
	if err == nil {
		t.Fatal("expected an error")
	}
	var oe *auth.OAuthError
	if !errors.As(err, &oe) {
		t.Fatalf("want *OAuthError, got %T: %v", err, err)
	}
	if oe.Code != "invalid_grant" {
		t.Errorf("Code: got %q, want invalid_grant", oe.Code)
	}
	if oe.Status != http.StatusBadRequest {
		t.Errorf("Status: got %d, want 400", oe.Status)
	}
	if oe.Description == "" {
		t.Error("Description should carry the server's error_description")
	}
}

func TestRedeemTokenExchange_RequiresInputs(t *testing.T) {
	if _, err := auth.RedeemTokenExchange(context.Background(), nil, "", "res", "tok"); err == nil {
		t.Error("expected error for empty token endpoint")
	}
	if _, err := auth.RedeemTokenExchange(context.Background(), nil, "https://e/token", "", "tok"); err == nil {
		t.Error("expected error for empty resource")
	}
	if _, err := auth.RedeemTokenExchange(context.Background(), nil, "https://e/token", "res", ""); err == nil {
		t.Error("expected error for empty subject token")
	}
}
