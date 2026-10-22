// Package saasv1 exposes the versioned platform RPC wire contract. The alias
// layer preserves the generated v1 wire contract at the public import path.
package saasv1

import (
	internal "github.com/alexQi/kerthus-core/gen/go/saas/v1"
	"go-micro.dev/v6/client"
	"go-micro.dev/v6/server"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type PlatformService = internal.PlatformService
type PlatformHandler = internal.PlatformHandler
type Kind = internal.Kind
type Empty = internal.Empty
type Result = internal.Result
type LoginRequest = internal.LoginRequest
type LoginReply = internal.LoginReply
type User = internal.User
type Tenant = internal.Tenant
type Member = internal.Member
type Org = internal.Org
type Position = internal.Position
type App = internal.App
type Resource = internal.Resource
type Operation = internal.Operation
type Role = internal.Role
type ResourceGrant = internal.ResourceGrant
type TenantApp = internal.TenantApp
type District = internal.District
type Audit = internal.Audit
type AuthReply = internal.AuthReply
type QueryRequest = internal.QueryRequest
type QueryReply = internal.QueryReply
type SaveRequest = internal.SaveRequest
type SaveReply = internal.SaveReply
type SaveRequest_User = internal.SaveRequest_User
type SaveRequest_Tenant = internal.SaveRequest_Tenant
type SaveRequest_Member = internal.SaveRequest_Member
type SaveRequest_Org = internal.SaveRequest_Org
type SaveRequest_Position = internal.SaveRequest_Position
type SaveRequest_App = internal.SaveRequest_App
type SaveRequest_Resource = internal.SaveRequest_Resource
type SaveRequest_Role = internal.SaveRequest_Role
type DeleteRequest = internal.DeleteRequest
type StatusRequest = internal.StatusRequest
type PasswordRequest = internal.PasswordRequest
type ApproveRequest = internal.ApproveRequest
type EntitlementsRequest = internal.EntitlementsRequest
type RoleMembersRequest = internal.RoleMembersRequest
type RoleResourcesRequest = internal.RoleResourcesRequest
type AppGrant = internal.AppGrant
type ApplicationTool = internal.ApplicationTool
type ModelCatalogItem = internal.ModelCatalogItem
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

const (
	Kind_KIND_UNSPECIFIED = internal.Kind_KIND_UNSPECIFIED
	Kind_USERS            = internal.Kind_USERS
	Kind_TENANTS          = internal.Kind_TENANTS
	Kind_MEMBERS          = internal.Kind_MEMBERS
	Kind_ORGS             = internal.Kind_ORGS
	Kind_POSITIONS        = internal.Kind_POSITIONS
	Kind_APPS             = internal.Kind_APPS
	Kind_RESOURCES        = internal.Kind_RESOURCES
	Kind_OPERATIONS       = internal.Kind_OPERATIONS
	Kind_TENANT_APPS      = internal.Kind_TENANT_APPS
	Kind_ROLES            = internal.Kind_ROLES
	Kind_DISTRICTS        = internal.Kind_DISTRICTS
	Kind_AUDITS           = internal.Kind_AUDITS
)

var (
	Kind_name                                           = internal.Kind_name
	Kind_value                                          = internal.Kind_value
	File_saas_v1_core_proto protoreflect.FileDescriptor = internal.File_saas_v1_core_proto
)

func NewPlatformService(name string, c client.Client) PlatformService {
	return internal.NewPlatformService(name, c)
}

func RegisterPlatformHandler(s server.Server, h PlatformHandler, opts ...server.HandlerOption) error {
	return internal.RegisterPlatformHandler(s, h, opts...)
}
