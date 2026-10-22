package models

import (
	"context"
	"testing"

	contractmodels "github.com/alexQi/kerthus-core/models"
	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
	"go-micro.dev/v6/client"
)

func TestReceiptFromReplyMapsStableFields(t *testing.T) {
	reply := &pb.ModelCallReply{
		CallId:     42,
		State:      "completed",
		RequestId:  "req-42",
		ResultJson: `{"ok":true}`,
		Diagnostic: &pb.ModelCallDiagnostic{Code: "provider_timeout", Model: "gpt-image-2", Message: "upstream timed out"},
		QueuePhase: "finished",
	}
	invocation := &contractmodels.Invocation{IdempotencyKey: "idem-42", Attempt: 2, Provider: "openai", RootRunID: "run-42"}
	receipt := receiptFromReply(reply, invocation)

	if receipt.CallID != 42 || receipt.State != contractmodels.StateSucceeded || receipt.RequestID != "req-42" {
		t.Fatalf("unexpected receipt identity/state: %#v", receipt)
	}
	if receipt.IdempotencyKey != "idem-42" || receipt.Attempt != 2 || receipt.Provider != "openai" || receipt.RootRunID != "run-42" {
		t.Fatalf("invocation metadata was not preserved: %#v", receipt)
	}
	if receipt.ErrorClass != "provider_timeout" || receipt.Model != "gpt-image-2" || receipt.Message != "upstream timed out" {
		t.Fatalf("diagnostic fields were not normalized: %#v", receipt)
	}
}

func TestReceiptFromReplyUnknownStateIsConservative(t *testing.T) {
	receipt := receiptFromReply(&pb.ModelCallReply{State: "future_state"}, nil)
	if receipt.State != contractmodels.StateUnknown {
		t.Fatalf("future state was not made unknown: %#v", receipt)
	}
}

func TestStableInvocationRejectsUnscopedInput(t *testing.T) {
	invocation := contractmodels.Invocation{Authorization: "job", Selector: "image", IdempotencyKey: "idem", Inputs: []contractmodels.Input{{Role: "image"}}}
	if err := invocation.Validate(); err == nil {
		t.Fatal("input without a media ID was accepted")
	}
}

type receiptPlatform struct {
	pb.PlatformService
	state                     string
	queries, rebinds, submits int
}

func (p *receiptPlatform) QueryApplicationModel(_ context.Context, _ *pb.ModelCallRequest, _ ...client.CallOption) (*pb.ModelCallReply, error) {
	p.queries++
	return &pb.ModelCallReply{CallId: 42, State: p.state}, nil
}
func (p *receiptPlatform) RebindApplicationModel(_ context.Context, _ *pb.ModelRebindRequest, _ ...client.CallOption) (*pb.ModelCallReply, error) {
	p.rebinds++
	return &pb.ModelCallReply{CallId: 42, State: p.state}, nil
}
func (p *receiptPlatform) SubmitApplicationModel(_ context.Context, _ *pb.ModelSubmitRequest, _ ...client.CallOption) (*pb.ModelCallReply, error) {
	p.submits++
	return &pb.ModelCallReply{CallId: 43, State: "completed"}, nil
}

func TestStableUnknownAndWaitingReceiptNeverResubmit(t *testing.T) {
	p := &receiptPlatform{state: "unknown"}
	c, err := NewStable(p, "fablex", "01234567890123456789012345678901")
	if err != nil {
		t.Fatal(err)
	}
	query := contractmodels.QueryRequest{Authorization: "grant", CallID: 42}
	for _, state := range []struct {
		wire string
		want contractmodels.State
	}{{"unknown", contractmodels.StateUnknown}, {"waiting_receipt", contractmodels.StateWaitingReceipt}} {
		p.state = state.wire
		got, err := c.Query(context.Background(), query)
		if err != nil || got.State != state.want {
			t.Fatal(got, err)
		}
		got, err = c.Reconcile(context.Background(), contractmodels.ReconcileRequest{Authorization: "grant", CallID: 42, IdempotencyKey: "original"})
		if err != nil || got.State != state.want || got.IdempotencyKey != "original" {
			t.Fatal(got, err)
		}
	}
	if p.queries != 2 || p.rebinds != 2 || p.submits != 0 {
		t.Fatal("read-only recovery submitted a paid call")
	}
}
