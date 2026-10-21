// Package rpcauth provides the client-side service credential binding used by
// public appkit RPC clients. The credential is only attached to RPC metadata;
// it is never part of application objects or model/media DTOs.
package rpcauth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"go-micro.dev/v6/metadata"
)

const CredentialHeader = "X-Kerthus-Service-Key"

func WithCredential(ctx context.Context, key string) context.Context {
	return metadata.Set(ctx, CredentialHeader, key)
}

// PrincipalID derives a stable non-secret identity for an application service
// credential. The platform still validates the credential before authorizing a
// request; this value is safe to register in an operation policy.
func PrincipalID(serviceKey, appCode string) string {
	d := sha256.Sum256([]byte("kerthus/service-principal/v1\x00" + serviceKey))
	return hex.EncodeToString(d[:])
}
