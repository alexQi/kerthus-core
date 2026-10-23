# v1 Compatibility

The `v1` protobuf and public Go contracts use append-only compatibility.

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

The compatibility tests in `rpc/saas/v1/compatibility_test.go` lock the public
method names, model receipt field numbers, diagnostic append-only fields, and
unknown-field decoding behavior. Platform integration tests are responsible for
proving that a deployed server accepts both client versions.
