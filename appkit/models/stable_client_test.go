package models

import (
	"testing"

	contractmodels "github.com/alexQi/kerthus-core/models"
	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
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
