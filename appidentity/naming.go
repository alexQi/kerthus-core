// Package appidentity defines stable application identifiers shared by registration and appkit.
package appidentity

import (
	"fmt"
	"regexp"
	"strings"
)

const databasePrefix = "kerthus_"
const mysqlIdentifierLimit = 64

var codePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,62}$`)

var reservedCodes = map[string]struct{}{
	"basic":  {},
	"saas":   {},
	"system": {},
}

func ValidCode(code string) bool {
	return codePattern.MatchString(code)
}

// DatabaseName is injective for valid application codes: '_' and '-' have distinct escapes.
func DatabaseName(code string) (string, error) {
	if !ValidCode(code) {
		return "", fmt.Errorf("invalid application code")
	}
	if _, reserved := reservedCodes[code]; reserved {
		return "", fmt.Errorf("application code %s is reserved by the platform", code)
	}
	var name strings.Builder
	name.Grow(len(databasePrefix) + len(code))
	name.WriteString(databasePrefix)
	for _, r := range code {
		switch r {
		case '_':
			name.WriteString("_u")
		case '-':
			name.WriteString("_h")
		default:
			name.WriteRune(r)
		}
	}
	if name.Len() > mysqlIdentifierLimit {
		return "", fmt.Errorf("application database name exceeds MySQL's %d-byte identifier limit", mysqlIdentifierLimit)
	}
	return name.String(), nil
}
