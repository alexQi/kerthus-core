package authorization

import (
	"testing"

	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
)

func TestAllowsOwnerAlwaysChecksTenant(t *testing.T) {
	d := Decision{TenantID: 7, AllWithinTenant: true}
	if !AllowsOwner(d, 7, 11) {
		t.Fatal("tenant-wide scope did not allow owner in its tenant")
	}
	if AllowsOwner(d, 8, 11) {
		t.Fatal("tenant-wide scope crossed tenant boundary")
	}
	if AllowsOwner(Decision{TenantID: 7, UserIDs: []int64{11}}, 7, 12) {
		t.Fatal("user scope crossed owner boundary")
	}
}

func TestDecisionFromReplyCopiesScope(t *testing.T) {
	reply := &pb.AccessReply{TenantId: 7, UserIds: []int64{11}, Authorization: "job-1", OperationId: "tenant.delete"}
	d := decisionFromReply(reply)
	reply.UserIds[0] = 99
	if d.TenantID != 7 || d.Authorization != "job-1" || d.OperationID != "tenant.delete" || d.UserIDs[0] != 11 {
		t.Fatalf("unexpected decision: %+v", d)
	}
}
