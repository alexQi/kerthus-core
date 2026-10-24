# v1 Compatibility

The `v1` protobuf and public Go contracts use append-only compatibility.

Independent frontend metadata adds `frontend_entry` (10), `frontend_base_path`
(11), and `frontend_home` (12) to `LoginReply`, which is also the authorization
context. `App` adds `frontend_base_path` (19) and `frontend_home` (20). The
existing `home` and `frontend_entry` field numbers and RPC methods are unchanged.
Older clients can continue to read their existing response fields.

Schema-v1 Manifests accept `bundled` and `independent`. Independent entries may
declare `frontend_base_path` and `frontend_home`; `Manifest.Frontend()` validates
the same-origin paths and resolves omitted values from the app code and legacy
`home`. The legacy `home` remains the permission resource path. Independent
launch fields participate in the fingerprint when supplied; omitted fields
preserve every existing bundled manifest fingerprint.

| Participant | Required behavior |
| --- | --- |
| Server N | Accept application clients using N and the previous v1 patch contract. |
| Application N | Decode responses from server N and the previous v1 patch contract when newly added fields are absent. |
| Both sides | Ignore unknown protobuf fields and preserve existing field numbers and RPC method names. |
| RPC authentication | Keep `X-Kerthus-Service-Key` and its service-principal binding unchanged. |

Adding a field uses a new field number. Existing fields and methods are never
removed or repurposed within v1. A client must not infer that an absent field
means a negative result; it uses the documented default and stable state
contract instead.

The `v1.0.3` descriptor fixture and tests in
`rpc/saas/v1/compatibility_test.go` lock every previously published message
field, enum value, and RPC method, as well as unknown-field decoding. The
service credential header has its own regression test. Platform integration
tests remain responsible for proving that a deployed server accepts both client
versions.
