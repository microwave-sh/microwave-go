# microwave-go

Go SDK for [Microwave](https://microwave.sh) by Mataki Labs. Two subpackages mirror the two server planes:

- [`management`](./management) — Management API client. Workspaces, permission sets, signing key sets, key specifications, trust exchanges, trust providers, trust federations, and trust federation bindings. Authenticated via a management key or a session JWT obtained through token exchange.
- [`auth`](./auth) — Auth plane client. Redeems an inbound OIDC assertion (Terraform Cloud workload identity, GitHub Actions, an external IdP) for a Microwave session JWT against a configured Trust Exchange.

Most consumers import only `management`. Federated consumers (Terraform providers, CI jobs) import both: `auth` to obtain a session, then `management` to do work with it.

## Installation

```sh
go get github.com/microwave-sh/microwave-go@latest
```

## Quick start — Management API

```go
package main

import (
    "context"
    "log"

    "github.com/microwave-sh/microwave-go/management"
)

func main() {
    client, err := management.NewClient(
        management.WithManagementKey("mw_live_..."),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    ps, err := client.PermissionSets.Create(ctx, &management.PermissionSetInput{
        Name:        "deployer",
        Description: "Deploy + upload, no destructive ops",
        Permissions: []management.PermissionInput{
            {Name: "deploys:write", Label: "Write deploys"},
            {Name: "deploys:read", Label: "Read deploys"},
            {Name: "sites:read", Label: "Read sites"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Println("created permission set:", ps.ID)
}
```

The management key can also come from `MICROWAVE_MANAGEMENT_KEY` — `WithManagementKey` is omitted in that case.

### Services

Each service has `Create`, `Get`, `Update`, `Delete`, and `List` (signing key sets skip `Update` — algorithm + kind are immutable).

| Field | Resource |
|---|---|
| `client.PermissionSets` | RBAC bundles bound into key specs |
| `client.SigningKeySets` | JWKS-managed signing material (asymmetric or symmetric) |
| `client.KeySpecs` | Key specifications — opaque + JWT formats |
| `client.TrustExchanges` | OIDC federation rules with CEL policy gates |
| `client.TrustProviders` | Microwave minting surfaces for downstream consumers |
| `client.TrustFederations` | OIDC-authenticated trust federations; includes `Redeem` for federation token exchange |
| `client.TrustFederationBindings` | Identity tuple bindings scoped to a trust federation |

### Errors

Non-2xx responses produce `*management.Error` values. Two helpers cover the common idempotency patterns:

```go
if err := client.PermissionSets.Delete(ctx, "ps_missing"); err != nil && !management.IsNotFound(err) {
    log.Fatal(err)
}

if _, err := client.PermissionSets.Create(ctx, in); err != nil && !management.IsConflict(err) {
    log.Fatal(err)
}
```

## Federation — `auth` subpackage

For federated consumers — Terraform Cloud runs, GitHub Actions jobs, internal services with workload identity — `auth.TokenExchange.Redeem` exchanges an inbound OIDC assertion for a Microwave session JWT against a configured Trust Exchange:

```go
import (
    "github.com/microwave-sh/microwave-go/auth"
    "github.com/microwave-sh/microwave-go/management"
)

authClient, _ := auth.NewClient() // defaults to https://auth.microwave.sh

result, err := authClient.TokenExchange.Redeem(ctx, "ex_tfc_admin", externalOIDCToken)
if err != nil || !result.Valid {
    log.Fatalf("exchange failed: err=%v code=%s rules=%v", err, result.Code, result.RuleResults)
}

mgmt, _ := management.NewClient(management.WithManagementKey(result.JWT))
```

A denied exchange returns `Valid=false` with a `Code` (and `RuleResults` when the denial came from CEL policy evaluation) — distinct from a transport-level error.

## API version

The `management` subpackage pins one Microwave Management API version (`management.APIVersion`). Bumping the module is the only way to move to a newer API version — date-versioned headers are not a runtime knob.

## Status

v0.x — surface is stabilizing for the resources listed above. Future versions add paginated list responses, lookup-by-name discovery, and pipeline-generated SDKs once `microwave-spec` ships.

## License

Apache 2.0 — see [LICENSE](LICENSE).
