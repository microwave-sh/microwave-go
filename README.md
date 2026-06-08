# microwave-go

Go SDK for the [Microwave Management API](https://api.microwave.sh) by Mataki Labs. Covers the workspace-admin surface: permission sets, signing key sets, key specifications, and trust exchanges.

> **Looking for the public utility surface?** Dish (225+ stateless utility endpoints at `dish.microwave.sh` — address parsing, barcode generation, country lookup, etc) lives in the separate sibling module [`microwave-sh/microwave-dish-go`](https://github.com/microwave-sh/microwave-dish-go). This module is for AKaaS workspace administration only.

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

## API version

This SDK pins one Microwave API version (`microwave.APIVersion`). Bumping the SDK is the only way to move to a newer API version — date-versioned headers are not a runtime knob.

## Status

v0.x — surface is stable for the four resources listed above. Future versions add `microwave_trust_provider`, paginated list responses, lookup-by-name data sources, and pipeline-generated SDKs once `microwave-spec` ships.

## License

Apache 2.0 — see [LICENSE](LICENSE).
