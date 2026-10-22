package media

import (
	"context"
	"errors"
	"testing"

	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
	"go-micro.dev/v6/client"
)

type readPlatform struct {
	pb.PlatformService
	reads   int
	revoked bool
}

func (p *readPlatform) ReadApplicationMedia(_ context.Context, req *pb.MediaReadRequest, _ ...client.CallOption) (*pb.MediaChunk, error) {
	p.reads++
	if p.revoked {
		return nil, errors.New("platform media grant revoked")
	}
	if req.Media.Auth.AppCode != "fablex" || req.Media.ObjectId != "object-1" || req.Media.OwnerId != 11 || req.Media.Purpose != "source" {
		return nil, errors.New("media scope lost")
	}
	return &pb.MediaChunk{Data: []byte("ok")}, nil
}

func TestMediaRechecksObjectACLAndPlatformGrant(t *testing.T) {
	p := &readPlatform{}
	allowed := true
	checks := 0
	c, err := New(p, "fablex", "01234567890123456789012345678901", func(_ context.Context, _ *pb.AccessReply, scope Scope) error {
		checks++
		if scope.Action == "read" && !allowed {
			return errors.New("object revoked")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	object, err := c.Bind(context.Background(), &pb.AccessReply{Authorization: "grant", TenantId: 7}, Scope{ObjectID: "object-1", Purpose: "source", OwnerID: 11})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = object.Read(context.Background(), 42, 0, 2, false); err != nil {
		t.Fatal(err)
	}
	allowed = false
	if _, err = object.Read(context.Background(), 42, 2, 2, false); err == nil || p.reads != 1 {
		t.Fatal("application ACL bypassed", err)
	}
	allowed = true
	p.revoked = true
	if _, err = object.Read(context.Background(), 42, 2, 2, false); err == nil || p.reads != 2 || checks != 4 {
		t.Fatal("platform revocation bypassed", err)
	}
}
