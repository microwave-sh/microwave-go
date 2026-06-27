package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// recordingProgress captures the phase events emitted during a login so tests
// can assert the Begin/Succeed/Fail sequence a CLI would render.
type recordingProgress struct{ events []string }

func (p *recordingProgress) Begin(m string)   { p.events = append(p.events, "begin:"+m) }
func (p *recordingProgress) Succeed(m string) { p.events = append(p.events, "ok:"+m) }
func (p *recordingProgress) Fail(m string)    { p.events = append(p.events, "fail:"+m) }

func TestLoginDeviceApprovalReportsProgress(t *testing.T) {
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
		}
	}))
	defer srv.Close()

	pr := &recordingProgress{}
	if _, err := loginDeviceApproval(context.Background(), LoginConfig{
		DeviceApprovalURL: srv.URL,
		OpenBrowser:       func(string) error { return nil },
		Progress:          pr,
	}, srv.Client()); err != nil {
		t.Fatalf("loginDeviceApproval: %v", err)
	}

	want := []string{
		"begin:Starting device authorization",
		"ok:Device authorization started",
		"begin:Waiting for approval",
		"ok:Approved",
	}
	if len(pr.events) != len(want) {
		t.Fatalf("events = %v, want %v", pr.events, want)
	}
	for i := range want {
		if pr.events[i] != want[i] {
			t.Fatalf("event[%d] = %q, want %q (full: %v)", i, pr.events[i], want[i], pr.events)
		}
	}
}

func TestLoginDeviceApprovalReportsFailOnDeny(t *testing.T) {
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

	pr := &recordingProgress{}
	if _, err := loginDeviceApproval(context.Background(), LoginConfig{
		DeviceApprovalURL: srv.URL,
		OpenBrowser:       func(string) error { return nil },
		Progress:          pr,
	}, srv.Client()); err == nil {
		t.Fatalf("expected a denied error")
	}
	// The last event must be a Fail so the CLI shows a red ✗, not a hung spinner.
	last := pr.events[len(pr.events)-1]
	if last != "fail:Login denied" {
		t.Fatalf("last event = %q, want fail:Login denied (full: %v)", last, pr.events)
	}
}

// A nil Progress reporter must be a no-op (the default), not a panic.
func TestLoginDeviceApprovalNilProgressIsNoop(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device":
			_ = json.NewEncoder(w).Encode(map[string]any{"device_code": "dc1", "user_code": "AAAA-0000", "verification_uri": "http://c/device", "expires_in": 300, "interval": 0})
		case "/auth/device/token":
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "t", "token_type": "Bearer", "expires_in": 60})
		}
	}))
	defer srv.Close()

	if _, err := loginDeviceApproval(context.Background(), LoginConfig{
		DeviceApprovalURL: srv.URL,
		OpenBrowser:       func(string) error { return nil },
	}, srv.Client()); err != nil {
		t.Fatalf("nil Progress should be a no-op: %v", err)
	}
}
