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
	var sawAuthorizeURL string
	polls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device":
			_ = json.NewEncoder(w).Encode(map[string]any{"device_code": "dc1", "user_code": "WXYZ", "authorize_url": "http://console/approve?uc=WXYZ", "expires_in": 300, "interval": 0})
		case "/auth/device/token":
			polls++
			if polls < 2 {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "pending"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "approved", "token": "session.jwt", "expires_in": 3600})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out bytes.Buffer
	creds, err := loginDeviceApproval(context.Background(), LoginConfig{
		DeviceApprovalURL: srv.URL,
		Output:            &out,
		OpenBrowser:       func(u string) error { sawAuthorizeURL = u; return nil },
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
	if sawAuthorizeURL == "" {
		t.Fatal("expected the authorize URL to be opened")
	}
	if !strings.Contains(out.String(), "WXYZ") {
		t.Fatalf("expected the user code shown in output, got %q", out.String())
	}
}

func TestLoginDeviceApprovalDeniedIsTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device":
			_ = json.NewEncoder(w).Encode(map[string]any{"device_code": "dc1", "user_code": "U", "authorize_url": "http://c/a", "expires_in": 300, "interval": 0})
		case "/auth/device/token":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "denied"})
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
			_ = json.NewEncoder(w).Encode(map[string]any{"device_code": "dc", "user_code": "UC", "authorize_url": "http://example/approve", "expires_in": 300, "interval": 0})
		case "/auth/device/token":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "approved", "token": "tok", "expires_in": 3600})
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
