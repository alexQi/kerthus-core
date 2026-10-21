// Package authorization is the application-side platform client. Applications
// must call Check before using a tenant ID or scope to query their own data.
package authorization

import (
	"bytes"
	"context"
	"fmt"
	"github.com/alexQi/kerthus-core/appcontract"
	"github.com/alexQi/kerthus-core/appkit/rpcauth"
	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
	"go-micro.dev/v6/metadata"
	"io"
	"net/http"
)

type Client struct {
	platform            pb.PlatformService
	appCode, serviceKey string
}

func New(platform pb.PlatformService, appCode, serviceKey string) (*Client, error) {
	if platform == nil || appCode == "" || len(serviceKey) < 32 {
		return nil, fmt.Errorf("platform, bound application code and service credential are required")
	}
	return &Client{platform, appCode, serviceKey}, nil
}
func (c *Client) Check(ctx context.Context, input *pb.Context, method, path string) (*pb.AccessReply, error) {
	ctx = metadata.Set(ctx, "X-Kerthus-Service-Key", c.serviceKey)
	return c.platform.CheckAccess(ctx, &pb.AccessRequest{Context: input, AppCode: c.appCode, Method: method, Path: path})
}
func (c *Client) CheckRequest(r *http.Request) (*pb.AccessReply, error) {
	if r == nil || r.URL == nil {
		return nil, fmt.Errorf("missing HTTP request")
	}
	if r.Header.Get(appcontract.TicketHeader) == "" {
		return nil, fmt.Errorf("missing platform delegation")
	}
	if r.Body == nil {
		r.Body = http.NoBody
	}
	body, e := io.ReadAll(io.LimitReader(r.Body, appcontract.MaxRequestBytes+1))
	if e != nil {
		return nil, e
	}
	r.Body.Close()
	if int64(len(body)) > appcontract.MaxRequestBytes {
		return nil, fmt.Errorf("application request too large")
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	ctx := metadata.Set(r.Context(), "X-Kerthus-Service-Key", c.serviceKey)
	return c.platform.CheckDelegation(ctx, &pb.DelegationCheck{AppCode: c.appCode, Ticket: r.Header.Get(appcontract.TicketHeader), Method: r.Method, Path: r.URL.Path, Fingerprint: appcontract.Fingerprint(r, body), RequestId: r.Header.Get(appcontract.RequestIDHeader)})
}

// AllowsOwner never treats an empty scope as unrestricted. The tenant predicate
// remains mandatory even for AllWithinTenant (including platform operators).
func AllowsOwner(scope *pb.AccessReply, tenantID, ownerID int64) bool {
	if scope == nil || tenantID <= 0 || ownerID <= 0 || scope.TenantId != tenantID {
		return false
	}
	if scope.AllWithinTenant {
		return true
	}
	for _, id := range scope.UserIds {
		if id == ownerID {
			return true
		}
	}
	return false
}

func (c *Client) Authorization(access *pb.AccessReply) *pb.AppAuthorization {
	return &pb.AppAuthorization{AppCode: c.appCode, Authorization: access.GetAuthorization()}
}
func (c *Client) CreateJob(ctx context.Context, access *pb.AccessReply, object, purpose string, owner, ttl int64) (*pb.DelegationReply, error) {
	return c.CreateJobWithInputs(ctx, access, object, purpose, owner, ttl, nil)
}
func (c *Client) CreateJobWithInputs(ctx context.Context, access *pb.AccessReply, object, purpose string, owner, ttl int64, inputs []*pb.ModelInputRef) (*pb.DelegationReply, error) {
	return c.platform.CreateJobAuthorization(metadata.Set(ctx, "X-Kerthus-Service-Key", c.serviceKey), &pb.JobAuthorizationRequest{Auth: c.Authorization(access), ObjectId: object, Purpose: purpose, OwnerId: owner, TtlSeconds: ttl, Inputs: inputs})
}

// ServicePrincipalID is the non-secret identity to register in the operation
// policy for this application worker.
func (c *Client) ServicePrincipalID(ctx context.Context) (string, error) {
	return c.platformServicePrincipal(ctx)
}

func (c *Client) platformServicePrincipal(ctx context.Context) (string, error) {
	// The platform derives this from the bound service credential; callers never
	// need to handle or persist the credential itself.
	return rpcauth.PrincipalID(c.serviceKey, c.appCode), nil
}

func (c *Client) CreateBackgroundJob(ctx context.Context, tenantID int64, principalID, object, purpose string, owner, mediaID int64, contentVersion, method, path string, ttl int64, idempotencyKey string) (*pb.DelegationReply, error) {
	return c.platform.CreateBackgroundJobAuthorization(rpcauth.WithCredential(ctx, c.serviceKey), &pb.BackgroundJobAuthorizationRequest{AppCode: c.appCode, PrincipalId: principalID, TenantId: tenantID, ObjectId: object, Purpose: purpose, OwnerId: owner, MediaId: mediaID, ContentVersion: contentVersion, Method: method, Path: path, TtlSeconds: ttl, IdempotencyKey: idempotencyKey})
}

// RenewBackgroundJob rotates a live service_job credential in place. The
// returned ticket keeps the same platform grant ID and therefore preserves
// model-call receipts and producer media provenance.
func (c *Client) RenewBackgroundJob(ctx context.Context, grantID int64, authorization, principalID string, tenantID int64, object, purpose string, owner, mediaID int64, contentVersion, method, path string, ttl int64, idempotencyKey string) (*pb.DelegationReply, error) {
	return c.platform.RenewBackgroundJobAuthorization(metadata.Set(ctx, "X-Kerthus-Service-Key", c.serviceKey), &pb.RenewBackgroundJobAuthorizationRequest{AppCode: c.appCode, PrincipalId: principalID, GrantId: grantID, Authorization: authorization, TenantId: tenantID, ObjectId: object, Purpose: purpose, OwnerId: owner, MediaId: mediaID, ContentVersion: contentVersion, Method: method, Path: path, TtlSeconds: ttl, IdempotencyKey: idempotencyKey})
}

func (c *Client) AuthorizeBackgroundModelInput(ctx context.Context, job string, media *pb.MediaRequest, contentVersion, role string) (*pb.ModelInputRef, error) {
	return c.platform.AuthorizeBackgroundModelInput(rpcauth.WithCredential(ctx, c.serviceKey), &pb.BackgroundModelInputAuthorizationRequest{Auth: &pb.AppAuthorization{AppCode: c.appCode, Authorization: job}, Media: media, ContentVersion: contentVersion, Role: role})
}
func (c *Client) CheckJob(ctx context.Context, authorization string) (*pb.AccessReply, error) {
	result, e := c.platform.CheckJobAuthorization(metadata.Set(ctx, "X-Kerthus-Service-Key", c.serviceKey), &pb.AppAuthorization{AppCode: c.appCode, Authorization: authorization})
	if e == nil && result != nil {
		result.Authorization = authorization
	}
	return result, e
}
func (c *Client) RevokeJob(ctx context.Context, authorization string) error {
	_, e := c.platform.RevokeJobAuthorization(metadata.Set(ctx, "X-Kerthus-Service-Key", c.serviceKey), &pb.AppAuthorization{AppCode: c.appCode, Authorization: authorization})
	return e
}

// DeriveJob freezes media uploaded by this parent job. The platform inherits the scope.
func (c *Client) DeriveJob(ctx context.Context, parent string, inputs []*pb.ModelInputRef) (*pb.DelegationReply, error) {
	return c.platform.DeriveJobAuthorization(metadata.Set(ctx, "X-Kerthus-Service-Key", c.serviceKey), &pb.JobDerivationRequest{Auth: &pb.AppAuthorization{AppCode: c.appCode, Authorization: parent}, Inputs: inputs})
}

// DeriveJobForQueue freezes exactly these media inputs for a bounded queue wait.
// The server caps expiry at the live parent's expiry; this never renews a parent.
func (c *Client) DeriveJobForQueue(ctx context.Context, parent string, inputs []*pb.ModelInputRef, expiresInSeconds int64) (*pb.DelegationReply, error) {
	if expiresInSeconds < 1 || expiresInSeconds > 86400 {
		return nil, fmt.Errorf("queue derivation lifetime must be between 1 and 86400 seconds")
	}
	return c.platform.DeriveJobAuthorization(metadata.Set(ctx, "X-Kerthus-Service-Key", c.serviceKey), &pb.JobDerivationRequest{Auth: &pb.AppAuthorization{AppCode: c.appCode, Authorization: parent}, Inputs: inputs, ExpiresInSeconds: expiresInSeconds})
}

// PromoteJobSource creates a source-only root analysis job from one ready
// media object produced by a source-less producer job. Scope is inherited from
// the producer authorization; only the frozen media versions are supplied.
func (c *Client) PromoteJobSource(ctx context.Context, producer string, mediaID int64, contentVersion string, accessVersion, ttl int64, idempotencyKey string) (*pb.DelegationReply, error) {
	return c.platform.PromoteJobSource(metadata.Set(ctx, "X-Kerthus-Service-Key", c.serviceKey), &pb.PromoteJobSourceRequest{Auth: &pb.AppAuthorization{AppCode: c.appCode, Authorization: producer}, MediaId: mediaID, ContentVersion: contentVersion, AccessVersion: accessVersion, TtlSeconds: ttl, IdempotencyKey: idempotencyKey})
}

// ResumeJob requires a fresh checked HTTP user request. Persist the original job
// token as a recovery handle and the returned token as the active credential.
// Application object ACL and cancellation checks must precede this call.
func (c *Client) ResumeJob(ctx context.Context, access *pb.AccessReply, originalJob, object, purpose string, owner, ttl int64) (*pb.DelegationReply, error) {
	return c.platform.ResumeJobAuthorization(metadata.Set(ctx, "X-Kerthus-Service-Key", c.serviceKey), &pb.JobResumeRequest{Auth: c.Authorization(access), JobAuthorization: originalJob, ObjectId: object, Purpose: purpose, OwnerId: owner, TtlSeconds: ttl})
}

// ReadJobReceipt never renews the job or contacts the model provider. Request
// must be the exact frozen original submission, including key and queue version.
func (c *Client) ReadJobReceipt(ctx context.Context, access *pb.AccessReply, originalJob string, frozen *pb.ModelSubmitRequest) (*pb.ModelCallReply, error) {
	return c.platform.ReadJobModelReceipt(metadata.Set(ctx, "X-Kerthus-Service-Key", c.serviceKey), &pb.JobReceiptRequest{Auth: c.Authorization(access), JobAuthorization: originalJob, Request: frozen})
}
