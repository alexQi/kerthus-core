// Package saasv1 exposes the versioned platform RPC wire contract. The alias
// layer is temporary: phase B moves generated code here without changing the
// application-facing method names.
package saasv1

import internal "github.com/alexQi/kerthus-core/gen/go/saas/v1"

type PlatformService = internal.PlatformService
type Context = internal.Context
type AccessRequest = internal.AccessRequest
type AccessReply = internal.AccessReply
type DelegationRequest = internal.DelegationRequest
type DelegationReply = internal.DelegationReply
type DelegationCheck = internal.DelegationCheck
type AppAuthorization = internal.AppAuthorization
type JobAuthorizationRequest = internal.JobAuthorizationRequest
type BackgroundJobAuthorizationRequest = internal.BackgroundJobAuthorizationRequest
type RenewBackgroundJobAuthorizationRequest = internal.RenewBackgroundJobAuthorizationRequest
type BackgroundModelInputAuthorizationRequest = internal.BackgroundModelInputAuthorizationRequest
type ModelInputAuthorizationRequest = internal.ModelInputAuthorizationRequest
type ModelInputRef = internal.ModelInputRef
type ModelCatalogReply = internal.ModelCatalogReply
type ModelSubmitRequest = internal.ModelSubmitRequest
type ModelLookupRequest = internal.ModelLookupRequest
type ModelCallRequest = internal.ModelCallRequest
type ModelCallReply = internal.ModelCallReply
type ModelCallDiagnostic = internal.ModelCallDiagnostic
type MediaBeginRequest = internal.MediaBeginRequest
type MediaRequest = internal.MediaRequest
type MediaPartRequest = internal.MediaPartRequest
type MediaReply = internal.MediaReply
type MediaReadRequest = internal.MediaReadRequest
type MediaChunk = internal.MediaChunk
type PlaybackReply = internal.PlaybackReply
type JobDerivationRequest = internal.JobDerivationRequest
type PromoteJobSourceRequest = internal.PromoteJobSourceRequest
type ModelRunRequest = internal.ModelRunRequest
type ModelRunControlRequest = internal.ModelRunControlRequest
type ModelRunReply = internal.ModelRunReply
type ModelRebindRequest = internal.ModelRebindRequest
type JobResumeRequest = internal.JobResumeRequest
type JobReceiptRequest = internal.JobReceiptRequest
type HistoricalModelReceiptRequest = internal.HistoricalModelReceiptRequest
type HistoricalModelReceiptReply = internal.HistoricalModelReceiptReply
type PlaybackReadRequest = internal.PlaybackReadRequest
type UploadRequest = internal.UploadRequest
type ConfirmUploadRequest = internal.ConfirmUploadRequest
type UploadReply = internal.UploadReply
type FileRequest = internal.FileRequest
type ApplicationToolsReply = internal.ApplicationToolsReply
type ApplicationToolRequest = internal.ApplicationToolRequest
