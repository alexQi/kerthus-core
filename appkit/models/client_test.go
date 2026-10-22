package models

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
	"go-micro.dev/v6/client"
	microerrors "go-micro.dev/v6/errors"
)

type replayPlatform struct {
	pb.PlatformService
	lookupErr         error
	found             *pb.ModelCallReply
	version           int32
	catalogs, submits int
	timeout           time.Duration
}

func (p *replayPlatform) LookupApplicationModel(context.Context, *pb.ModelLookupRequest, ...client.CallOption) (*pb.ModelCallReply, error) {
	return p.found, p.lookupErr
}
func (p *replayPlatform) ListApplicationModels(context.Context, *pb.AppAuthorization, ...client.CallOption) (*pb.ModelCatalogReply, error) {
	p.catalogs++
	return &pb.ModelCatalogReply{Models: []*pb.ModelCatalogItem{{Selector: "text", SchedulerVersion: p.version}}}, nil
}
func (p *replayPlatform) SubmitApplicationModel(_ context.Context, req *pb.ModelSubmitRequest, opts ...client.CallOption) (*pb.ModelCallReply, error) {
	p.submits++
	var options client.CallOptions
	for _, option := range opts {
		option(&options)
	}
	p.timeout = options.RequestTimeout
	if req.Auth == nil || req.Auth.AppCode != "fablex" || req.Auth.Authorization != "grant" {
		return nil, errors.New("submission lost job scope")
	}
	return &pb.ModelCallReply{CallId: 1, SchedulerVersion: 1, RootRunId: "root"}, nil
}

func TestSchedulerReplayOnlySubmitsAfterConfirmedAbsence(t *testing.T) {
	p := &replayPlatform{found: &pb.ModelCallReply{CallId: 1, SchedulerVersion: 1, RootRunId: "root"}}
	c, err := New(p, "fablex", "01234567890123456789012345678901")
	if err != nil {
		t.Fatal(err)
	}
	req := &pb.ModelSubmitRequest{SchedulerVersion: 1, RootRunId: "root", Selector: "text"}
	for i := 0; i < 2; i++ {
		if _, err := c.SubmitRequest(context.Background(), "grant", req); err != nil {
			t.Fatal(err)
		}
	}
	if p.catalogs != 0 || p.submits != 0 {
		t.Fatal("replay submitted an existing call")
	}
	for _, failure := range []error{microerrors.Forbidden("kerthus.saas", "revoked"), context.DeadlineExceeded, microerrors.NotFound("transport", "missing route")} {
		p.lookupErr = failure
		if _, err := c.SubmitRequest(context.Background(), "grant", req); err == nil {
			t.Fatal("lookup failure accepted", failure)
		}
	}
	if p.submits != 0 {
		t.Fatal("dependency failure resubmitted")
	}
	p.lookupErr = microerrors.NotFound("kerthus.saas", "missing")
	if _, err := c.Lookup(context.Background(), "grant", req); !errors.Is(err, ErrCallNotFound) {
		t.Fatal(err)
	}
	p.version = 1
	if _, err := c.SubmitRequest(context.Background(), "grant", req); err != nil {
		t.Fatal(err)
	}
	if p.submits != 1 || p.timeout != ModelCallRPCTimeout {
		t.Fatal(p.submits, p.timeout)
	}
	p.version = 0
	if _, err := c.SubmitRequest(context.Background(), "grant", req); !errors.Is(err, ErrSchedulerUnsupported) {
		t.Fatal(err)
	}
}

func TestAdmissionRequiresExplicitNotSent(t *testing.T) {
	legacy := microerrors.New("kerthus.saas", "当前应用模型的每日调用额度已用完", 429)
	_, err := admissionResult(nil, legacy)
	var admission *AdmissionError
	if !errors.As(err, &admission) || !admission.NotSent || !errors.Is(err, ErrDailyQuota) {
		t.Fatal(err)
	}
	_, err = admissionResult(nil, microerrors.New("kerthus.model.admission", `{"code":"model_queue_full","not_sent":false}`, 429))
	if errors.As(err, &admission) {
		t.Fatal("uncertain admission was treated as safe")
	}
	_, err = admissionResult(nil, microerrors.New("kerthus.model.admission", `{"code":"model_queue_full","not_sent":true,"retry_after_seconds":5}`, 429))
	if !errors.As(err, &admission) || admission.Code != "model_queue_full" || admission.RetryAfter != 5*time.Second {
		t.Fatal(err)
	}
}
