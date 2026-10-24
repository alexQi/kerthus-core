// Package appmanifest contains the public, provider-neutral application
// manifest contract. File loading and platform catalog conversion belong to
// the Kerthus platform and are intentionally outside this package.
package appmanifest

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/alexQi/kerthus-core/appcontract"
	"github.com/alexQi/kerthus-core/appidentity"
)

type Manifest struct {
	SchemaVersion      int         `json:"schema_version" yaml:"schema_version"`
	AppCode            string      `json:"app_code" yaml:"app_code"`
	Name               string      `json:"name" yaml:"name"`
	ReleaseVersion     string      `json:"release_version" yaml:"release_version"`
	PlatformAPIVersion string      `json:"platform_api_version" yaml:"platform_api_version"`
	ResourceVersion    int64       `json:"resource_version" yaml:"resource_version"`
	HTTPContract       string      `json:"http_contract" yaml:"http_contract"`
	ServiceKey         string      `json:"service_key" yaml:"service_key"`
	RoutePrefix        string      `json:"route_prefix" yaml:"route_prefix"`
	FrontendEntry      string      `json:"frontend_entry" yaml:"frontend_entry"`
	FrontendBasePath   string      `json:"frontend_base_path,omitempty" yaml:"frontend_base_path,omitempty"`
	FrontendHome       string      `json:"frontend_home,omitempty" yaml:"frontend_home,omitempty"`
	Home               string      `json:"home" yaml:"home"`
	Resources          []Resource  `json:"resources" yaml:"resources"`
	Operations         []Operation `json:"operations,omitempty" yaml:"-"`
}

type Resource struct {
	OpenWith   string         `json:"open_with,omitempty" yaml:"open_with,omitempty"`
	Redirect   string         `json:"redirect,omitempty" yaml:"redirect,omitempty"`
	Meta       map[string]any `json:"meta,omitempty" yaml:"meta,omitempty"`
	Code       string         `json:"code" yaml:"code"`
	ParentCode string         `json:"parent_code" yaml:"parent_code"`
	Name       string         `json:"name" yaml:"name"`
	Type       string         `json:"type" yaml:"type"`
	Path       string         `json:"path" yaml:"path"`
	Component  string         `json:"component" yaml:"component"`
	Icon       string         `json:"icon" yaml:"icon"`
	Sort       int32          `json:"sort" yaml:"sort"`
	DataScope  bool           `json:"data_scope" yaml:"data_scope"`
}

type Operation struct {
	Summary     string   `json:"-" yaml:"summary"`
	Description string   `json:"-" yaml:"description"`
	Tags        []string `json:"-" yaml:"tags"`
	Parameters  []any    `json:"-" yaml:"parameters"`
	RequestBody any      `json:"-" yaml:"requestBody"`
	Responses   any      `json:"-" yaml:"responses"`
	Security    any      `json:"-" yaml:"security"`
	Deprecated  bool     `json:"-" yaml:"deprecated"`

	Policy       *appcontract.Policy `json:"policy,omitempty" yaml:"x-kerthus"`
	ID           string              `json:"operation_id" yaml:"operationId"`
	AppCode      string              `json:"app_code" yaml:"x-app-code"`
	ResourceCode string              `json:"resource_code" yaml:"x-resource-code"`
	Action       string              `json:"action" yaml:"x-action"`
	Group        string              `json:"group" yaml:"x-group"`
	Public       bool                `json:"public" yaml:"x-public"`
	Method       string              `json:"method" yaml:"-"`
	Path         string              `json:"path" yaml:"-"`
}

type Spec struct {
	Info       any                             `json:"info,omitempty" yaml:"info"`
	Servers    any                             `json:"servers,omitempty" yaml:"servers"`
	Components any                             `json:"components,omitempty" yaml:"components"`
	Security   any                             `json:"security,omitempty" yaml:"security"`
	Tags       any                             `json:"tags,omitempty" yaml:"tags"`
	OpenAPI    string                          `json:"openapi" yaml:"openapi"`
	Paths      map[string]map[string]Operation `json:"paths" yaml:"paths"`
}

var paramSegment = regexp.MustCompile(`^\{[a-zA-Z][a-zA-Z0-9_]*\}$`)

