package rpcauth

import "testing"

func TestV1CredentialHeaderRemainsStable(t *testing.T) {
	if CredentialHeader != "X-Kerthus-Service-Key" {
		t.Fatalf("service credential header changed: %q", CredentialHeader)
	}
	if PrincipalID("service-key", "fablex") != PrincipalID("service-key", "other") {
		t.Fatal("service principal derivation unexpectedly depends on app code")
	}
}
