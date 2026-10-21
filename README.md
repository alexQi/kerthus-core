# kerthus-core

`github.com/alexQi/kerthus-core` is the public, provider-neutral Go module for Kerthus applications. It contains tenant/application scope contracts, authorization decisions, versioned RPC DTOs, model invocation/receipt contracts, media references, operation policy validation, and application database naming/DSN checks.

The module deliberately does not contain platform stores, Provider implementations, FableX domain code, migrations, or business ACLs. Applications keep their own business database and enforce object ownership after platform authorization.

## Versioning

The v1 API is the generated wire contract under `gen/go/saas/v1`, exposed through `rpc/saas/v1`. Applications should import the public packages listed below:

- `context`, `authorization`, `models`, `media`
- `appcontract`, `appidentity`, `appkit/database`
- `appkit/authorization`, `appkit/models`, `appkit/media`
- `rpc/saas/v1`

The protobuf field numbers and service method names are compatibility commitments for v1. The platform remains responsible for Provider selection, model execution, receipt persistence, retries, quotas, and key custody.
