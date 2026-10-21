package authorization

import (
	"context"
	"fmt"

	"github.com/alexQi/kerthus-core/appkit/rpcauth"
	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
	"google.golang.org/protobuf/proto"
)

// AuthorizeHistoricalReceipt must load the original call, frozen request, object
// and source relationships from application storage and check current business
// ACL. Never trust scope, call IDs or source lists supplied by browser JSON.
type AuthorizeHistoricalReceipt func(context.Context, *pb.AccessReply, *pb.HistoricalModelReceiptRequest) error

// ReadHistoricalReceipt only reads a stored ledger snapshot using a current
// checked HTTP request. Input/source permits must be fresh. It cannot authorize
// polling, retry, execution continuation or repair a missing historical grant.
func (c *Client) ReadHistoricalReceipt(ctx context.Context, access *pb.AccessReply, req *pb.HistoricalModelReceiptRequest, acl AuthorizeHistoricalReceipt) (*pb.HistoricalModelReceiptReply, error) {
	if access == nil || access.Authorization == "" || req == nil || acl == nil {
		return nil, fmt.Errorf("current request, frozen receipt request and object/source ACL callback required")
	}
	frozen := proto.Clone(req).(*pb.HistoricalModelReceiptRequest)
	frozen.Auth = c.Authorization(access)
	if err := acl(ctx, access, proto.Clone(frozen).(*pb.HistoricalModelReceiptRequest)); err != nil {
		return nil, err
	}
	return c.platform.ReadHistoricalModelReceipt(rpcauth.WithCredential(ctx, c.serviceKey), frozen)
}
