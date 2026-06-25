package management

// TrustExchangeProvider names a known OIDC issuer shape. The enum is open at
// the SDK boundary so a new provider on the server (e.g. an addition between
// SDK releases) doesn't break consumers.
type TrustExchangeProvider string

const (
	TrustExchangeProviderGitHub     TrustExchangeProvider = "github"
	TrustExchangeProviderGoogle     TrustExchangeProvider = "google"
	TrustExchangeProviderAuth0      TrustExchangeProvider = "auth0"
	TrustExchangeProviderClerk      TrustExchangeProvider = "clerk"
	TrustExchangeProviderCustomOIDC TrustExchangeProvider = "custom_oidc"
)

// TrustExchange is the read shape of a trust exchange. The CEL Policy is the
// gate that decides whether an incoming assertion can mint an output token
// against OutputKeySpecID.
type TrustExchange struct {
	ID               string                `json:"id"`
	Name             string                `json:"name"`
	Description      string                `json:"description,omitempty"`
	Type             string                `json:"type"`
	Provider         TrustExchangeProvider `json:"provider"`
	Issuer           string                `json:"issuer"`
	DiscoveryURL     string                `json:"discovery_url,omitempty"`
	JWKSURL          string                `json:"jwks_url,omitempty"`
	AllowedAudiences []string              `json:"allowed_audiences"`
	Policy           string                `json:"policy"`
	OutputKeySpecID  string                `json:"output_key_spec_id"`
	Active           bool                  `json:"active"`
	// UpstreamClientID echoes the OIDC relying-party client id Microwave uses to
	// broker an interactive login at the exchange's upstream issuer. The matching
	// secret is write-only and never returned.
	UpstreamClientID string `json:"upstream_client_id,omitempty"`
	CreatedAt        Time   `json:"created_at"`
	UpdatedAt        Time   `json:"updated_at"`
}

// TrustExchangeInput is the write shape for Create and Update. Active is a
// pointer so an omitted field defaults to active while explicit false is
// honored; this matches the server contract.
type TrustExchangeInput struct {
	Name             string                `json:"name"`
	Description      string                `json:"description,omitempty"`
	Type             string                `json:"type"`
	Provider         TrustExchangeProvider `json:"provider"`
	Issuer           string                `json:"issuer"`
	DiscoveryURL     string                `json:"discovery_url,omitempty"`
	JWKSURL          string                `json:"jwks_url,omitempty"`
	AllowedAudiences []string              `json:"allowed_audiences"`
	Policy           string                `json:"policy"`
	OutputKeySpecID  string                `json:"output_key_spec_id"`
	Active           *bool                 `json:"active,omitempty"`
	// UpstreamClientID / UpstreamClientSecret register Microwave as an OIDC
	// relying party at the exchange's upstream issuer, enabling the brokered
	// interactive (authorization-code / device) CLI login. Set both to enable it;
	// the secret is write-only and never returned on read.
	UpstreamClientID     string `json:"upstream_client_id,omitempty"`
	UpstreamClientSecret string `json:"upstream_client_secret,omitempty"`
}
