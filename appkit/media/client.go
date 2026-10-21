// Package media requires the application to authorize a business object before accessing platform media.
package media

import (
	"context"
	"fmt"
	"github.com/alexQi/kerthus-core/appcontract"
	"github.com/alexQi/kerthus-core/appkit/rpcauth"
	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
	"io"
	"mime"
	"net/http"
	"time"
)

type Scope struct {
	Action            string
	ObjectID, Purpose string
	OwnerID           int64
}
type AuthorizeObject func(context.Context, *pb.AccessReply, Scope) error
type Client struct {
	platform  pb.PlatformService
	code, key string
	acl       AuthorizeObject
}

func New(platform pb.PlatformService, code, key string, acl AuthorizeObject) (*Client, error) {
	if platform == nil || code == "" || len(key) < 32 || acl == nil {
		return nil, fmt.Errorf("media requires app identity and application object authorization callback")
	}
	return &Client{platform, code, key, acl}, nil
}

// Bind is called only with verified access and a scope loaded from application storage, not directly from browser JSON.
func (c *Client) Bind(ctx context.Context, access *pb.AccessReply, scope Scope) (*Object, error) {
	if access == nil || access.Authorization == "" || scope.OwnerID <= 0 || scope.ObjectID == "" || scope.Purpose == "" {
		return nil, fmt.Errorf("verified object context missing")
	}
	if e := c.acl(ctx, access, scope); e != nil {
		return nil, e
	}
	return &Object{client: c, scope: scope, access: access, auth: &pb.AppAuthorization{AppCode: c.code, Authorization: access.Authorization}}, nil
}

type Object struct {
	access *pb.AccessReply
	client *Client
	scope  Scope
	auth   *pb.AppAuthorization
}

func (o *Object) authorize(ctx context.Context, action string) error {
	scope := o.scope
	scope.Action = action
	return o.client.acl(ctx, o.access, scope)
}
func (o *Object) ctx(ctx context.Context) context.Context {
	return rpcauth.WithCredential(ctx, o.client.key)
}
func (o *Object) request(id int64) *pb.MediaRequest {
	return &pb.MediaRequest{Auth: o.auth, MediaId: id, ObjectId: o.scope.ObjectID, Purpose: o.scope.Purpose, OwnerId: o.scope.OwnerID}
}
func (o *Object) Begin(ctx context.Context, filename, contentType, key string, size int64) (*pb.MediaReply, error) {
	if e := o.authorize(ctx, "upload"); e != nil {
		return nil, e
	}

	return o.client.platform.BeginApplicationMedia(o.ctx(ctx), &pb.MediaBeginRequest{Auth: o.auth, ObjectId: o.scope.ObjectID, Purpose: o.scope.Purpose, OwnerId: o.scope.OwnerID, Filename: filename, ContentType: contentType, Size: size, IdempotencyKey: key})
}
func (o *Object) Part(ctx context.Context, id, number int64, data []byte) (*pb.MediaReply, error) {
	if e := o.authorize(ctx, "upload"); e != nil {
		return nil, e
	}

	return o.client.platform.WriteApplicationMediaPart(o.ctx(ctx), &pb.MediaPartRequest{Media: o.request(id), Number: number, Data: data, Digest: appcontract.Digest(data)})
}
func (o *Object) Confirm(ctx context.Context, id int64) (*pb.MediaReply, error) {
	if e := o.authorize(ctx, "upload"); e != nil {
		return nil, e
	}

	return o.client.platform.ConfirmApplicationMedia(o.ctx(ctx), o.request(id))
}
func (o *Object) Delete(ctx context.Context, id int64) error {
	if e := o.authorize(ctx, "delete"); e != nil {
		return e
	}

	_, e := o.client.platform.DeleteApplicationMedia(o.ctx(ctx), o.request(id))
	return e
}
func (o *Object) Playback(ctx context.Context, id int64) (*pb.PlaybackReply, error) {
	if e := o.authorize(ctx, "read"); e != nil {
		return nil, e
	}

	return o.client.platform.CreateApplicationPlayback(o.ctx(ctx), o.request(id))
}
func (o *Object) Read(ctx context.Context, id, offset, length int64, metadata bool) (*pb.MediaChunk, error) {
	if e := o.authorize(ctx, "read"); e != nil {
		return nil, e
	}

	return o.client.platform.ReadApplicationMedia(o.ctx(ctx), &pb.MediaReadRequest{Media: o.request(id), Offset: offset, Length: length, MetadataOnly: metadata})
}
func (o *Object) ServeHTTP(w http.ResponseWriter, r *http.Request, id int64) {
	ServeContent(w, r, func(offset, length int64, metadata bool) (*pb.MediaChunk, error) {
		return o.Read(r.Context(), id, offset, length, metadata)
	})
}

// ServeContent supports byte ranges and HEAD while holding at most one 1 MiB chunk in memory.
// Every chunk is separately authorized by the platform. An in-flight chunk is the revocation window.
func ServeContent(w http.ResponseWriter, r *http.Request, read func(int64, int64, bool) (*pb.MediaChunk, error)) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "method not allowed", 405)
		return
	}
	meta, e := read(0, 0, true)
	if e != nil || meta == nil || meta.Media == nil {
		http.Error(w, "media unavailable", 403)
		return
	}
	info := meta.Media
	w.Header().Set("Content-Type", info.ContentType)
	if info.ContentType == "application/zip" || r.URL.Query().Get("download") == "1" {
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": info.Filename}))
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("ETag", fmt.Sprintf(`"media-%d-%d"`, info.MediaId, info.Size))
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(30 * time.Minute))
	http.ServeContent(w, r, info.Filename, time.Time{}, &reader{size: info.Size, fetch: read})
}

type reader struct {
	size, offset, cacheStart int64
	cache                    []byte
	fetch                    func(int64, int64, bool) (*pb.MediaChunk, error)
}

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		offset += r.offset
	case io.SeekEnd:
		offset += r.size
	default:
		return 0, fmt.Errorf("invalid seek")
	}
	if offset < 0 {
		return 0, fmt.Errorf("invalid offset")
	}
	r.offset = offset
	return offset, nil
}
func (r *reader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.offset >= r.size {
		return 0, io.EOF
	}
	if r.cache == nil || r.offset < r.cacheStart || r.offset >= r.cacheStart+int64(len(r.cache)) {
		v, e := r.fetch(r.offset, 1<<20, false)
		if e != nil {
			return 0, e
		}
		if v == nil || len(v.Data) == 0 {
			return 0, io.ErrUnexpectedEOF
		}
		r.cache = v.Data
		r.cacheStart = r.offset
	}
	n := copy(p, r.cache[r.offset-r.cacheStart:])
	r.offset += int64(n)
	return n, nil
}

func (o *Object) AuthorizeModelInput(ctx context.Context, id int64, version, role string) (*pb.ModelInputRef, error) {
	if e := o.authorize(ctx, "reference"); e != nil {
		return nil, e
	}
	return o.client.platform.AuthorizeApplicationModelInput(o.ctx(ctx), &pb.ModelInputAuthorizationRequest{Media: o.request(id), ContentVersion: version, Role: role})
}
func (o *Object) RevokeAccess(ctx context.Context, id int64) error {
	if e := o.authorize(ctx, "revoke"); e != nil {
		return e
	}
	_, e := o.client.platform.RevokeApplicationMediaAccess(o.ctx(ctx), o.request(id))
	return e
}
