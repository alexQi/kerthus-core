// Package context defines the platform context shared by independently
// deployed applications. It contains identity and scope data only; it does
// not encode application business rules.
package context

import "fmt"

// App identifies the platform application that owns a request.
type App struct {
	ID   int64
	Code string
}

// Scope is the smallest portable context an application may persist beside a
// job or object reference. Credentials and provider secrets are deliberately
// excluded.
type Scope struct {
	TenantID  int64
	App       App
	UserID    int64
	OwnerID   int64
	RequestID string
	JobID     string
	ObjectID  string
	Purpose   string
}

func (s Scope) Validate() error {
	if s.TenantID <= 0 || s.App.ID <= 0 || s.App.Code == "" {
		return fmt.Errorf("tenant and application identity are required")
	}
	if s.UserID < 0 || s.OwnerID < 0 {
		return fmt.Errorf("user and owner IDs cannot be negative")
	}
	if s.ObjectID == "" && s.JobID == "" && s.RequestID == "" {
		return fmt.Errorf("request, job, or object scope is required")
	}
	return nil
}
