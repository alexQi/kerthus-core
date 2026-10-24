package appmanifest

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/alexQi/kerthus-core/appidentity"
)

// Frontend describes a same-origin application mount and the route within it.
// The manifest's Home remains the application's authorization resource path.
type Frontend struct {
	Entry    string
	BasePath string
	Home     string
}

// Frontend validates and resolves launch metadata. Legacy independent entries
// derive their mount from AppCode and their frontend route from Home.
func (m Manifest) Frontend() (Frontend, error) {
	f := Frontend{Entry: m.FrontendEntry}
	switch m.FrontendEntry {
	case "bundled":
		if m.FrontendBasePath != "" || m.FrontendHome != "" {
			return f, fmt.Errorf("bundled frontends cannot declare an independent mount")
		}
		return f, nil
	case "independent":
	default:
		return f, fmt.Errorf("frontend entry must be bundled or independent")
	}
	if !appidentity.ValidCode(m.AppCode) {
		return f, fmt.Errorf("invalid frontend application code")
	}
	f.BasePath = m.FrontendBasePath
	if f.BasePath == "" {
		f.BasePath = "/" + m.AppCode
	}
	if !frontendPath(f.BasePath) || len(f.BasePath) > 191 {
		return f, fmt.Errorf("invalid frontend base path")
	}
	f.BasePath = strings.TrimRight(f.BasePath, "/")
	if f.BasePath == "" {
		return f, fmt.Errorf("frontend base path cannot be root")
	}
	first := strings.ToLower(strings.Split(strings.TrimPrefix(f.BasePath, "/"), "/")[0])
	switch first {
	case "api", "assets", "resource":
		return f, fmt.Errorf("frontend base path is reserved by the platform")
	}
	f.Home = m.FrontendHome
	if f.Home == "" {
		prefix := "/" + m.AppCode
		if m.Home != prefix && !strings.HasPrefix(m.Home, prefix+"/") {
			return f, fmt.Errorf("frontend home is required when legacy home is outside the application")
		}
		f.Home = strings.TrimPrefix(m.Home, prefix)
		if f.Home == "" {
			f.Home = "/"
		}
	}
	if !frontendPath(f.Home) || len(f.Home) > 1024 {
		return f, fmt.Errorf("invalid frontend home")
	}
	return f, nil
}

func frontendPath(value string) bool {
	if !strings.HasPrefix(value, "/") || strings.Contains(value, "//") || strings.ContainsAny(value, "?#%\\") || strings.ContainsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
		return false
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "." || segment == ".." {
			return false
		}
	}
	return true
}
