package management

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Error wraps a non-2xx response. The Microwave API returns two envelope
// shapes: a Problem+JSON style {status, title, detail, errors[]} for request
// schema validation, and {type, message} for domain validation (e.g. a CEL
// policy compile failure). We decode both so the human-actionable detail — the
// field that's wrong, or the compile error — always reaches the caller instead
// of a bare "422 Unprocessable Entity". When neither decodes, the raw body is
// preserved.
type Error struct {
	StatusCode int           `json:"status"`
	Title      string        `json:"title"`
	Detail     string        `json:"detail"`
	Message    string        `json:"message"`
	Type       string        `json:"type,omitempty"`
	Instance   string        `json:"instance,omitempty"`
	Errors     []ErrorDetail `json:"errors,omitempty"`
	RawBody    string        `json:"-"`
}

// ErrorDetail is one field-level validation failure from the {errors[]}
// envelope. The two server error shapes name the offending field differently:
// the Huma schema envelope uses `location` (+ `value`), the Mataki domain
// envelope uses `field` (+ `code`). Both are decoded so the field is surfaced
// regardless of which envelope produced it.
type ErrorDetail struct {
	Message  string `json:"message"`
	Location string `json:"location,omitempty"`
	Value    any    `json:"value,omitempty"`
	Field    string `json:"field,omitempty"`
	Code     string `json:"code,omitempty"`
}

func (e *Error) Error() string {
	// Prefer the most specific human-readable detail: the Problem `detail`, then
	// the domain `message`, then the raw body.
	detail := e.Detail
	if detail == "" {
		detail = e.Message
	}
	var b strings.Builder
	fmt.Fprintf(&b, "microwave: %d %s", e.StatusCode, e.Title)
	switch {
	case detail != "":
		fmt.Fprintf(&b, ": %s", detail)
	case len(e.Errors) == 0 && e.RawBody != "":
		fmt.Fprintf(&b, ": %s", e.RawBody)
	}
	// Append field-level validation errors so callers see WHICH field failed and
	// why (e.g. provider must be one of …; policy compile error at body.policy).
	for _, fe := range e.Errors {
		fmt.Fprintf(&b, "\n  - %s", fe.Message)
		if loc := fe.Location; loc != "" {
			fmt.Fprintf(&b, " (%s)", loc)
		} else if fe.Field != "" {
			fmt.Fprintf(&b, " (%s)", fe.Field)
		}
	}
	return b.String()
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
