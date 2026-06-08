package auth_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/microwave-sh/microwave-go/auth"
)

func newTestServer(t *testing.T, handler http.Handler) (*httptest.Server, *auth.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := auth.NewClient(auth.WithEndpoint(srv.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return srv, client
}

func TestRedeemHappyPath(t *testing.T) {
	var sawPath, sawMethod, sawCT string
	var sawBody struct {
		Token string `json:"token"`
	}
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		sawMethod = r.Method
		sawCT = r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &sawBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(auth.ExchangeResult{
			Valid:   true,
			JWT:     "mw_session_jwt_value",
			Subject: "tfc/sandbar/main/plan",
		})
	}))

	result, err := client.TokenExchange.Redeem(context.Background(), "ex_tfc_admin", "ey...tfc...token")
	if err != nil {
		t.Fatalf("Redeem: %v", err)
	}
	if !result.Valid {
		t.Errorf("Valid: got false, want true")
	}
	if result.JWT != "mw_session_jwt_value" {
		t.Errorf("JWT: got %q, want %q", result.JWT, "mw_session_jwt_value")
	}
	if sawPath != "/trust-exchanges/ex_tfc_admin/exchange" {
		t.Errorf("path: got %q", sawPath)
	}
	if sawMethod != http.MethodPost {
		t.Errorf("method: got %q, want POST", sawMethod)
	}
	if sawCT != "application/json" {
		t.Errorf("Content-Type: got %q", sawCT)
	}
	if sawBody.Token != "ey...tfc...token" {
		t.Errorf("body token: got %q", sawBody.Token)
	}
}

func TestRedeemPolicyDenied(t *testing.T) {
	// A denied exchange returns 200 with valid:false in the body — NOT an error.
	// This is the contract the provider needs to distinguish from a transport
	// failure: it can log the Code and RuleResults without falling back to a
	// generic "exchange failed" message.
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(auth.ExchangeResult{
			Valid:       false,
			Code:        "policy_denied",
			RuleResults: map[string]bool{"repository_allowed": false},
		})
	}))

	result, err := client.TokenExchange.Redeem(context.Background(), "ex_tfc_admin", "ey...")
	if err != nil {
		t.Fatalf("Redeem returned error for policy denial: %v", err)
	}
	if result.Valid {
		t.Error("Valid: got true, want false")
	}
	if result.Code != "policy_denied" {
		t.Errorf("Code: got %q, want policy_denied", result.Code)
	}
	if v, ok := result.RuleResults["repository_allowed"]; !ok || v {
		t.Errorf("RuleResults[repository_allowed]: got %v, want false (present)", result.RuleResults)
	}
}

func TestRedeemUnknownExchange(t *testing.T) {
	// Unknown exchange ID is a transport-level 404 — comes back as an error,
	// not a valid:false result.
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"title":"not found","detail":"no such exchange"}`))
	}))

	_, err := client.TokenExchange.Redeem(context.Background(), "ex_missing", "ey...")
	if err == nil {
		t.Fatal("expected error for unknown exchange")
	}
	if !auth.IsNotFound(err) {
		t.Errorf("IsNotFound: got false, want true (err=%v)", err)
	}
}

func TestRedeemEscapesExchangeID(t *testing.T) {
	// httptest decodes URL.Path on receive; URL.RawPath preserves the wire
	// form. Check RawPath so the assertion catches a regression in URL.PathEscape.
	var sawRawPath string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawRawPath = r.URL.RawPath
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(auth.ExchangeResult{Valid: true})
	}))

	_, _ = client.TokenExchange.Redeem(context.Background(), "ex/with/slashes", "tok")
	if !strings.Contains(sawRawPath, "ex%2Fwith%2Fslashes") {
		t.Errorf("exchange id not URL-escaped: raw=%q", sawRawPath)
	}
}

func TestUserAgentSet(t *testing.T) {
	var sawUA string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(auth.ExchangeResult{Valid: true})
	}))

	_, _ = client.TokenExchange.Redeem(context.Background(), "ex_abc", "tok")
	if !strings.HasPrefix(sawUA, "microwave-go-auth/") {
		t.Errorf("User-Agent: got %q, want prefix microwave-go-auth/", sawUA)
	}
}
