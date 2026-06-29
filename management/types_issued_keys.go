package management

import "time"

// IssueKeyInput is the write contract for issuing a key from a spec. Mirrors
// the server's dto.IssueKeyInput. Subject is the principal the key acts as
// (AuthPipe sets it to the customer workspace id); Name is the human label
// surfaced in console lists and is required by the server for opaque specs.
type IssueKeyInput struct {
	Subject   string         `json:"subject"`
	Audience  string         `json:"audience,omitempty"`
	Name      string         `json:"name,omitempty"`
	Scopes    []string       `json:"scopes,omitempty"`
	Claims    map[string]any `json:"claims,omitempty"`
	ExpiresIn string         `json:"expires_in,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// IssuedKeyResult is the response from issuing a key. Key is the raw secret,
// returned exactly once. Mirrors the server's dto.IssueKeyResult.
type IssuedKeyResult struct {
	ID        string     `json:"id"`
	Key       string     `json:"key"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Subject   string     `json:"subject"`
	SpecID    string     `json:"spec_id"`
	Scopes    []string   `json:"scopes"`
	CreatedAt time.Time  `json:"created_at"`
}

// IssuedKey is the read shape of an issued key (search results). Mirrors the
// server's dto.IssuedKey. Status is one of "active", "revoked", "expired",
// "rotating".
type IssuedKey struct {
	ID             string         `json:"id"`
	SpecID         string         `json:"spec_id"`
	Subject        string         `json:"subject"`
	Name           string         `json:"name"`
	Scopes         []string       `json:"scopes"`
	Claims         map[string]any `json:"claims"`
	Metadata       map[string]any `json:"metadata"`
	Status         string         `json:"status"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	LastVerifiedAt *time.Time     `json:"last_verified_at,omitempty"`
	RevokedAt      *time.Time     `json:"revoked_at,omitempty"`
}
