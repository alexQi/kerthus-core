package appmanifest

import "testing"

func TestIndependentFrontendContract(t *testing.T) {
	m := validManifest()
	m.FrontendEntry = "independent"
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	f, err := m.Frontend()
	if err != nil || f.BasePath != "/notes" || f.Home != "/home" {
		t.Fatalf("legacy independent entry: %+v %v", f, err)
	}
	m.FrontendBasePath, m.FrontendHome = "/notes-web/", "/dashboard"
	f, err = m.Frontend()
	if err != nil || f.BasePath != "/notes-web" || f.Home != "/dashboard" {
		t.Fatalf("explicit independent entry: %+v %v", f, err)
	}
	before, _ := m.Fingerprint()
	m.FrontendHome = "/projects"
	after, _ := m.Fingerprint()
	if before == after {
		t.Fatal("frontend route change did not change manifest fingerprint")
	}
	m.FrontendBasePath = "/other-mount"
	next, _ := m.Fingerprint()
	if next == after {
		t.Fatal("frontend mount change did not change manifest fingerprint")
	}
}

func TestIndependentFrontendRejectsUnsafePaths(t *testing.T) {
	for _, path := range []string{"https://elsewhere.test/app", "//elsewhere.test/app", "/app/../admin", "/app/./admin", "/app/%2e%2e/admin", "/app?x=1", "/app#section", "/app\\admin", "/app\nadmin", "/app//admin"} {
		t.Run(path, func(t *testing.T) {
			for _, base := range []bool{true, false} {
				m := validManifest()
				m.FrontendEntry = "independent"
				m.FrontendBasePath, m.FrontendHome = "/notes", "/home"
				if base {
					m.FrontendBasePath = path
				} else {
					m.FrontendHome = path
				}
				if err := m.Validate(); err == nil {
					t.Fatalf("unsafe frontend path accepted (base=%v)", base)
				}
			}
		})
	}
	for _, path := range []string{"/", "/api", "/api/apps", "/ASSETS/app", "/resource/files"} {
		m := validManifest()
		m.FrontendEntry, m.FrontendBasePath = "independent", path
		if err := m.Validate(); err == nil {
			t.Fatalf("reserved frontend base accepted: %s", path)
		}
	}
	m := validManifest()
	m.FrontendBasePath = "/notes"
	if err := m.Validate(); err == nil {
		t.Fatal("bundled frontend accepted an independent mount")
	}
}
