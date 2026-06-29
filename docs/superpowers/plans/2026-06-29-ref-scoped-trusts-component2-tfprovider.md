# Ref-scoped Trusts — Component 2 (SDK + TF provider glob_fields) Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expose `glob_fields` on trust federations through the microwave-go
management SDK and the Terraform provider, so the `sandbar-github-ci` federation
can declare `glob_fields = ["ref"]` (Component 3/rollout applies that TF).

**Architecture:** Two repos in dependency order. (A) microwave-go: add
`GlobFields []string` to the three `management` federation structs, mirroring
`IdentityFields`; merge + release a new tag. (B) terraform-provider-microwave:
bump microwave-go to that tag and add an optional `glob_fields` list attribute to
the `microwave_trust_federation` resource, wired through all five
identity_fields touchpoints.

**Tech Stack:** Go 1.26; terraform-plugin-framework; microwave-go SDK; the
Microwave management API (Component 1, merged, already accepts/returns
`glob_fields`).

## Global Constraints

- Go 1.26 (latest patch); never older.
- Conventional Commits; commits SSH-signed; no `Co-Authored-By` trailer.
- Mirror the EXISTING `identity_fields` field/attribute at every site, verbatim
  in shape (json tags, optionality, null/unknown guards) — only the name
  changes to `glob_fields` / `GlobFields`.
- `glob_fields` is OPTIONAL everywhere; omitting it must be byte-identical to
  today (empty → server default `[]` → exact matching). No required-field break.
- CROSS-REPO ORDER: Task 1 (microwave-go) must merge AND release a tag before
  Task 2 (provider) can `go get` it. Do not start Task 2 until the controller
  confirms the new microwave-go version is published.

---

### Task 1: microwave-go — `GlobFields` on the management federation types

**Repo:** microwave-go (worktree on branch `feat/management-glob-fields`).
**Files:**
- Modify: `management/types_trust_federations.go` (three structs)
- Test: `management/parity_test.go` (add a round-trip test)

**Interfaces:**
- Produces: `GlobFields []string` (json `glob_fields`, `omitempty` on the two
  write structs) on `management.TrustFederation`, `TrustFederationInput`,
  `TrustFederationUpdateInput`.

- [ ] **Step 1: Write the failing round-trip test**

In `management/parity_test.go`, mirroring the existing
`TestTrustFederationPolicyField`:

```go
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
```

Confirm `encoding/json` and `strings` are imported in the test file (the existing
tests use them).

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./management/ -run TestTrustFederationGlobFieldsField -v`
Expected: FAIL — `unknown field GlobFields`.

- [ ] **Step 3: Add the field to all three structs**

In `management/types_trust_federations.go`, add a line immediately after each
`IdentityFields` field:

- In `TrustFederation` (read): after `IdentityFields  []string \`json:"identity_fields"\``:
  ```go
	GlobFields      []string      `json:"glob_fields,omitempty"`
  ```
- In `TrustFederationInput` (create): after its `IdentityFields`:
  ```go
	GlobFields      []string      `json:"glob_fields,omitempty"`
  ```
- In `TrustFederationUpdateInput` (update): after its `IdentityFields  []string \`json:"identity_fields,omitempty"\``:
  ```go
	GlobFields      []string `json:"glob_fields,omitempty"`
  ```

