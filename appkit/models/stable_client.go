package models

import (
	"context"
	"fmt"

	"github.com/alexQi/kerthus-core/appkit/rpcauth"
	contractmodels "github.com/alexQi/kerthus-core/models"
	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
	"go-micro.dev/v6/client"
	"google.golang.org/protobuf/proto"
)

// StableClient is the application-facing model client. It converts the
// transport reply into the versioned models.Receipt contract so applications
// do not depend on protobuf field names or platform diagnostics.
//
// Client remains available as the v1 wire-compatible client for advanced
// scheduler and migration use cases. New application code should prefer this
// facade and persist only the stable receipt fields it needs.
type StableClient struct{ raw *Client }

func NewStable(p pb.PlatformService, code, key string) (*StableClient, error) {
	raw, err := New(p, code, key)
	if err != nil {
		return nil, err
	}
	return &StableClient{raw: raw}, nil
}

func (c *StableClient) Submit(ctx context.Context, invocation contractmodels.Invocation) (contractmodels.Receipt, error) {
	if err := invocation.Validate(); err != nil {
		return contractmodels.Receipt{}, err
	}
	req := &pb.ModelSubmitRequest{
		Selector:        invocation.Selector,
		IdempotencyKey:  invocation.IdempotencyKey,
		Prompt:          invocation.Prompt,
		Size:            invocation.Size,
		Seconds:         invocation.Seconds,
		Texts:           append([]string(nil), invocation.Texts...),
		MaxOutputTokens: invocation.MaxOutputTokens,
		RootRunId:       invocation.RootRunID,
		Inputs:          make([]*pb.ModelInputRef, 0, len(invocation.Inputs)),
	}
	for _, input := range invocation.Inputs {
		req.Inputs = append(req.Inputs, &pb.ModelInputRef{MediaId: input.MediaID, ContentVersion: input.ContentVersion, Role: input.Role})
	}
	reply, err := c.raw.submitFrozen(ctx, invocation.Authorization, req)
	return receiptFromReply(reply, &invocation), err
}

func (c *StableClient) Query(ctx context.Context, request contractmodels.QueryRequest) (contractmodels.Receipt, error) {
	if err := request.Validate(); err != nil {
		return contractmodels.Receipt{}, err
	}
	reply, err := c.raw.Query(ctx, request.Authorization, request.CallID)
	return receiptFromReply(reply, nil), err
}

// Reconcile reads the existing ledger entry by call ID. It never submits a
// second provider request and is safe to use after an unknown response.
func (c *StableClient) Reconcile(ctx context.Context, request contractmodels.ReconcileRequest) (contractmodels.Receipt, error) {
	if err := request.Validate(); err != nil {
		return contractmodels.Receipt{}, err
	}
	reply, err := c.raw.Rebind(ctx, request.Authorization, request.CallID)
	invocation := &contractmodels.Invocation{IdempotencyKey: request.IdempotencyKey, Selector: request.Selector, RootRunID: request.RootRunID}
	return receiptFromReply(reply, invocation), err
}

func (c *StableClient) Cancel(ctx context.Context, request contractmodels.QueryRequest) (contractmodels.Receipt, error) {
	if err := request.Validate(); err != nil {
		return contractmodels.Receipt{}, err
	}
	reply, err := c.raw.Cancel(ctx, request.Authorization, request.CallID)
	return receiptFromReply(reply, nil), err
}

func (c *Client) submitFrozen(ctx context.Context, jobAuthorization string, frozen *pb.ModelSubmitRequest) (*pb.ModelCallReply, error) {
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
	return admissionResult(c.platform.SubmitApplicationModel(rpcauth.WithCredential(ctx, c.key), req, client.WithRequestTimeout(ModelCallRPCTimeout)))
}

func receiptFromReply(reply *pb.ModelCallReply, invocation *contractmodels.Invocation) contractmodels.Receipt {
	if reply == nil {
		return contractmodels.Receipt{}
	}
	receipt := contractmodels.Receipt{
		CallID:     reply.GetCallId(),
		State:      contractmodels.NormalizeState(reply.GetState()),
		RequestID:  reply.GetRequestId(),
		Message:    reply.GetMessage(),
		ResultJSON: reply.GetResultJson(),
		UsageJSON:  reply.GetUsageJson(),
		QueuePhase: reply.GetQueuePhase(),
		RootRunID:  reply.GetRootRunId(),
	}
	if invocation != nil {
		receipt.IdempotencyKey = invocation.IdempotencyKey
		receipt.Attempt = invocation.Attempt
		receipt.Provider = invocation.Provider
		receipt.Model = invocation.Model
		if receipt.RootRunID == "" {
			receipt.RootRunID = invocation.RootRunID
		}
	}
	if diagnostic := reply.GetDiagnostic(); diagnostic != nil {
		receipt.Diagnostic = &contractmodels.Diagnostic{
			Code: diagnostic.GetCode(), Phase: diagnostic.GetPhase(), HTTPStatus: diagnostic.GetHttpStatus(),
			UpstreamRequestID: diagnostic.GetUpstreamRequestId(), Submission: diagnostic.GetSubmission(),
			Message: diagnostic.GetMessage(), RetrySafe: diagnostic.GetRetrySafe(),
			UpstreamStatusCode: diagnostic.GetUpstreamStatusCode(), UpstreamLogID: diagnostic.GetUpstreamLogId(),
			RequestBytes: diagnostic.GetRequestBytes(), RequestSHA256: diagnostic.GetRequestSha256(),
			RouteSHA256: diagnostic.GetRouteSha256(), Protocol: diagnostic.GetProtocol(), Model: diagnostic.GetModel(),
			StartedAtUnixMS: diagnostic.GetStartedAtUnixMs(), FinishedAtUnixMS: diagnostic.GetFinishedAtUnixMs(),
			DNSMS: diagnostic.GetDnsMs(), ConnectMS: diagnostic.GetConnectMs(), TLSMS: diagnostic.GetTlsMs(),
			WriteMS: diagnostic.GetWriteMs(), FirstByteMS: diagnostic.GetFirstByteMs(), ReadMS: diagnostic.GetReadMs(),
			HTTPVersion: diagnostic.GetHttpVersion(), ALPN: diagnostic.GetAlpn(), RemoteAddr: diagnostic.GetRemoteAddr(),
			UpstreamErrorCode: diagnostic.GetUpstreamErrorCode(), UpstreamErrorType: diagnostic.GetUpstreamErrorType(),
			UpstreamMessage: diagnostic.GetUpstreamMessage(), RetryClass: diagnostic.GetRetryClass(),
			RetryAfterSeconds: diagnostic.GetRetryAfterSeconds(), RetryPolicy: diagnostic.GetRetryPolicy(),
			ExecutionOutcome: diagnostic.GetExecutionOutcome(),
		}
		receipt.ErrorClass = diagnostic.GetCode()
		if receipt.Message == "" {
			receipt.Message = diagnostic.GetMessage()
		}
		if receipt.Model == "" {
			receipt.Model = diagnostic.GetModel()
		}
	}
	if receipt.ErrorClass == "" && receipt.State == contractmodels.StateFailed {
		receipt.ErrorClass = reply.GetRecovery()
	}
	return receipt
}
