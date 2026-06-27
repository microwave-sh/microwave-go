package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// A device-approval login that names a trust exchange must forward it in the
// /auth/device request body so the server mints through that exchange instead
// of the SYSTEM CLI exchange.
func TestLoginDeviceApprovalForwardsTrustExchangeID(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device":
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			_ = json.NewEncoder(w).Encode(map[string]any{"device_code": "dc1", "user_code": "WXYZ-1234", "verification_uri": "http://c/device", "expires_in": 300, "interval": 0})
		case "/auth/device/token":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "approved", "token": "t", "expires_in": 60})
		}
	}))
	defer srv.Close()

	if _, err := loginDeviceApproval(context.Background(), LoginConfig{
		DeviceApprovalURL: srv.URL,
		TrustExchangeID:   "tex_abc",
		OpenBrowser:       func(string) error { return nil },
	}, srv.Client()); err != nil {
		t.Fatalf("loginDeviceApproval: %v", err)
	}
	if gotBody["trust_exchange_id"] != "tex_abc" {
		t.Fatalf("device request body = %v, want trust_exchange_id=tex_abc", gotBody)
	}
}

// An empty TrustExchangeID must NOT send the key at all (server selects its
// SYSTEM CLI exchange), preserving the existing `microwave login` behavior.
func TestLoginDeviceApprovalOmitsEmptyTrustExchangeID(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device":
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			_ = json.NewEncoder(w).Encode(map[string]any{"device_code": "dc1", "user_code": "AAAA-0000", "verification_uri": "http://c/device", "expires_in": 300, "interval": 0})
		case "/auth/device/token":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "approved", "token": "t", "expires_in": 60})
		}
	}))
	defer srv.Close()

	if _, err := loginDeviceApproval(context.Background(), LoginConfig{
		DeviceApprovalURL: srv.URL,
		OpenBrowser:       func(string) error { return nil },
	}, srv.Client()); err != nil {
		t.Fatalf("loginDeviceApproval: %v", err)
	}
	if _, present := gotBody["trust_exchange_id"]; present {
		t.Fatalf("trust_exchange_id must be omitted when empty, got %v", gotBody)
	}
}

// Login with an explicit device-approval mode needs no MetadataURL — the flow
// never reads the authorization-server metadata.
func TestLoginDeviceApprovalNeedsNoMetadata(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device":
			_ = json.NewEncoder(w).Encode(map[string]any{"device_code": "dc1", "user_code": "AAAA-0000", "verification_uri": "http://c/device", "expires_in": 300, "interval": 0})
		case "/auth/device/token":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "approved", "token": "session.jwt", "expires_in": 60})
		default:
			t.Errorf("metadata must not be fetched; unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	creds, err := Login(context.Background(), LoginConfig{
		Mode:              LoginDeviceApproval,
		DeviceApprovalURL: srv.URL,
		TrustExchangeID:   "tex_abc",
		HTTPClient:        srv.Client(),
		OpenBrowser:       func(string) error { return nil },
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if creds.AccessToken != "session.jwt" {
		t.Fatalf("AccessToken = %q", creds.AccessToken)
	}
}
