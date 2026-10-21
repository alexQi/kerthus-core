// Package authorization contains transport-neutral authorization contracts.
// Applications still enforce their own object and business ACL after the
// platform returns a verified decision.
package authorization

import "fmt"

type Context struct {
	Token           string
	TenantID        int64
	AppID           int64
	UnitID          int64
	SectionID       int64
	RequestID       string
	DelegatedUserID int64
}

type Request struct {
	Context Context
	AppCode string
	Method  string
	Path    string
}

type Decision struct {
	UserID             int64
	TenantID           int64
	AppID              int64
	AllWithinTenant    bool
	UserIDs            []int64
	ServiceKey         string
	RoutePrefix        string
	OperationID        string
	ResourceVersion    int64
	UnitID             int64
	SectionID          int64
	RequestID          string
	Authorization      string
	ExpiresAt          int64
	PolicyJSON         string
	ServicePrincipalID string
	GrantedOwnerID     int64
	AuthorizationKind  string
	GrantedObjectID    string
	GrantedPurpose     string
}

func (r Request) Validate() error {
	if r.AppCode == "" || r.Method == "" || r.Path == "" {
		return fmt.Errorf("application, method, and path are required")
	}
	if r.Context.TenantID < 0 || r.Context.AppID < 0 || r.Context.UnitID < 0 || r.Context.SectionID < 0 {
		return fmt.Errorf("authorization context is invalid")
	}
	return nil
}

// AllowsOwner applies the platform scope to an application-owned row. The
// tenant predicate is always mandatory, including for tenant-wide decisions.
func AllowsOwner(d Decision, tenantID, ownerID int64) bool {
	if d.TenantID <= 0 || tenantID <= 0 || ownerID <= 0 || d.TenantID != tenantID {
		return false
	}
	if d.AllWithinTenant {
		return true
	}
	for _, id := range d.UserIDs {
		if id == ownerID {
			return true
		}
	}
	return false
}
