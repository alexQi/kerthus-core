# kerthus-core

`github.com/alexQi/kerthus-core` is the public, provider-neutral Go module for Kerthus applications. It contains tenant/application scope contracts, authorization decisions, versioned RPC DTOs, model invocation/receipt contracts, media references, operation policy validation, and application database naming/DSN checks.

The module deliberately does not contain platform stores, Provider implementations, FableX domain code, migrations, or business ACLs. Applications keep their own business database and enforce object ownership after platform authorization.

## Versioning

The v1 API is the generated wire contract under `gen/go/saas/v1`, exposed through `rpc/saas/v1`. Applications should import the public packages listed below:

- `context`, `authorization`, `models`, `media`
- `appcontract`, `appidentity`, `appkit/database`
- `appkit/authorization`, `appkit/models`, `appkit/media`
- `appmanifest`
- `rpc/saas/v1`

For model work, prefer `appkit/models.NewStable`. Its `StableClient` maps
transport replies to `models.Receipt`, including the provider-neutral
`models.Diagnostic`, and exposes submit/query/reconcile/cancel without requiring
application code to inspect protobuf diagnostics. The older
`appkit/models.Client` remains a compatibility wire client for scheduler and
migration code; it is a low-level escape hatch, not a second execution or
receipt implementation.

Application-facing RPC access is capability-scoped by the platform service key
and app code. The v1 generated service still contains legacy management methods
for wire compatibility; applications must use only the documented application
allowlist (authorization, job/model, media, health and operation contracts),
and must never call tenant or platform administration methods.

The protobuf field numbers and service method names are compatibility commitments for v1. The platform remains responsible for Provider selection, model execution, receipt persistence, retries, quotas, and key custody.

The v1 compatibility rule is append-only: a server accepts application clients
from the current and previous v1 patch contract, and a current application
accepts responses from the previous patch contract with newly added fields
absent. Unknown protobuf fields are ignored. The service credential header and
its binding format remain unchanged. `appmanifest` contains only shared DTOs,
validation, and fingerprints; filesystem loading, OpenAPI parsing, and catalog
conversion stay in the platform.

See [docs/compatibility.md](docs/compatibility.md) for the version matrix and
wire-contract test scope.
