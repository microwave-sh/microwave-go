package auth

import (
	"context"
	"net/http"
	"net/url"
)

// TokenExchangeService is the Auth plane surface for redeeming inbound OIDC
// assertions against configured Trust Exchanges.
type TokenExchangeService struct {
	client *Client
}

// ExchangeResult is the decoded response from a Redeem call. Two outcomes:
//
//   - Valid==true: JWT holds the minted Microwave session token; use it as
//     the Bearer credential against the Management API.
//   - Valid==false: Code names the denial reason (e.g. "policy_denied",
//     "audience_mismatch", "issuer_unknown", "assertion_expired").
//     RuleResults breaks down which CEL clauses passed and failed when the
//     denial came from policy evaluation.
//
// A non-nil error from Redeem indicates a transport-level failure (network,
// 5xx, unknown exchange ID), not a denied exchange — denied exchanges return
// a Valid==false result with no error.
type ExchangeResult struct {
	Valid       bool            `json:"valid"`
	Code        string          `json:"code,omitempty"`
	JWT         string          `json:"jwt,omitempty"`
	Subject     string          `json:"subject,omitempty"`
	Scopes      []string        `json:"scopes,omitempty"`
	Claims      map[string]any  `json:"claims,omitempty"`
	Inbound     map[string]any  `json:"inbound,omitempty"`
	RuleResults map[string]bool `json:"rule_results,omitempty"`
}

// Redeem exchanges an inbound OIDC token for a Microwave session JWT under
// the named Trust Exchange. The exchangeID identifies the server-side rule
// (issuer + allowed audiences + CEL policy + output key spec) that gates the
// redemption.
func (s *TokenExchangeService) Redeem(ctx context.Context, exchangeID, inboundToken string) (*ExchangeResult, error) {
	path := "/trust-exchanges/" + url.PathEscape(exchangeID) + "/exchange"
	body := struct {
		Token string `json:"token"`
	}{Token: inboundToken}

	var out ExchangeResult
	if err := s.client.doRequest(ctx, http.MethodPost, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
