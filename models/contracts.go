// Package models contains provider-neutral model invocation and receipt
// contracts. It intentionally contains no provider client, database model, or
// application-domain semantics.
package models

import (
	"fmt"
	"strings"
)

type State string

const (
	StateQueued         State = "queued"
	StateRunning        State = "running"
	StateSucceeded      State = "succeeded"
	StateFailed         State = "failed"
	StateUnknown        State = "unknown"
	StateWaitingReceipt State = "waiting_receipt"
)

// NormalizeState maps legacy platform state spellings to the stable public
// vocabulary. Unknown future values intentionally remain unknown so callers
// never treat an unrecognized state as a successful side effect.
func NormalizeState(value string) State {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "queued", "pending", "accepted":
		return StateQueued
	case "running", "processing", "in_progress", "in-progress":
		return StateRunning
	case "succeeded", "success", "completed", "done":
		return StateSucceeded
	case "failed", "failure", "error", "cancelled", "canceled":
		return StateFailed
	case "waiting_receipt", "waiting-receipt", "reconcile":
		return StateWaitingReceipt
	case "unknown", "lost", "timeout", "timed_out":
		return StateUnknown
	default:
		return StateUnknown
	}
}

// Input is a scoped media reference for a model invocation. Authorization is
// deliberately absent: the application obtains a short-lived job permit via
// the media/authorization clients and the platform binds it server-side.
type Input struct {
	MediaID        int64
	ContentVersion string
	Role           string
}

type Invocation struct {
	Authorization   string
	Selector        string
	IdempotencyKey  string
	Prompt          string
	Size            string
	Seconds         int32
	Texts           []string
	Inputs          []Input
	MaxOutputTokens int64
	Provider        string
	Model           string
	Attempt         int32
	RootRunID       string
}

// Diagnostic is the provider-neutral execution diagnostic returned with a
// model receipt. It contains bounded execution metadata and retry guidance;
// provider credentials, prompts, URLs, and raw upstream bodies are excluded.
type Diagnostic struct {
	Code               string
	Phase              string
	HTTPStatus         int32
	UpstreamRequestID  string
	Submission         string
	Message            string
	RetrySafe          bool
	UpstreamStatusCode string
	UpstreamLogID      string
	RequestBytes       int32
	RequestSHA256      string
	RouteSHA256        string
	Protocol           string
	Model              string
	StartedAtUnixMS    int64
	FinishedAtUnixMS   int64
	DNSMS              int64
	ConnectMS          int64
	TLSMS              int64
	WriteMS            int64
	FirstByteMS        int64
	ReadMS             int64
	HTTPVersion        string
	ALPN               string
	RemoteAddr         string
	UpstreamErrorCode  string
	UpstreamErrorType  string
	UpstreamMessage    string
	RetryClass         string
	RetryAfterSeconds  int64
	RetryPolicy        string
	ExecutionOutcome   string
}

type Receipt struct {
	CallID         int64
	State          State
	IdempotencyKey string
	Attempt        int32
	Provider       string
	Model          string
	ErrorClass     string
	Message        string
	RequestID      string
	ResultJSON     string
	UsageJSON      string
	QueuePhase     string
	RootRunID      string
	Diagnostic     *Diagnostic
}

type QueryRequest struct {
	Authorization string
	CallID        int64
}

type ReconcileRequest struct {
	Authorization  string
	CallID         int64
	Selector       string
	IdempotencyKey string
	RootRunID      string
}

func (i Invocation) Validate() error {
	if i.Authorization == "" || i.Selector == "" || i.IdempotencyKey == "" {
		return fmt.Errorf("authorization, selector, and idempotency key are required")
	}
	if i.Attempt < 0 {
		return fmt.Errorf("attempt cannot be negative")
	}
	if i.Seconds < 0 || i.MaxOutputTokens < 0 {
		return fmt.Errorf("seconds and max output tokens cannot be negative")
	}
	for _, input := range i.Inputs {
		if input.MediaID <= 0 {
			return fmt.Errorf("model input media ID is required")
		}
	}
	return nil
}

func (q QueryRequest) Validate() error {
	if q.Authorization == "" || q.CallID <= 0 {
		return fmt.Errorf("authorization and call ID are required")
	}
	return nil
}

func (r ReconcileRequest) Validate() error {
	if r.Authorization == "" || r.CallID <= 0 || r.IdempotencyKey == "" {
		return fmt.Errorf("authorization, call ID, and idempotency key are required")
	}
	return nil
}
