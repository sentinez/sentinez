// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apiserver

import (
	"context"

	analyticfac "github.com/sentinez/controlplane/apps/analytic/v1/factory"
	commfac "github.com/sentinez/controlplane/apps/communication/v1/factory"
	commhdl "github.com/sentinez/controlplane/apps/communication/v1/handler"
	greeterfac "github.com/sentinez/controlplane/apps/greeter/v1/factory"
	iamfac "github.com/sentinez/controlplane/apps/iam/v1/factory"
	securityfac "github.com/sentinez/controlplane/apps/security/v1/factory"
	securityhdl "github.com/sentinez/controlplane/apps/security/v1/handler"
	tenantfac "github.com/sentinez/controlplane/apps/tenant/v1/factory"
	tenanthdl "github.com/sentinez/controlplane/apps/tenant/v1/handler"
	"github.com/sentinez/controlplane/pkg/passkey"
	"github.com/sentinez/core/szgrpc/client"
	"github.com/sentinez/pkg/szlog"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/apps/apiserver/controlplane/v1"
	"github.com/sentinez/sentinez/internal/apps/apiserver/handlers"
	"github.com/sentinez/sentinez/internal/apps/apiserver/middleware"
	"github.com/sentinez/sentinez/pkg/interceptor"
)

func (srv *Server) Initialize(
	ctx context.Context, conf *settingpb.Config) error {

	srv.useMiddleware(conf)

	// register all custom handlers to the server
	handlers.RegisterSwaggerRoutes(srv.server.HTTPMux())

	security, err := securityfac.NewDefaultService(ctx, conf)
	if err != nil {
		return err
	}

	tenant, err := tenantfac.NewDefaultService(ctx, conf)
	if err != nil {
		return err
	}

	// Consul down or a bad proxy.yaml must not stop the API.
	if err := importProxy(ctx, conf, tenant, security); err != nil {
		szlog.Warnf("proxy import: %v", err)
	}

	return srv.visitServices(ctx, conf,
		securityhdl.New(security), tenanthdl.New(tenant))
}

// visitServices registers the services of the control plane, each with the
// audit log in front of its authorization.
func (srv *Server) visitServices(ctx context.Context,
	conf *settingpb.Config, security securitypb.SecurityServiceServer,
	tenant tenantpb.TenantServiceServer) error {

	analytic, err := analyticfac.NewDefaultHandler(ctx, conf)
	if err != nil {
		return err
	}

	communication, err := srv.useNotifications(ctx, conf)
	if err != nil {
		return err
	}

	audit, exporter, err := auditInterceptor(conf)
	if err != nil {
		return err
	}

	srv.auditExporter = exporter

	return srv.Visit(ctx,
		controlplane.NewAnalytic(analytic, audit),
		controlplane.NewCommunication(communication, audit),
		controlplane.NewGreeter(greeterfac.NewDefaultHandler(conf), audit),
		controlplane.NewIAM(iamfac.NewDefaultHandler(ctx, conf), audit),
		controlplane.NewSecurity(security, audit),
		controlplane.NewTenant(tenant, audit),
	)
}

// useNotifications serves the notification stream of the communication
// handler it returns as server-sent events, authorized as on the network.
func (srv *Server) useNotifications(ctx context.Context,
	conf *settingpb.Config) (*commhdl.Communication, error) {
	communication, err := commfac.NewDefaultHandler(ctx, conf)
	if err != nil {
		return nil, err
	}

	cli, conn := client.NewLocalCommunication(communication,
		interceptor.LocalOptions()...)
	srv.notifications = conn

	handlers.RegisterServerSentEvents(srv.server.HTTPMux(), cli)

	return communication, nil
}

// useMiddleware loads the middleware of the api server.
func (srv *Server) useMiddleware(conf *settingpb.Config) {
	srv.server.Use(middleware.Logging)

	origins := passkey.ClientOrigins(conf)
	if len(origins) == 0 {
		szlog.Warnf("no SZ_CLIENT_ORIGIN: cross origin calls are refused")
	}

	srv.server.Use(middleware.CORS(origins))
	srv.server.Use(middleware.SwaggerDefaultRoot)
	// security is left out: SecRule payloads legitimately carry attack
	// patterns and would be rejected.
	srv.server.Use(middleware.Scope(
		middleware.Injection, "/iam", "/tenant", "/greeter", "/analytic",
		"/communication",
	))
}
