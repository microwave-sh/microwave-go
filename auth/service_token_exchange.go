package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// TokenExchangeResult is the decoded RFC 8693 token-exchange response from a
// RedeemTokenExchange call: a Microwave-issued JWT plus its lifetime and scope.
type TokenExchangeResult struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int
	Scope       string
}

// RedeemTokenExchange performs an RFC 8693 token exchange at an absolute token
// endpoint, swapping an inbound OIDC subject token for a Microwave-issued JWT.
// The resource indicator selects the server-side trust exchange or federation,
// so the caller is not pinned to a specific one — the endpoint + resource come
// from the relying party's own discovery (e.g. a CI provider's /auth/config).
//
// On an RFC 6749 §5.2 failure it returns a typed *OAuthError carrying the
// server's error code + error_description, so callers surface "invalid_grant:
// policy denied" rather than a bare "HTTP 400". httpClient may be nil.
func RedeemTokenExchange(ctx context.Context, httpClient *http.Client, tokenEndpoint, resource, subjectToken string) (*TokenExchangeResult, error) {
	if strings.TrimSpace(tokenEndpoint) == "" || strings.TrimSpace(resource) == "" {
		return nil, fmt.Errorf("microwave/auth: token endpoint and resource are required")
	}
	if strings.TrimSpace(subjectToken) == "" {
		return nil, fmt.Errorf("microwave/auth: subject token is required")
	}
	if httpClient == nil {
		httpClient = defaultHTTPClient()
	}
	form := url.Values{
		"grant_type":           {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"subject_token_type":   {"urn:ietf:params:oauth:token-type:jwt"},
		"requested_token_type": {"urn:ietf:params:oauth:token-type:jwt"},
		"subject_token":        {subjectToken},
		"resource":             {resource},
	}
	tok, err := postToken(ctx, httpClient, tokenEndpoint, form)
	if err != nil {
		return nil, err
	}
	return &TokenExchangeResult{
		AccessToken: tok.AccessToken,
		TokenType:   tok.TokenType,
		ExpiresIn:   tok.ExpiresIn,
		Scope:       tok.Scope,
	}, nil
}

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
