package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/alexQi/kerthus-core/appkit/rpcauth"
	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
	"go-micro.dev/v6/client"
	microerrors "go-micro.dev/v6/errors"
	"google.golang.org/protobuf/proto"
	"time"
)

// ModelCallRPCTimeout covers the platform's ten-minute provider window and
// leaves time for receipt persistence and the RPC response.
const ModelCallRPCTimeout = 10*time.Minute + 30*time.Second

var ErrCallNotFound = errors.New("platform model call not found")

// Client is for application workers. Supply a job authorization, never a login token.
type Client struct {
	platform  pb.PlatformService
	code, key string
}

func New(p pb.PlatformService, code, key string) (*Client, error) {
	if p == nil || code == "" || len(key) < 32 {
		return nil, fmt.Errorf("application model identity missing")
	}
	return &Client{p, code, key}, nil
}
func (c *Client) auth(token string) *pb.AppAuthorization {
	return &pb.AppAuthorization{AppCode: c.code, Authorization: token}
}
func (c *Client) Catalog(ctx context.Context, authorization string) (*pb.ModelCatalogReply, error) {
	return c.platform.ListApplicationModels(rpcauth.WithCredential(ctx, c.key), c.auth(authorization))
}
func (c *Client) Submit(ctx context.Context, jobAuthorization, selector, key, prompt, size string, seconds int32) (*pb.ModelCallReply, error) {
	return c.SubmitWithInputs(ctx, jobAuthorization, selector, key, prompt, size, seconds, nil)
}
func (c *Client) SubmitWithInputs(ctx context.Context, jobAuthorization, selector, key, prompt, size string, seconds int32, inputs []*pb.ModelInputRef) (*pb.ModelCallReply, error) {
	frozen := make([]*pb.ModelInputRef, 0, len(inputs))
	for _, ref := range inputs {
		if ref == nil {
			return nil, fmt.Errorf("nil model input")
		}
		copy := proto.Clone(ref).(*pb.ModelInputRef)
		copy.Authorization = ""
		frozen = append(frozen, copy)
	}

	return admissionResult(c.platform.SubmitApplicationModel(rpcauth.WithCredential(ctx, c.key), &pb.ModelSubmitRequest{Auth: c.auth(jobAuthorization), Selector: selector, IdempotencyKey: key, Prompt: prompt, Size: size, Seconds: seconds, Inputs: frozen}, client.WithRequestTimeout(ModelCallRPCTimeout)))
}
func (c *Client) Query(ctx context.Context, jobAuthorization string, id int64) (*pb.ModelCallReply, error) {
	return c.platform.QueryApplicationModel(rpcauth.WithCredential(ctx, c.key), &pb.ModelCallRequest{Auth: c.auth(jobAuthorization), CallId: id}, client.WithRequestTimeout(ModelCallRPCTimeout))
}
func (c *Client) Cancel(ctx context.Context, jobAuthorization string, id int64) (*pb.ModelCallReply, error) {
	return c.platform.CancelApplicationModel(rpcauth.WithCredential(ctx, c.key), &pb.ModelCallRequest{Auth: c.auth(jobAuthorization), CallId: id})
}

// Lookup only reads the platform ledger using the exact frozen submission. It does not retry it.
func (c *Client) Lookup(ctx context.Context, jobAuthorization string, frozen *pb.ModelSubmitRequest) (*pb.ModelCallReply, error) {
	if frozen == nil {
		return nil, fmt.Errorf("model request missing")
	}
	req := proto.Clone(frozen).(*pb.ModelSubmitRequest)
	req.Auth = c.auth(jobAuthorization)
	for _, ref := range req.Inputs {
		if ref == nil {
			return nil, fmt.Errorf("nil model input")
		}
		ref.Authorization = ""
	}
	reply, err := c.platform.LookupApplicationModel(rpcauth.WithCredential(ctx, c.key), &pb.ModelLookupRequest{Request: req})
	if err != nil {
		rpcErr := microerrors.FromError(err)
		if rpcErr.Id == "kerthus.saas" && rpcErr.Code == 404 {
			return nil, ErrCallNotFound
		}
	}
	return reply, err
}

// SubmitRequest supports research batches; credentials and media permits never come from the caller.
func (c *Client) SubmitRequest(ctx context.Context, jobAuthorization string, frozen *pb.ModelSubmitRequest) (*pb.ModelCallReply, error) {
	if frozen == nil {
		return nil, fmt.Errorf("model request missing")
	}
	if frozen.SchedulerVersion != 0 {
		if frozen.SchedulerVersion != 1 || frozen.RootRunId == "" {
			return nil, ErrSchedulerUnsupported
		}
		// Replays only need an authenticated ledger lookup, not the entire
		// catalog. No capability/permission cache survives between requests.
		existing, lookupErr := c.Lookup(ctx, jobAuthorization, frozen)
		if lookupErr == nil {
			if existing == nil || existing.SchedulerVersion != 1 || existing.RootRunId != frozen.RootRunId {
				return nil, ErrSchedulerUnsupported
			}
			return existing, nil
		}
		if !errors.Is(lookupErr, ErrCallNotFound) {
			return nil, lookupErr
		}
		catalog, err := c.Catalog(ctx, jobAuthorization)
		if err != nil {
			return nil, err
		}
		supported := false
		for _, item := range catalog.Models {
			if item.Selector == frozen.Selector && item.SchedulerVersion == 1 {
				supported = true
				break
			}
		}
		if !supported {
			return nil, ErrSchedulerUnsupported
		}
	}
	req := proto.Clone(frozen).(*pb.ModelSubmitRequest)
	req.Auth = c.auth(jobAuthorization)
	for _, ref := range req.Inputs {
		if ref == nil {
			return nil, fmt.Errorf("nil model input")
		}
		ref.Authorization = ""
	}
	return admissionResult(c.platform.SubmitApplicationModel(rpcauth.WithCredential(ctx, c.key), req, client.WithRequestTimeout(ModelCallRPCTimeout)))
}

// ControlRun is safe to replay from an application outbox with the same sequence.
func (c *Client) ControlRun(ctx context.Context, authorization, rootRunID, state string, seq int64) (*pb.ModelRunReply, error) {
	return c.platform.ControlApplicationModelRun(rpcauth.WithCredential(ctx, c.key), &pb.ModelRunControlRequest{Auth: c.auth(authorization), RootRunId: rootRunID, State: state, ControlSeq: seq})
}
func (c *Client) QueryRun(ctx context.Context, authorization, rootRunID string) (*pb.ModelRunReply, error) {
	return c.platform.QueryApplicationModelRun(rpcauth.WithCredential(ctx, c.key), &pb.ModelRunRequest{Auth: c.auth(authorization), RootRunId: rootRunID})
}
func (c *Client) Rebind(ctx context.Context, authorization string, callID int64) (*pb.ModelCallReply, error) {
	return c.platform.RebindApplicationModel(rpcauth.WithCredential(ctx, c.key), &pb.ModelRebindRequest{Auth: c.auth(authorization), CallId: callID})
}

func (c *Client) SetModelExecutionControl(ctx context.Context, authorization, rootRunID, state string, seq int64) (*pb.ModelRunReply, error) {
	return c.ControlRun(ctx, authorization, rootRunID, state, seq)
}
