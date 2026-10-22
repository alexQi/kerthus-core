package authorization

import (
	"context"
	"testing"

	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
	"go-micro.dev/v6/client"
	microerrors "go-micro.dev/v6/errors"
	"go-micro.dev/v6/metadata"
)

type grantPlatform struct {
	pb.PlatformService
	revoked bool
	checks  int
}

func (p *grantPlatform) CheckJobAuthorization(ctx context.Context, req *pb.AppAuthorization, _ ...client.CallOption) (*pb.AccessReply, error) {
	p.checks++
	md, _ := metadata.FromContext(ctx)
	if md["X-Kerthus-Service-Key"] == "" || req.AppCode != "fablex" || req.Authorization != "grant" {
		tpanic()
	}
	if p.revoked {
		return nil, microerrors.Forbidden("kerthus.saas", "revoked")
	}
	return &pb.AccessReply{TenantId: 7, UserIds: []int64{11}}, nil
}
func (p *grantPlatform) RevokeJobAuthorization(_ context.Context, req *pb.AppAuthorization, _ ...client.CallOption) (*pb.Result, error) {
	if req.Authorization != "grant" {
		tpanic()
	}
	p.revoked = true
	return &pb.Result{}, nil
}
func tpanic() { panic("unscoped test request") }

func TestJobRevocationIsCheckedOnEveryRequest(t *testing.T) {
	p := &grantPlatform{}
	c, err := New(p, "fablex", "01234567890123456789012345678901")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.CheckJob(context.Background(), "grant"); err != nil {
		t.Fatal(err)
	}
	if err = c.RevokeJob(context.Background(), "grant"); err != nil {
		t.Fatal(err)
	}
	if _, err = c.CheckJob(context.Background(), "grant"); err == nil || p.checks != 2 {
		t.Fatal("revoked job was reused", err)
	}
}
