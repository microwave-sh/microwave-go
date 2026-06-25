package management

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

// SigningKeySet is the read shape of a signing key set. It is returned bare by
// Create; Get returns it nested inside a SigningKeySetDetail alongside the
// individual keys.
type SigningKeySet struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Kind      SigningKeySetKind `json:"kind"`
	Algorithm string            `json:"algorithm"`
	CreatedAt Time              `json:"created_at"`
	DeletedAt *Time             `json:"deleted_at,omitempty"`
}

// SigningKeySetDetail is the read shape returned by Get: the set plus its
// individual keys. Keys rotate on their own schedule via the per-key endpoints.
type SigningKeySetDetail struct {
	Set  SigningKeySet `json:"set"`
	Keys []SigningKey  `json:"keys"`
}

// SigningKey is one key inside a signing key set. AsymmetricPublicMaterial is
// populated only for asymmetric kinds; SecretRef points at the stored secret
// for symmetric kinds.
type SigningKey struct {
	ID                       string                    `json:"id"`
	SetID                    string                    `json:"set_id"`
	Status                   string                    `json:"status"`
	AsymmetricPublicMaterial *AsymmetricPublicMaterial `json:"asymmetric_public_material,omitempty"`
	SecretRef                string                    `json:"secret_ref,omitempty"`
	CreatedAt                Time                      `json:"created_at"`
	RevokedAt                *Time                     `json:"revoked_at,omitempty"`
}

// AsymmetricPublicMaterial is the JWK public material for an asymmetric signing
// key.
type AsymmetricPublicMaterial struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use,omitempty"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
	Crv string `json:"crv,omitempty"`
}

// SigningKeySetInput is the write shape for Create. Kind and Algorithm are
// immutable once set; renaming is done via Update (PATCH).
type SigningKeySetInput struct {
	Name      string            `json:"name"`
	Kind      SigningKeySetKind `json:"kind"`
	Algorithm string            `json:"algorithm"`
}
