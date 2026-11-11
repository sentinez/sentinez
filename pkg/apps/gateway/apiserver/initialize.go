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
	greeterfac "github.com/sentinez/controlplane/apps/greeter/v1/factory"
	iamfac "github.com/sentinez/controlplane/apps/iam/v1/factory"
	securityfac "github.com/sentinez/controlplane/apps/security/v1/factory"
	securityhdl "github.com/sentinez/controlplane/apps/security/v1/handler"
	tenantfac "github.com/sentinez/controlplane/apps/tenant/v1/factory"
	tenanthdl "github.com/sentinez/controlplane/apps/tenant/v1/handler"
	"github.com/sentinez/controlplane/pkg/passkey"
	"github.com/sentinez/pkg/szlog"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/pkg/apps/gateway/apiserver/controlplane/v1"
	"github.com/sentinez/sentinez/pkg/apps/gateway/apiserver/handlers"
	"github.com/sentinez/sentinez/pkg/apps/gateway/apiserver/middleware"
)

func (srv *Server) Initialize(
	ctx context.Context, conf *settingpb.Config) error {

	flag := conf.GetFlag()

	// load all middleware and handlers of api server
	srv.server.Use(middleware.Logging)
	origins := passkey.ClientOrigins(conf)
	if len(origins) == 0 {
		szlog.Warnf("no SENZ_CLIENT_ORIGIN: cross origin calls are refused")
	}

	srv.server.Use(middleware.CORS(origins))
	srv.server.Use(middleware.SwaggerDefaultRoot)
	// security is left out: SecRule payloads legitimately carry attack
	// patterns and would be rejected.
	srv.server.Use(middleware.Scope(
		middleware.Injection, "/iam", "/tenant", "/greeter", "/analytic",
	))

	// register all custom handlers to the server
	handlers.RegisterSwaggerRoutes(srv.server.HTTPMux(), flag)

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

	analytic, err := analyticfac.NewDefaultHandler(ctx, conf)
	if err != nil {
		return err
	}

	return srv.Visit(ctx,
		controlplane.NewAnalytic(analytic),
		controlplane.NewGreeter(greeterfac.NewDefaultHandler(conf)),
		controlplane.NewIAM(iamfac.NewDefaultHandler(ctx, conf)),
		controlplane.NewSecurity(securityhdl.New(security)),
		controlplane.NewTenant(tenanthdl.New(tenant)),
	)
}
