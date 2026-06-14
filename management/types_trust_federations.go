package management

// FederationKey is the catalog key that identifies which trust federation
// template applies to a binding. The enum is open at the SDK boundary so new
// entries added server-side (e.g. between SDK releases) don't break consumers.
type FederationKey string

// TrustFederation is the read shape of a trust federation definition. Each
// federation defines the OIDC issuer, audience, identity fields, and optional
// CEL policy override used when evaluating federation redemption requests.
type TrustFederation struct {
	ID              string        `json:"id"`
	WorkspaceID     string        `json:"workspace_id"`
	Key             FederationKey `json:"key"`
	Label           string        `json:"label"`
	Description     string        `json:"description"`
	LogoURL         string        `json:"logo_url"`
	DocsURL         string        `json:"docs_url"`
	Issuer          string        `json:"issuer"`
	Audience        string        `json:"audience"`
	IdentityFields  []string      `json:"identity_fields"`
	OutputKeySpecID string        `json:"output_key_spec_id,omitempty"`
	PolicyOverride  string        `json:"policy_override,omitempty"`
	CreatedAt       Time          `json:"created_at"`
	UpdatedAt       Time          `json:"updated_at"`
}

// TrustFederationInput is the write shape for creating a trust federation.
type TrustFederationInput struct {
	Key             FederationKey `json:"key"`
	Label           string        `json:"label"`
	Description     string        `json:"description,omitempty"`
	LogoURL         string        `json:"logo_url,omitempty"`
	DocsURL         string        `json:"docs_url,omitempty"`
	Issuer          string        `json:"issuer,omitempty"`
	Audience        string        `json:"audience,omitempty"`
	IdentityFields  []string      `json:"identity_fields"`
	OutputKeySpecID string        `json:"output_key_spec_id,omitempty"`
	PolicyOverride  string        `json:"policy_override,omitempty"`
}

// TrustFederationUpdateInput is the write shape for patching a trust
// federation. Pointer fields distinguish nil (skip) from &"" (clear).
type TrustFederationUpdateInput struct {
	Label           *string  `json:"label,omitempty"`
	Description     *string  `json:"description,omitempty"`
	LogoURL         *string  `json:"logo_url,omitempty"`
	DocsURL         *string  `json:"docs_url,omitempty"`
	Issuer          *string  `json:"issuer,omitempty"`
	Audience        *string  `json:"audience,omitempty"`
	IdentityFields  []string `json:"identity_fields,omitempty"`
	OutputKeySpecID *string  `json:"output_key_spec_id,omitempty"`
	PolicyOverride  *string  `json:"policy_override,omitempty"`
}

// RedeemFederationInput is the request body for the federation redemption
// endpoint. Token is an OIDC JWT issued by the federation's configured issuer.
type RedeemFederationInput struct {
	Token string `json:"token"`
}

// RedeemFederationResult is the response body from a successful federation
// redemption. Token is a minted Microwave token valid until ExpiresAt.
type RedeemFederationResult struct {
	Token     string         `json:"token"`
	ExpiresAt Time           `json:"expires_at"`
	Scopes    []string       `json:"scopes"`
	Claims    map[string]any `json:"claims"`
}

// TrustFederationBinding is the read shape of a trust federation binding.
// A binding pins a specific identity tuple (e.g. a GitHub repository + ref)
// to a federation and optionally asserts output claims on every minted token.
type TrustFederationBinding struct {
	ID            string         `json:"id"`
	WorkspaceID   string         `json:"workspace_id"`
	FederationKey FederationKey  `json:"federation_key"`
	Identity      map[string]any `json:"identity"`
	OutputClaims  map[string]any `json:"output_claims,omitempty"`
	CreatedAt     Time           `json:"created_at"`
	UpdatedAt     Time           `json:"updated_at"`
}

// TrustFederationBindingInput is the write shape for creating a trust
// federation binding. Bindings are immutable post-create; to change identity
// claims, delete and recreate.
type TrustFederationBindingInput struct {
	FederationKey FederationKey  `json:"federation_key"`
	Identity      map[string]any `json:"identity"`
	OutputClaims  map[string]any `json:"output_claims,omitempty"`
}
