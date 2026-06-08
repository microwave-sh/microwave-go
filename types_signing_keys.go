package microwave

// SigningKeySetKind is the algorithmic family of a signing key set. The kind
// is part of the composite primary key (kind, name) — that's why GET / PATCH /
// DELETE paths take both segments.
type SigningKeySetKind string

const (
	// SigningKeySetKindAsymmetric covers RS256, ES256, EdDSA, etc.
	SigningKeySetKindAsymmetric SigningKeySetKind = "asymmetric"
	// SigningKeySetKindSymmetric covers HS256/384/512.
	SigningKeySetKindSymmetric SigningKeySetKind = "symmetric"
)

// SigningKeySet is the read shape of a signing key set. Individual keys are
// managed via separate /api/signing-key-sets/{kind}/{name}/keys/* endpoints and
// not modelled in v1 of this SDK — set-level CRUD is the typical IaC surface;
// individual keys rotate on their own schedule.
type SigningKeySet struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Kind      SigningKeySetKind `json:"kind"`
	Algorithm string            `json:"algorithm"`
	CreatedAt Time              `json:"created_at"`
	DeletedAt *Time             `json:"deleted_at,omitempty"`
}

// SigningKeySetInput is the write shape for Create. Update is not currently
// supported by the API (algorithm + kind are immutable; renaming requires
// recreate). Keep Input symmetric with the wire contract regardless.
type SigningKeySetInput struct {
	Name      string            `json:"name"`
	Kind      SigningKeySetKind `json:"kind"`
	Algorithm string            `json:"algorithm"`
}

// SigningKeySetList is the paginated list response.
type SigningKeySetList struct {
	Items []SigningKeySet `json:"items"`
}
