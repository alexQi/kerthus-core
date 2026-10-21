// Package media contains references to platform-managed objects. The object
// itself remains behind the platform media contract; applications own ACLs.
package media

type Scope struct {
	TenantID       int64
	AppID          int64
	ObjectID       string
	Purpose        string
	OwnerID        int64
	MediaID        int64
	ContentVersion string
}

type Reference struct {
	Scope
	Role string
}

type Chunk struct {
	Reference
	Offset int64
	Length int64
	Data   []byte
	Done   bool
}
