# Migrating an application to kerthus-core v1

1. Add `github.com/alexQi/kerthus-core v1.0.1` (or the latest compatible v1 patch) to the application module.
2. Replace platform contract imports with the public packages in this module. Do not import `kerthus/internal/...` or generated types from a Kerthus checkout.
3. Use `appidentity.DatabaseName(appCode)` and `appkit/database.ValidateDSN` before opening the application database. The application database must be `kerthus_<encoded app code>` and must not be the platform database.
4. Use `authorization.Client.Check` or `CheckRequest` for current platform scope, then apply the application's own object ACL. Never accept tenant, actor, owner, or upstream URL values from model input as authority.
5. Use `appkit/models.NewStable` with a job authorization for model calls. Persist the idempotency key and reconcile `unknown` or `waiting_receipt` results before retrying. Do not implement a second Provider, receipt, or retry ledger. Use the raw `appkit/models.Client` only when a versioned scheduler operation has no stable facade yet.
6. Use `appkit/media` with an application ACL callback. Media references must retain tenant/app/object/purpose/owner scope; do not expose platform storage tables.

The v1 module has no FableX business types. Creative, research, scripts, production workflows, HTTP handlers, business migrations, and the `kerthus_fablex` schema remain in the application repository.
