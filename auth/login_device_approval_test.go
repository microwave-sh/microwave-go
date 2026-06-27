package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoginDeviceApprovalRequestsPrintsAndPolls(t *testing.T) {
	var openedURL string
	polls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device":
			_ = json.NewEncoder(w).Encode(map[string]any{"device_code": "dc1", "user_code": "WXYZ-1234", "verification_uri": "http://console/device", "expires_in": 300, "interval": 0})
		case "/auth/device/token":
			polls++
			if polls < 2 {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "authorization_pending"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "session.jwt", "token_type": "Bearer", "expires_in": 3600})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out bytes.Buffer
	creds, err := loginDeviceApproval(context.Background(), LoginConfig{
		DeviceApprovalURL: srv.URL,
		Output:            &out,
		OpenBrowser:       func(u string) error { openedURL = u; return nil },
	}, srv.Client())
	if err != nil {
		t.Fatalf("loginDeviceApproval: %v", err)
	}
	if creds.AccessToken != "session.jwt" {
		t.Fatalf("AccessToken = %q", creds.AccessToken)
	}
	if polls < 2 {
		t.Fatalf("expected to poll until approved, polls=%d", polls)
	}
	// Security invariant: only the static verification_uri is opened/printed.
	// The device_code secret must never reach the browser or the operator.
	if openedURL != "http://console/device" {
		t.Fatalf("opened URL = %q, want the static verification_uri", openedURL)
	}
	if strings.Contains(openedURL, "dc1") || strings.Contains(out.String(), "dc1") {
		t.Fatalf("device_code must not be shown to the operator; opened=%q out=%q", openedURL, out.String())
	}
	if !strings.Contains(out.String(), "WXYZ-1234") {
		t.Fatalf("expected the user code shown in output, got %q", out.String())
	}
}

func TestLoginDeviceApprovalDeniedIsTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device":
			_ = json.NewEncoder(w).Encode(map[string]any{"device_code": "dc1", "user_code": "UCOD-9999", "verification_uri": "http://c/device", "expires_in": 300, "interval": 0})
		case "/auth/device/token":
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "access_denied"})
		}
	}))
	defer srv.Close()

	_, err := loginDeviceApproval(context.Background(), LoginConfig{
		DeviceApprovalURL: srv.URL, Output: io.Discard, OpenBrowser: func(string) error { return nil },
	}, srv.Client())
	if err == nil {
		t.Fatal("expected an error on denial")
	}
	var oe *OAuthError
	if !asOAuthError(err, &oe) || oe.Code != "access_denied" {
		t.Fatalf("want typed OAuthError access_denied, got %v", err)
	}
}

// TestLoginAutoRespectsAdvertisedDeviceApprovalFlow proves that LoginAuto, when
// the AS metadata advertises cli_login_flow=device_approval, drives the
// device-approval flow — with no client-id.
func TestLoginAutoRespectsAdvertisedDeviceApprovalFlow(t *testing.T) {
	var hitDevice bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/oauth-authorization-server":
			// Real microwave metadata always advertises a token_endpoint (for the
			// other grants); the device-approval flow simply doesn't use it.
			_ = json.NewEncoder(w).Encode(map[string]any{"issuer": "x", "token_endpoint": "http://x/token", "cli_login_flow": "device_approval"})
		case "/auth/device":
			hitDevice = true
			_ = json.NewEncoder(w).Encode(map[string]any{"device_code": "dc", "user_code": "UCOD-1234", "verification_uri": "http://example/device", "expires_in": 300, "interval": 0})
		case "/auth/device/token":
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "token_type": "Bearer", "expires_in": 3600})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	creds, err := Login(context.Background(), LoginConfig{
		MetadataURL:       srv.URL + "/.well-known/oauth-authorization-server",
		DeviceApprovalURL: srv.URL,
		Mode:              LoginAuto,
		Output:            io.Discard,
		OpenBrowser:       func(string) error { return nil },
		HTTPClient:        srv.Client(),
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if !hitDevice {
		t.Fatal("expected the device-approval endpoint to be used")
	}
	if creds.AccessToken != "tok" {
		t.Fatalf("AccessToken = %q", creds.AccessToken)
	}
}
