# microwave-go

Go SDK for the [Microwave Management API](https://api.microwave.sh) by Mataki Labs. Covers the workspace-admin surface: permission sets, signing key sets, key specifications, and trust exchanges.

## Installation

```sh
go get github.com/microwave-sh/microwave-go
```

## Quick start

```go
package main

import (
    "context"
    "log"

    microwave "github.com/microwave-sh/microwave-go"
)

func main() {
    client, err := microwave.NewClient(
        microwave.WithManagementKey("mw_live_..."),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    ps, err := client.PermissionSets.Create(ctx, &microwave.PermissionSetInput{
        Name:        "deployer",
        Description: "Deploy + upload, no destructive ops",
        Permissions: []microwave.PermissionInput{
            {Resource: "deploys", Action: "create"},
            {Resource: "deploys", Action: "activate"},
            {Resource: "blobs", Action: "upload"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Println("created permission set:", ps.ID)
}
```

The management key can also come from `MICROWAVE_MANAGEMENT_KEY` — `WithManagementKey` is omitted in that case.

## Services

The client exposes one service per Management API resource family. Each service has `Create`, `Get`, `Update`, `Delete`, and `List` (signing key sets skip `Update` — algorithm + kind are immutable).

| Field | Resource |
|---|---|
| `client.PermissionSets` | RBAC bundles bound into key specs |
| `client.SigningKeySets` | JWKS-managed signing material (asymmetric or symmetric) |
| `client.KeySpecs` | Key specifications — opaque + JWT formats |
| `client.TrustExchanges` | OIDC federation rules with CEL policy gates |

## Errors

Non-2xx responses produce `*microwave.Error` values. Two helpers cover the common idempotency patterns:

```go
if err := client.PermissionSets.Delete(ctx, "ps_missing"); err != nil && !microwave.IsNotFound(err) {
    log.Fatal(err)
}

if _, err := client.PermissionSets.Create(ctx, in); err != nil && !microwave.IsConflict(err) {
    log.Fatal(err)
}
```

## Federation (`microwave-go/auth`)

For federated consumers — Terraform Cloud runs, GitHub Actions jobs, internal services with workload identity — the `auth` subpackage redeems an inbound OIDC assertion for a Microwave session JWT against a configured Trust Exchange:

```go
import (
    microwave "github.com/microwave-sh/microwave-go"
    "github.com/microwave-sh/microwave-go/auth"
)

authClient, _ := auth.NewClient() // defaults to https://auth.microwave.sh

result, err := authClient.TokenExchange.Redeem(ctx, "ex_tfc_admin", externalOIDCToken)
if err != nil || !result.Valid {
    log.Fatalf("exchange failed: err=%v code=%s rules=%v", err, result.Code, result.RuleResults)
}

mgmt, _ := microwave.NewClient(microwave.WithManagementKey(result.JWT))
```

A denied exchange returns `Valid=false` with a `Code` (and `RuleResults` when the denial came from CEL policy evaluation) — distinct from a transport-level error.

## API version

This SDK pins one Microwave API version (`microwave.APIVersion`). Bumping the SDK is the only way to move to a newer API version — date-versioned headers are not a runtime knob.

## Status

v0.x — surface is stable for the four resources listed above. Future versions add `microwave_trust_provider`, paginated list responses, lookup-by-name data sources, and pipeline-generated SDKs once `microwave-spec` ships.

## License

Apache 2.0 — see [LICENSE](LICENSE).
