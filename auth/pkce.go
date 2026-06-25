package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// pkcePair is an RFC 7636 code_verifier and its S256 code_challenge.
type pkcePair struct {
	verifier  string
	challenge string
}

// newPKCE generates a high-entropy RFC 7636 verifier (32 random bytes,
// base64url-encoded → 43 chars) and its S256 challenge. S256 is the only
// method the brokered AS accepts.
func newPKCE() (pkcePair, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return pkcePair{}, fmt.Errorf("microwave/auth: generate pkce verifier: %w", err)
	}
	verifier := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(verifier))
	return pkcePair{
		verifier:  verifier,
		challenge: base64.RawURLEncoding.EncodeToString(sum[:]),
	}, nil
}

// randomURLToken returns a URL-safe random string of n bytes' entropy, used
// for the OAuth `state` parameter and the device-flow nonce.
func randomURLToken(n int) (string, error) {
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("microwave/auth: generate random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
