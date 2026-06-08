package microwave

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Error wraps a non-2xx response. The Microwave API uses Problem+JSON style
// envelopes (status / title / detail); when the body doesn't decode as one,
// the raw body is preserved so callers can surface it.
type Error struct {
	StatusCode int    `json:"status"`
	Title      string `json:"title"`
	Detail     string `json:"detail"`
	Type       string `json:"type,omitempty"`
	Instance   string `json:"instance,omitempty"`
	RawBody    string `json:"-"`
}

func (e *Error) Error() string {
	switch {
	case e.Detail != "":
		return fmt.Sprintf("microwave: %d %s: %s", e.StatusCode, e.Title, e.Detail)
	case e.Title != "":
		return fmt.Sprintf("microwave: %d %s", e.StatusCode, e.Title)
	case e.RawBody != "":
		return fmt.Sprintf("microwave: %d: %s", e.StatusCode, e.RawBody)
	default:
		return fmt.Sprintf("microwave: %d %s", e.StatusCode, http.StatusText(e.StatusCode))
	}
}

// IsNotFound reports whether err represents a 404. Callers use this to make
// idempotent "delete if exists" / "create if absent" flows readable.
func IsNotFound(err error) bool {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsConflict reports whether err represents a 409 — typically a unique-name
// collision on Create or a state-machine violation on Update.
func IsConflict(err error) bool {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusConflict
	}
	return false
}

func decodeError(resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)
	apiErr := &Error{
		StatusCode: resp.StatusCode,
		RawBody:    string(raw),
	}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, apiErr)
	}
	if apiErr.StatusCode == 0 {
		apiErr.StatusCode = resp.StatusCode
	}
	if apiErr.Title == "" {
		apiErr.Title = http.StatusText(resp.StatusCode)
	}
	return apiErr
}
