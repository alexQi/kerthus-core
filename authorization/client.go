package authorization

import (
	"context"
	"fmt"
	"net/http"

	internal "github.com/alexQi/kerthus-core/appkit/authorization"
	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
)

// Client is the phase-A public authorization facade. It translates the
// platform wire reply into the transport-neutral Decision contract so an
// application does not need to depend on generated implementation details.
// The compatibility appkit client remains available for job and media flows
// that have not yet moved to the standalone core module.
type Client struct {
	inner *internal.Client
}

func New(platform pb.PlatformService, appCode, serviceKey string) (*Client, error) {
	inner, err := internal.New(platform, appCode, serviceKey)
	if err != nil {
		return nil, err
	}
	return &Client{inner: inner}, nil
}

func (c *Client) Check(ctx context.Context, input *pb.Context, method, path string) (Decision, error) {
	if c == nil || c.inner == nil {
		return Decision{}, fmt.Errorf("authorization client is nil")
	}
	reply, err := c.inner.Check(ctx, input, method, path)
	if err != nil {
		return Decision{}, err
	}
	return decisionFromReply(reply), nil
}

func (c *Client) CheckRequest(r *http.Request) (Decision, error) {
	if c == nil || c.inner == nil {
		return Decision{}, fmt.Errorf("authorization client is nil")
	}
	reply, err := c.inner.CheckRequest(r)
	if err != nil {
		return Decision{}, err
	}
	return decisionFromReply(reply), nil
}

func decisionFromReply(reply *pb.AccessReply) Decision {
	if reply == nil {
		return Decision{}
	}
	return Decision{
		UserID:             reply.GetUserId(),
		TenantID:           reply.GetTenantId(),
		AppID:              reply.GetAppId(),
		AllWithinTenant:    reply.GetAllWithinTenant(),
		UserIDs:            append([]int64(nil), reply.GetUserIds()...),
		ServiceKey:         reply.GetServiceKey(),
		RoutePrefix:        reply.GetRoutePrefix(),
		OperationID:        reply.GetOperationId(),
		ResourceVersion:    reply.GetResourceVersion(),
		UnitID:             reply.GetUnitId(),
		SectionID:          reply.GetSectionId(),
		RequestID:          reply.GetRequestId(),
		Authorization:      reply.GetAuthorization(),
		ExpiresAt:          reply.GetExpiresAt(),
		PolicyJSON:         reply.GetPolicyJson(),
		ServicePrincipalID: reply.GetServicePrincipalId(),
		GrantedOwnerID:     reply.GetGrantedOwnerId(),
		AuthorizationKind:  reply.GetAuthorizationKind(),
		GrantedObjectID:    reply.GetGrantedObjectId(),
		GrantedPurpose:     reply.GetGrantedPurpose(),
	}
}
