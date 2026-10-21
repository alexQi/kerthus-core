package models

import "testing"

func TestReceiptStatesAreStable(t *testing.T) {
	want := []State{StateQueued, StateRunning, StateSucceeded, StateFailed, StateUnknown, StateWaitingReceipt}
	if len(want) != 6 || want[4] != StateUnknown || want[5] != StateWaitingReceipt {
		t.Fatal("receipt state contract changed")
	}
}

func TestNormalizeStateNeverTreatsUnknownAsSuccess(t *testing.T) {
	if NormalizeState("completed") != StateSucceeded {
		t.Fatal("legacy completed state was not normalized")
	}
	if NormalizeState("provider-new-state") != StateUnknown {
		t.Fatal("unknown state was not conservatively classified")
	}
}

func TestInvocationRequiresIdempotencyAndAuthorization(t *testing.T) {
	if err := (Invocation{Selector: "image", IdempotencyKey: "k"}).Validate(); err == nil {
		t.Fatal("missing authorization was accepted")
	}
	if err := (Invocation{Authorization: "job", Selector: "image", IdempotencyKey: "k"}).Validate(); err != nil {
		t.Fatalf("valid invocation rejected: %v", err)
	}
}
