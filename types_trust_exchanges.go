package microwave

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
	CreatedAt        Time                  `json:"created_at"`
	UpdatedAt        Time                  `json:"updated_at"`
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
}

// TrustExchangeList is the paginated list response.
type TrustExchangeList struct {
	Items []TrustExchange `json:"items"`
}
