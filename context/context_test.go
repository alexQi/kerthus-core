package context

import "testing"

func TestScopeRequiresTenantAppAndScopeReference(t *testing.T) {
	s := Scope{TenantID: 1, App: App{ID: 2, Code: "fablex"}, RequestID: "req-1"}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	s.RequestID = ""
	if err := s.Validate(); err == nil {
		t.Fatal("scope without request, job, or object was accepted")
	}
}
