package management

// TrustProviderType is the protocol a trust provider speaks. Only "oidc" is
// supported; the server rejects any other value.
type TrustProviderType = string

// TrustProviderTypeOIDC is the only trust provider type the server accepts.
const TrustProviderTypeOIDC TrustProviderType = "oidc"

// TrustProvider is the read shape of a trust provider. Where a Trust Exchange
// consumes an external OIDC assertion and mints a Microwave token, a Trust
// Provider does the inverse: it lets an external party authenticate against
// a Microwave-issued client key spec (ClientKeySpecID) and mint a token under
// the output key spec, with a CEL policy gating the mint.
type TrustProvider struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Type            string `json:"type"`
	ClientKeySpecID string `json:"client_key_spec_id"`
	OutputKeySpecID string `json:"output_key_spec_id"`
	Policy          string `json:"policy"`
	Active          bool   `json:"active"`
	CreatedAt       Time   `json:"created_at"`
	UpdatedAt       Time   `json:"updated_at"`
}

// TrustProviderInput is the write shape for Create and Update. Active is a
// pointer so an omitted field defaults to active while explicit false is
// honored, matching the server contract.
type TrustProviderInput struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Type            string `json:"type"`
	ClientKeySpecID string `json:"client_key_spec_id"`
	OutputKeySpecID string `json:"output_key_spec_id"`
	Policy          string `json:"policy"`
	Active          *bool  `json:"active,omitempty"`
}
