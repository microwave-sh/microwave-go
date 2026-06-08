package management

// KeyFormat is the on-the-wire shape of keys issued from a spec. Opaque means
// stateful + Microwave-verified + revocable (an "API key"). JWT means
// stateless + client-verifiable + never stored.
type KeyFormat string

const (
	KeyFormatOpaque KeyFormat = "opaque"
	KeyFormatJWT    KeyFormat = "jwt"
)

// OpaqueConfig is the format-specific config for Format == "opaque".
type OpaqueConfig struct {
	// Prefix is the visible product-namespaced prefix on issued keys, e.g.
	// "sbr_live_" for a Sandbar production key spec.
	Prefix string `json:"prefix,omitempty"`
}

// JWTConfig is the format-specific config for Format == "jwt". Issuer is
// server-derived (https://{spec-id}.microwave.sh) and only meaningful in
// responses — Create/Update ignore it.
type JWTConfig struct {
	Algorithm string `json:"algorithm,omitempty"`
	Issuer    string `json:"issuer,omitempty"`
}

// ExpiryPolicy describes the TTL and rotation rules for keys issued from a
// spec. DefaultTTL/MaxTTL are Go-style duration strings ("24h", "30d", "0s"
// for never).
type ExpiryPolicy struct {
	DefaultTTL           string `json:"default_ttl"`
	MaxTTL               string `json:"max_ttl"`
	AllowNever           bool   `json:"allow_never"`
	RotationReminderDays int    `json:"rotation_reminder_days"`
}

// ClaimPolicy is one row of the unified claim contract. Mode is one of
// "default", "override", "required", "optional", "forbidden". Value is a
// pointer so the JSON null/omitted case is distinguishable from an explicit
// value — required for default/override modes.
type ClaimPolicy struct {
	Mode          string `json:"mode"`
	Value         *any   `json:"value,omitempty"`
	AllowedValues []any  `json:"allowed_values,omitempty"`
}

// ClaimsConfig is the full claim contract for a spec: per-claim policies plus
// a wildcard policy that applies to any claim not explicitly listed.
type ClaimsConfig struct {
	Per      map[string]ClaimPolicy `json:"per,omitempty"`
	Wildcard *ClaimPolicy           `json:"wildcard,omitempty"`
}

// OverridePolicy controls what the issuer can override at issue time. Empty
// list = nothing overridable.
type OverridePolicy struct {
	Claims []string `json:"claims,omitempty"`
}

// WebhookConfig describes the optional webhook subscription for spec
// lifecycle events.
type WebhookConfig struct {
	URL    string   `json:"url,omitempty"`
	Events []string `json:"events,omitempty"`
}

// KeySpec is the read shape of a key spec. PermissionSet is populated only by
// some endpoints (e.g. Get) — Create/Update return PermissionSetID only.
type KeySpec struct {
	ID                     string         `json:"id"`
	Name                   string         `json:"name"`
	Description            string         `json:"description,omitempty"`
	Format                 KeyFormat      `json:"format"`
	PermissionSetID        string         `json:"permission_set_id,omitempty"`
	PermissionSet          *PermissionSet `json:"permission_set,omitempty"`
	SigningKeySetID        string         `json:"signing_key_set_id,omitempty"`
	Opaque                 OpaqueConfig   `json:"opaque,omitempty"`
	JWT                    JWTConfig      `json:"jwt,omitempty"`
	Expiry                 ExpiryPolicy   `json:"expiry"`
	Claims                 ClaimsConfig   `json:"claims"`
	OverridePolicy         OverridePolicy `json:"override_policy"`
	Webhooks               WebhookConfig  `json:"webhooks"`
	WebhookSigningKeySetID string         `json:"webhook_signing_key_set_id,omitempty"`
	CreatedAt              Time           `json:"created_at"`
	UpdatedAt              Time           `json:"updated_at"`
}

// KeySpecInput is the write shape for Create and Update. Updates replace the
// configuration wholesale.
type KeySpecInput struct {
	Name                   string         `json:"name"`
	Description            string         `json:"description,omitempty"`
	Format                 KeyFormat      `json:"format"`
	PermissionSetID        string         `json:"permission_set_id,omitempty"`
	SigningKeySetID        string         `json:"signing_key_set_id,omitempty"`
	Opaque                 OpaqueConfig   `json:"opaque,omitempty"`
	JWT                    JWTConfig      `json:"jwt,omitempty"`
	Expiry                 ExpiryPolicy   `json:"expiry"`
	Claims                 ClaimsConfig   `json:"claims"`
	OverridePolicy         OverridePolicy `json:"override_policy"`
	Webhooks               WebhookConfig  `json:"webhooks"`
	WebhookSigningKeySetID string         `json:"webhook_signing_key_set_id,omitempty"`
}

// KeySpecList is the paginated list response.
type KeySpecList struct {
	Items []KeySpec `json:"items"`
}
