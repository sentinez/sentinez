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

	greeterfac "github.com/sentinez/controlplane/greeter/v1/factory"
	iamfac "github.com/sentinez/controlplane/iam/v1/factory"
	securityfac "github.com/sentinez/controlplane/security/v1/factory"
	tenantfac "github.com/sentinez/controlplane/tenant/v1/factory"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/setting/v1"
	"github.com/sentinez/sentinez/pkg/apps/acz/apiserver/controlplane/v1"
	"github.com/sentinez/sentinez/pkg/apps/acz/apiserver/handlers"
	"github.com/sentinez/sentinez/pkg/apps/acz/apiserver/middleware"
)

func (srv *Server) Initialize(
	ctx context.Context, conf *settingpb.Config) error {

	flag := conf.GetFlag()

	// load all middleware and handlers of api server
	srv.server.Use(middleware.Logging)
	srv.server.Use(middleware.AllowCORS)

	// register all custom handlers to the server
	handlers.RegisterSwaggerRoutes(srv.server.HTTPMux(), flag)

	return srv.Visit(ctx,
		controlplane.NewGreeter(greeterfac.NewDefaultHandler(conf)),
		controlplane.NewIAM(iamfac.NewDefaultHandler(ctx, conf)),
		controlplane.NewSecurity(securityfac.NewDefaultHandler(ctx, conf)),
		controlplane.NewTenant(tenantfac.NewDefaultHandler(ctx, conf)),
	)
}