// Validate checks the manifest fields shared by registration and application
// tooling. Filesystem, OpenAPI, and component checks remain in the platform.
func (m Manifest) Validate() error {
	if m.SchemaVersion != 1 || m.PlatformAPIVersion != "v1" || m.ResourceVersion < 1 {
		return fmt.Errorf("unsupported manifest version")
	}
	if !appidentity.ValidCode(m.AppCode) || m.Name == "" || m.ReleaseVersion == "" || m.HTTPContract == "" {
		return fmt.Errorf("invalid application identity or release metadata")
	}
	if _, err := appidentity.DatabaseName(m.AppCode); err != nil {
		return fmt.Errorf("invalid application database identity: %w", err)
	}
	if m.RoutePrefix != "/api/apps/"+m.AppCode+"/v1" || m.ServiceKey == "" {
		return fmt.Errorf("route prefix or service key is invalid")
	}
	if _, err := m.Frontend(); err != nil {
		return err
	}
	if !strings.HasPrefix(m.Home, "/"+m.AppCode+"/") {
		return fmt.Errorf("frontend entry or home is invalid")
	}
	resources := make(map[string]Resource, len(m.Resources))
	for _, r := range m.Resources {
		if r.Code == "" || r.Name == "" || !strings.HasPrefix(r.Code, m.AppCode+":") {
			return fmt.Errorf("resource identity is invalid")
		}
		if _, exists := resources[r.Code]; exists {
			return fmt.Errorf("duplicate resource %s", r.Code)
		}
		switch r.Type {
		case "menu", "view", "action", "field":
		default:
			return fmt.Errorf("invalid resource type %s", r.Type)
		}
		switch r.OpenWith {
		case "", "component", "route", "inside", "outside":
		default:
			return fmt.Errorf("unsupported resource open_with")
		}
		resources[r.Code] = r
	}
	for _, r := range m.Resources {
		seen := map[string]bool{r.Code: true}
		for parent := r.ParentCode; parent != ""; {
			if seen[parent] {
				return fmt.Errorf("resource cycle")
			}
			seen[parent] = true
			p, ok := resources[parent]
			if !ok {
				return fmt.Errorf("missing parent resource %s", parent)
			}
			parent = p.ParentCode
		}
	}
	ids := make(map[string]bool, len(m.Operations))
	routes := make(map[string]bool, len(m.Operations))
	for i, op := range m.Operations {
		if op.ID == "" || op.AppCode != m.AppCode || op.ResourceCode == "" || op.Action == "" {
			return fmt.Errorf("operation metadata is incomplete")
		}
		resource, resourceOK := resources[op.ResourceCode]
		if ids[op.ID] || !resourceOK || resource.Code == "" {
			return fmt.Errorf("operation identity or resource is invalid")
		}
		switch strings.ToUpper(op.Method) {
		case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE":
		default:
			return fmt.Errorf("operation method is invalid")
		}
		if op.Path == "" || !strings.HasPrefix(op.Path, m.RoutePrefix+"/") {
			return fmt.Errorf("operation route is invalid")
		}
		if strings.ContainsAny(op.Path, "?#%") || strings.Contains(op.Path, "//") {
			return fmt.Errorf("operation route is invalid")
		}
		parts := strings.Split(op.Path, "/")
		for i, part := range parts {
			if part == "." || part == ".." || (strings.ContainsAny(part, "{}") && !paramSegment.MatchString(part)) {
				return fmt.Errorf("operation path segment is invalid")
			}
			if paramSegment.MatchString(part) {
				parts[i] = "{}"
			}
		}
		key := strings.ToUpper(op.Method) + " " + strings.Join(parts, "/")
		if routes[key] {
			return fmt.Errorf("duplicate or ambiguous operation route")
		}
		for _, previous := range m.Operations[:i] {
			if strings.ToUpper(previous.Method) == strings.ToUpper(op.Method) && routesOverlap(previous.Path, op.Path) {
				return fmt.Errorf("ambiguous application routes")
			}
		}
		if op.Policy != nil {
			if err := op.Policy.Validate(); err != nil {
				return fmt.Errorf("operation %s policy: %w", op.ID, err)
			}
		}
		ids[op.ID], routes[key] = true, true
	}
	return nil
}

func routesOverlap(leftPath, rightPath string) bool {
	left, right := strings.Split(leftPath, "/"), strings.Split(rightPath, "/")
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] && !paramSegment.MatchString(left[i]) && !paramSegment.MatchString(right[i]) {
			return false
		}
	}
	return true
}

// Fingerprint returns the schema-v1 SHA-256 identity of a manifest. Call
// Validate or the platform loader before using it for registration decisions.
func (m Manifest) Fingerprint() (string, error) {
	// The platform registered schema-v1 manifests with Go field names in the
	// fingerprint JSON. Keep that byte representation even though the public
	// DTO has snake_case JSON tags for interchange.
	legacy := struct {
		SchemaVersion      int
		AppCode            string
		Name               string
		ReleaseVersion     string
		PlatformAPIVersion string
		ResourceVersion    int64
		HTTPContract       string
		ServiceKey         string
		RoutePrefix        string
		FrontendEntry      string
		Home               string
		Resources          []fingerprintResource
		Operations         []Operation
		FrontendBasePath   string `json:",omitempty"`
		FrontendHome       string `json:",omitempty"`
	}{
		m.SchemaVersion, m.AppCode, m.Name, m.ReleaseVersion,
		m.PlatformAPIVersion, m.ResourceVersion, m.HTTPContract,
		m.ServiceKey, m.RoutePrefix, m.FrontendEntry, m.Home,
		nil, m.Operations, m.FrontendBasePath, m.FrontendHome,
	}
	if m.Resources != nil {
		legacy.Resources = make([]fingerprintResource, 0, len(m.Resources))
		for _, resource := range m.Resources {
			legacy.Resources = append(legacy.Resources, fingerprintResource{
				OpenWith: resource.OpenWith, Redirect: resource.Redirect,
				Meta: resource.Meta, Code: resource.Code, ParentCode: resource.ParentCode,
				Name: resource.Name, Type: resource.Type, Path: resource.Path,
				Component: resource.Component, Icon: resource.Icon, Sort: resource.Sort,
				DataScope: resource.DataScope,
			})
		}
	}
	b, err := json.Marshal(legacy)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

type fingerprintResource struct {
	OpenWith   string `json:",omitempty"`
	Redirect   string `json:",omitempty"`
	Meta       map[string]any
	Code       string
	ParentCode string
	Name       string
	Type       string
	Path       string
	Component  string
	Icon       string
	Sort       int32
	DataScope  bool
}

// OperationIDs returns deterministic operation ordering for catalog tooling.
func (m Manifest) OperationIDs() []string {
	ids := make([]string, 0, len(m.Operations))
	for _, op := range m.Operations {
		ids = append(ids, op.ID)
	}
	sort.Strings(ids)
	return ids
}
