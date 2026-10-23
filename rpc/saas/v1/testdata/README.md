# v1.0.3 protobuf baseline

`v1.0.3.protoset` is the descriptor set for
`api/proto/saas/v1/core.proto` at the published `v1.0.3` tag. The source file
has SHA-256 `fd46a9694c7bddc53be82998808a295ff91592ef6f9a19eb54aca3c990b0a885`.

It was generated with:

```sh
protoc -I api/proto \
  --descriptor_set_out=rpc/saas/v1/testdata/v1.0.3.protoset \
  api/proto/saas/v1/core.proto
```

The compatibility test checks every published message field, enum value, and
RPC method against this baseline. New v1 fields and methods may be appended.
