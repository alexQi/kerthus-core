package appmanifest

import (
	"strings"
	"testing"
)

func validManifest() Manifest {
	return Manifest{
		SchemaVersion: 1, AppCode: "notes", Name: "Notes", ReleaseVersion: "1.0.0",
		PlatformAPIVersion: "v1", ResourceVersion: 1, HTTPContract: "openapi.yaml", ServiceKey: "kerthus.notes",
		RoutePrefix: "/api/apps/notes/v1", FrontendEntry: "bundled", Home: "/notes/home",
		Resources:  []Resource{{Code: "notes:home", Name: "Home", Type: "view", Path: "/notes/home"}},
		Operations: []Operation{{ID: "notes.read", AppCode: "notes", ResourceCode: "notes:home", Action: "notes.read", Method: "GET", Path: "/api/apps/notes/v1/notes/{id}"}},
	}
}

func TestManifestValidationAndFingerprint(t *testing.T) {
	m := validManifest()
	first, err := m.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	second, err := m.Fingerprint()
	if err != nil || first != second || len(first) != 64 {
		t.Fatalf("unstable fingerprint: %q %q %v", first, second, err)
	}
	const legacyHash = "62ae7d7dcc6f0ccb0fddfcb38e835bd16429cc0d1fa2df6394c6ee06ce4b7588"
	if first != legacyHash {
		t.Fatalf("legacy fingerprint changed: got %s want %s", first, legacyHash)
	}
	if got := m.OperationIDs(); len(got) != 1 || got[0] != "notes.read" {
		t.Fatalf("unexpected operation IDs: %#v", got)
	}
}

func TestFingerprintAlsoCoversPlatformBuiltins(t *testing.T) {
	m := validManifest()
	m.AppCode = "basic"
	m.RoutePrefix = "/system"
	if err := m.Validate(); err == nil {
		t.Fatal("builtin app unexpectedly passed external app validation")
	}
	if hash, err := m.Fingerprint(); err != nil || len(hash) != 64 {
		t.Fatalf("builtin manifest could not be fingerprinted: %q %v", hash, err)
	}
}

func TestManifestRejectsInvalidSharedContract(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Manifest)
	}{
		{"reserved app", func(m *Manifest) { m.AppCode = "saas" }},
		{"route escape", func(m *Manifest) { m.Operations[0].Path = "/api/apps/other/v1/read" }},
		{"resource cycle", func(m *Manifest) { m.Resources[0].ParentCode = m.Resources[0].Code }},
		{"ambiguous routes", func(m *Manifest) {
			m.Operations = append(m.Operations, Operation{ID: "notes.other", AppCode: "notes", ResourceCode: "notes:home", Action: "notes.other", Method: "GET", Path: "/api/apps/notes/v1/notes/{name}"})
		}},
		{"static parameter overlap", func(m *Manifest) {
			m.Operations = append(m.Operations, Operation{ID: "notes.search", AppCode: "notes", ResourceCode: "notes:home", Action: "notes.search", Method: "GET", Path: "/api/apps/notes/v1/notes/search"})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := validManifest()
			tc.mutate(&m)
			if err := m.Validate(); err == nil {
				t.Fatal("invalid manifest accepted")
			}
		})
	}
	if err := (Manifest{SchemaVersion: 1, AppCode: strings.Repeat("a", 63)}).Validate(); err == nil {
		t.Fatal("oversized application identity accepted")
	}
}
