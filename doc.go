// Package microwave is the module root for the Microwave Go SDK by Mataki Labs.
//
// Microwave's two server planes have separate Go clients in two subpackages:
//
//   - [github.com/microwave-sh/microwave-go/management] — Management API
//     client. Workspaces, permission sets, signing key sets, key
//     specifications, trust exchanges, trust providers, and trust bindings.
//     Authenticated via a management key or a session JWT obtained through
//     token exchange.
//
//   - [github.com/microwave-sh/microwave-go/auth] — Auth plane client. Redeems
//     an inbound OIDC assertion (Terraform Cloud workload identity, GitHub
//     Actions, an external IdP) for a Microwave session JWT against a
//     configured Trust Exchange. Unauthenticated at the HTTP layer; the
//     assertion is the credential.
//
// Most consumers import only the management subpackage. Federated consumers
// (Terraform providers, CI jobs) import both: auth to get a session, then
// management to do work with that session.
package microwave
