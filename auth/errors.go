package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Error wraps a non-2xx response from the Auth plane. Failed exchanges
// against a known-good exchange ID (policy denial, expired assertion,
// audience mismatch) come back as 200 with valid=false in the body — those
// are NOT Errors, they're ExchangeResults with Code populated. This type
// covers transport-level failures: 404 (unknown exchange), 400 (malformed
// body), 5xx, network errors.
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
		return fmt.Sprintf("microwave/auth: %d %s: %s", e.StatusCode, e.Title, e.Detail)
	case e.Title != "":
		return fmt.Sprintf("microwave/auth: %d %s", e.StatusCode, e.Title)
	case e.RawBody != "":
		return fmt.Sprintf("microwave/auth: %d: %s", e.StatusCode, e.RawBody)
	default:
		return fmt.Sprintf("microwave/auth: %d %s", e.StatusCode, http.StatusText(e.StatusCode))
	}
}

// IsNotFound reports whether err represents a 404 — typically a malformed or
// unknown exchange ID.
func IsNotFound(err error) bool {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
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