(Match the surrounding struct's column formatting; run gofmt in Step 4.)

- [ ] **Step 4: Run the test + gofmt + build**

Run: `gofmt -w management/types_trust_federations.go && go test ./management/ -run TestTrustFederationGlobFieldsField -v && go build ./...`
Expected: gofmt silent; test PASS; build OK. Then `gofmt -l management/` must be empty.

- [ ] **Step 5: Commit**

```bash
git add management/types_trust_federations.go management/parity_test.go
git commit -m "feat(management): add GlobFields to trust federation types"
```

---

### Task 2: terraform-provider-microwave — `glob_fields` attribute

**PRECONDITION (controller-gated):** Task 1 is merged to microwave-go `main`
AND a new tag (e.g. `v0.14.0`) is published. Do NOT start until the controller
provides the exact version string. Work in a terraform-provider-microwave
worktree.

**Files:**
- Modify: `go.mod` / `go.sum` (bump `github.com/microwave-sh/microwave-go`)
- Modify: `internal/provider/resource_trust_federation.go` (model, schema, the
  three mapping funcs)
- Modify: `internal/provider/errors.go` (`trustFederationFields` map)
- Test: `internal/provider/resource_trust_federation_test.go`

**Interfaces:**
- Consumes: `management.TrustFederation{}.GlobFields`,
  `TrustFederationInput{}.GlobFields`, `TrustFederationUpdateInput{}.GlobFields`
  (Task 1); existing `stringListToSlice`/`stringSliceToList` in
  `internal/provider/helpers.go`.

- [ ] **Step 1: Bump the SDK dependency**

Run (use the exact version the controller provides for `<VER>`):
```bash
go get github.com/microwave-sh/microwave-go@<VER>
go mod tidy
```
Expected: `go.mod` shows the new version; `go build ./...` still compiles.

- [ ] **Step 2: Write the failing test**

In `internal/provider/resource_trust_federation_test.go`, mirror an existing
`identity_fields` round-trip/mapping test. If the file tests the wire mappers
directly, add:

```go
func TestTrustFederationGlobFieldsRoundTrip(t *testing.T) {
	ctx := context.Background()
	m := &trustFederationModel{}
	// out has GlobFields from the API; fromWire must populate the model.
	out := &management.TrustFederation{
		IdentityFields: []string{"repository", "ref"},
		GlobFields:     []string{"ref"},
	}
	if diags := trustFederationFromWire(ctx, m, out); diags.HasError() {
		t.Fatalf("fromWire: %v", diags)
	}
	var got []string
	m.GlobFields.ElementsAs(ctx, &got, false)
	if len(got) != 1 || got[0] != "ref" {
		t.Fatalf("model.GlobFields = %v, want [ref]", got)
	}
	// toWire must send it back.
	in, diags := trustFederationToWire(ctx, m)
	if diags.HasError() {
		t.Fatalf("toWire: %v", diags)
	}
	if len(in.GlobFields) != 1 || in.GlobFields[0] != "ref" {
		t.Fatalf("wire.GlobFields = %v, want [ref]", in.GlobFields)
	}
}
```

If the existing tests are acceptance-style (`resource.Test` with TF config) and
the mappers are unexported but reachable in-package, the above works
(same-package test). If the package conventions differ, mirror the closest
existing `identity_fields` test exactly.

- [ ] **Step 3: Run to verify it fails**

Run: `go test ./internal/provider/ -run TestTrustFederationGlobFieldsRoundTrip -v`
Expected: FAIL — `m.GlobFields` undefined.

- [ ] **Step 4: Add the model field + schema attribute**

In `internal/provider/resource_trust_federation.go`:

- `trustFederationModel` (after `IdentityFields  types.List \`tfsdk:"identity_fields"\``):
  ```go
	GlobFields      types.List   `tfsdk:"glob_fields"`
  ```
- In `Schema`, after the `"identity_fields": schema.ListAttribute{...}` block,
  add a sibling (copy that block, rename, keep `ElementType: types.StringType`,
  mark `Optional: true`; if identity_fields is `Required`, glob_fields is
  `Optional` — it is not required):
  ```go
	"glob_fields": schema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Description: "Identity fields matched as glob patterns (trailing-* prefix or literal) instead of exact. Must be a subset of identity_fields.",
	},
  ```

- [ ] **Step 5: Wire the three mapping functions**

In `trustFederationToWire` (after `IdentityFields: fields,` in the returned
`*management.TrustFederationInput`):
```go
	globs, gdiags := stringListToSlice(ctx, m.GlobFields)
	diags.Append(gdiags...)
	// ... set GlobFields: globs in the returned struct literal
```
(Read the function and follow its existing diag-collection idiom; set
`GlobFields: globs` on the returned input.)

In `trustFederationUpdatePatch`, mirror the existing null/unknown-guarded
`IdentityFields` block:
```go
	if !plan.GlobFields.IsNull() && !plan.GlobFields.IsUnknown() {
		globs, gdiags := stringListToSlice(ctx, plan.GlobFields)
		diags.Append(gdiags...)
		patch.GlobFields = globs
	}
```

In `trustFederationFromWire`, after the `m.IdentityFields = fields` assignment:
```go
	globs, gdiags := stringSliceToList(ctx, out.GlobFields)
	diags.Append(gdiags...)
	m.GlobFields = globs
```
(Match the exact diag/return idiom used for IdentityFields right above each
site.)

- [ ] **Step 6: Update the error-field map**

In `internal/provider/errors.go`, add to `trustFederationFields`:
```go
	"glob_fields":        "glob_fields",
```

- [ ] **Step 7: Run the test + build + gofmt**

Run: `go test ./internal/provider/ -run TestTrustFederationGlobFieldsRoundTrip -v && go build ./... && gofmt -l internal/`
Expected: test PASS; build OK; gofmt silent.

- [ ] **Step 8: Run the provider's federation test suite (no regressions)**

Run: `go test ./internal/provider/ -run TrustFederation -count=1`
Expected: `ok` (existing federation tests unaffected).

- [ ] **Step 9: Commit**

```bash
git add go.mod go.sum internal/provider/resource_trust_federation.go internal/provider/errors.go internal/provider/resource_trust_federation_test.go
git commit -m "feat: add glob_fields to microwave_trust_federation resource"
```

---

## Self-Review

**Spec coverage (Component 2):** "Add an optional `glob_fields` (list of string)
attribute to the `microwave_trust_federation` resource, plumbed to the federation
create/update API; round-trips like `identity_fields`; existing configs default
to empty → exact behavior." Task 2 implements the attribute + all five
touchpoints; Task 1 is the SDK prerequisite the spec implies (the provider talks
to the API via microwave-go). ✓

**Placeholder scan:** none — full struct lines, schema block, mapping snippets,
and test code are concrete. Step 5's mapping snippets say "follow the existing
diag idiom" because the exact `diags` variable/return shape must match the
function being edited; the implementer reads the adjacent IdentityFields lines —
this is mirroring, not a placeholder.

**Type consistency:** `GlobFields []string` (SDK) ↔ `GlobFields types.List`
(provider model) ↔ `glob_fields` (tfsdk tag + json tag + error-map key)
consistent across both tasks. `stringListToSlice`/`stringSliceToList` are the
existing helpers used for IdentityFields.

**Cross-repo gate:** Task 2's PRECONDITION blocks on Task 1's release; the
controller supplies the version for Step 1.
