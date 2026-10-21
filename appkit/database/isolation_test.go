package database

import "testing"

func TestValidateDSN(t *testing.T) {
	for name, test := range map[string]struct {
		dsn  string
		code string
		ok   bool
	}{
		"own database":               {"app:secret@tcp(127.0.0.1:3306)/kerthus_fablex?parseTime=true", "fablex", true},
		"hyphen escaped":             {"app:secret@tcp(127.0.0.1:3306)/kerthus_media_htools", "media-tools", true},
		"underscore escaped":         {"app:secret@tcp(127.0.0.1:3306)/kerthus_media_utools", "media_tools", true},
		"old ambiguous name":         {"app:secret@tcp(127.0.0.1:3306)/kerthus_media_tools", "media-tools", false},
		"platform code and database": {"app:secret@tcp(127.0.0.1:3306)/kerthus_saas", "saas", false},
		"platform database":          {"app:secret@tcp(127.0.0.1:3306)/kerthus_saas", "fablex", false},
		"another app":                {"app:secret@tcp(127.0.0.1:3306)/kerthus_notes", "fablex", false},
		"missing database":           {"app:secret@tcp(127.0.0.1:3306)/", "fablex", false},
		"invalid code":               {"app:secret@tcp(127.0.0.1:3306)/kerthus_FableX", "FableX", false},
		"database name too long":     {"app:secret@tcp(127.0.0.1:3306)/kerthus_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},
	} {
		t.Run(name, func(t *testing.T) {
			err := ValidateDSN(test.dsn, test.code)
			if (err == nil) != test.ok {
				t.Fatalf("ValidateDSN() error = %v, want success %v", err, test.ok)
			}
		})
	}
}
