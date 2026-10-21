package appidentity

import (
	"strings"
	"testing"
)

func TestDatabaseNameBoundaries(t *testing.T) {
	for code, want := range map[string]string{
		"fablex":      "kerthus_fablex",
		"media-tools": "kerthus_media_htools",
		"media_tools": "kerthus_media_utools",
	} {
		got, err := DatabaseName(code)
		if err != nil || got != want {
			t.Fatalf("DatabaseName(%q) = %q, %v; want %q", code, got, err, want)
		}
	}
	for _, code := range []string{"saas", "system", "basic", strings.Repeat("a", 63)} {
		if got, err := DatabaseName(code); err == nil {
			t.Fatalf("DatabaseName(%q) unexpectedly accepted as %q", code, got)
		}
	}
	left, _ := DatabaseName("media-tools")
	right, _ := DatabaseName("media_tools")
	if left == right {
		t.Fatalf("distinct application codes collide on %q", left)
	}
	exact, err := DatabaseName(strings.Repeat("a", 56))
	if err != nil || len(exact) != 64 {
		t.Fatalf("64-byte database name boundary rejected: %q, %v", exact, err)
	}
	if got, err := DatabaseName(strings.Repeat("a", 57)); err == nil {
		t.Fatalf("65-byte database name accepted as %q", got)
	}
}
