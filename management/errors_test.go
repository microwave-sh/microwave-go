package management_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/microwave-sh/microwave-go/management"
)

// TestError_SurfacesDomainMessage pins that a {type, message} domain envelope
// (e.g. a CEL policy compile failure) reaches the caller instead of a bare
// "422 Unprocessable Entity".
func TestError_SurfacesDomainMessage(t *testing.T) {
	e := &management.Error{StatusCode: 422, Title: "Unprocessable Entity"}
	if err := json.Unmarshal([]byte(`{"type":"invalid_input","message":"trust exchange policy compile error: undeclared reference to 'output'"}`), e); err != nil {
		t.Fatal(err)
	}
	got := e.Error()
	if !strings.Contains(got, "undeclared reference to 'output'") {
		t.Fatalf("Error() = %q, want it to surface the compile detail", got)
	}
}

// TestError_SurfacesFieldErrors pins that Problem+JSON field-level errors[] are
// surfaced, so the caller sees which field is wrong and why.
func TestError_SurfacesFieldErrors(t *testing.T) {
	e := &management.Error{}
	body := `{"title":"Unprocessable Entity","status":422,"detail":"validation failed","errors":[{"message":"expected value to be one of \"github, google, auth0, custom_oidc\"","location":"body.provider","value":"clerk"}]}`
	if err := json.Unmarshal([]byte(body), e); err != nil {
		t.Fatal(err)
	}
	got := e.Error()
	if !strings.Contains(got, "one of") || !strings.Contains(got, "body.provider") {
		t.Fatalf("Error() = %q, want it to surface the field-level error", got)
	}
}
