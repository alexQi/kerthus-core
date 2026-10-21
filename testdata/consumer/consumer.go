package consumer

import (
	"context"

	"github.com/alexQi/kerthus-core/appcontract"
	"github.com/alexQi/kerthus-core/appidentity"
	appauth "github.com/alexQi/kerthus-core/appkit/authorization"
	appdb "github.com/alexQi/kerthus-core/appkit/database"
	appmedia "github.com/alexQi/kerthus-core/appkit/media"
	appmodels "github.com/alexQi/kerthus-core/appkit/models"
	"github.com/alexQi/kerthus-core/authorization"
	appctx "github.com/alexQi/kerthus-core/context"
	mediacontract "github.com/alexQi/kerthus-core/media"
	modelcontract "github.com/alexQi/kerthus-core/models"
	rpcv1 "github.com/alexQi/kerthus-core/rpc/saas/v1"
)

// Compile verifies the public module can be consumed without a Kerthus checkout
// or any internal package import.
func Compile(platform rpcv1.PlatformService, serviceKey string) error {
	if _, err := appidentity.DatabaseName("fablex"); err != nil {
		return err
	}
	if err := appdb.ValidateDSN("user:secret@tcp(localhost:3306)/kerthus_fablex", "fablex"); err != nil {
		return err
	}
	_ = appcontract.Digest([]byte("public"))
	if err := (authorization.Request{Context: authorization.Context{TenantID: 1}, AppCode: "fablex", Method: "GET", Path: "/api/apps/fablex/v1/health"}).Validate(); err != nil {
		return err
	}
	if _, err := authorization.New(platform, "fablex", serviceKey); err != nil {
		return err
	}
	if _, err := appauth.New(platform, "fablex", serviceKey); err != nil {
		return err
	}
	if _, err := appmodels.New(platform, "fablex", serviceKey); err != nil {
		return err
	}
	if _, err := appmedia.New(platform, "fablex", serviceKey, func(context.Context, *rpcv1.AccessReply, appmedia.Scope) error { return nil }); err != nil {
		return err
	}
	var scope mediacontract.Scope
	var invocation modelcontract.Invocation
	_ = appctx.Scope{TenantID: 1, App: appctx.App{ID: 1, Code: "fablex"}, RequestID: "req-1"}
	_ = scope
	_ = invocation
	return nil
}
