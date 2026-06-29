package management_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/microwave-sh/microwave-go/management"
)

// TestKeySpecClaimsWireShape pins the claim contract, override policy, and
// webhook config to the server's JSON shape. These three nested types silently
// drifted (claims sent as {per,wildcard}, override as {claims:[]}, webhook as
// {url}) which dropped every claim/override/webhook on the wire. Guard the
// exact keys so a regression fails loudly instead of silently no-op-ing.
func TestKeySpecClaimsWireShape(t *testing.T) {
	var wsID any = "ws_42"
	in := management.KeySpecInput{
		Name:   "sandbar-cli-session",
		Format: management.KeyFormatJWT,
		Claims: management.ClaimsConfig{
			AllowUnlisted: true,
			Claims: []management.ClaimPolicy{
				{Key: "workspace_id", Mode: "allowed", Value: &wsID},
			},
		},
		OverridePolicy: management.OverridePolicy{AllowCustomScopes: true},
		Webhooks:       management.WebhookConfig{Endpoint: "https://example.test/hook"},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)

	mustContain := []string{
		`"claims":[`,                             // ClaimsConfig.Claims (not "per"/"wildcard")
		`"allow_unlisted":true`,                  // ClaimsConfig.AllowUnlisted
		`"key":"workspace_id"`,                   // ClaimPolicy.Key (was entirely absent)
		`"mode":"allowed"`,                       // ClaimPolicy.Mode
		`"allow_custom_scopes":true`,             // OverridePolicy (not "claims")
		`"endpoint":"https://example.test/hook"`, // WebhookConfig (not "url")
	}
	for _, want := range mustContain {
		if !strings.Contains(got, want) {
			t.Errorf("KeySpecInput JSON missing %q\n got: %s", want, got)
		}
	}
	mustNotContain := []string{`"per":`, `"wildcard":`, `"url":`}
	for _, bad := range mustNotContain {
		if strings.Contains(got, bad) {
			t.Errorf("KeySpecInput JSON still emits stale key %q\n got: %s", bad, got)
		}
	}
}

// TestSigningKeySetGetDecodesDetail pins that Get decodes the {set,keys}
// envelope, not a bare SigningKeySet (which left every field zero).
func TestSigningKeySetGetDecodesDetail(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(management.SigningKeySetDetail{
			Set:  management.SigningKeySet{ID: "sks_1", Name: "sandbar-cli-sessions", Kind: management.SigningKeySetKindAsymmetric, Algorithm: "ES256"},
			Keys: []management.SigningKey{{ID: "key_1", SetID: "sks_1", Status: "active"}},
		})
	}))

	out, err := client.SigningKeySets.Get(context.Background(), management.SigningKeySetKindAsymmetric, "sandbar-cli-sessions")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if out.Set.ID != "sks_1" || out.Set.Algorithm != "ES256" {
		t.Errorf("set not decoded: %+v", out.Set)
	}
	if len(out.Keys) != 1 || out.Keys[0].ID != "key_1" {
		t.Errorf("keys not decoded: %+v", out.Keys)
	}
}

// TestTrustFederationPolicyField pins that the primary CEL policy round-trips —
// the SDK previously had no `policy` field at all, only `policy_override`.
func TestTrustFederationPolicyField(t *testing.T) {
	in := management.TrustFederationInput{
		Key:    management.FederationKey("terraform_cloud"),
		Label:  "Terraform Cloud",
		Policy: `assertion.sub != ""`,
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"policy":"assertion.sub != \"\""`) {
		t.Errorf("TrustFederationInput JSON missing policy field\n got: %s", b)
	}
}

// TestTrustFederationGlobFieldsField pins that GlobFields round-trip on
// trust federation types. The server accepts glob_fields on federations;
// the SDK must preserve this field.
func TestTrustFederationGlobFieldsField(t *testing.T) {
	in := management.TrustFederationInput{
		Key:            "k",
		IdentityFields: []string{"repository", "ref"},
		GlobFields:     []string{"ref"},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"glob_fields":["ref"]`) {
		t.Errorf("TrustFederationInput JSON missing glob_fields\n got: %s", b)
	}

	var out management.TrustFederation
	if err := json.Unmarshal([]byte(`{"glob_fields":["ref"]}`), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.GlobFields) != 1 || out.GlobFields[0] != "ref" {
		t.Errorf("TrustFederation.GlobFields = %v, want [ref]", out.GlobFields)
	}

	// omitempty: an empty GlobFields must NOT appear in the write JSON.
	empty, _ := json.Marshal(management.TrustFederationInput{Key: "k", IdentityFields: []string{"repository"}})
	if strings.Contains(string(empty), "glob_fields") {
		t.Errorf("empty GlobFields must be omitted, got: %s", empty)
	}
}
