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

	"github.com/sentinez/sentinez/internal/apiserver/handlers"
	"github.com/sentinez/sentinez/internal/apiserver/middleware"
	"github.com/sentinez/sentinez/internal/apiserver/services/v1"
	greeterfac "github.com/sentinez/sentinez/internal/core/greeter/v1/factory"
	iamfac "github.com/sentinez/sentinez/internal/core/iam/v1/factory"
	tenantfac "github.com/sentinez/sentinez/internal/core/tenant/v1/factory"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func (srv *Server) Bootloader(ctx context.Context) error {
	rctx := runner.GetContext(ctx)
	flag := rctx.GetFlag()

	// load all middleware and handlers of api server
	srv.server.Use(middleware.Logging)
	srv.server.Use(middleware.AllowCORS)

	// register all custom handlers to the server
	handlers.RegisterSwaggerRoutes(srv.server.HTTPMux(), flag)

	// NOTE: Make sure the gRPC server is running properly and accessible
	// Create file at registrar, inherit base package, override function,
	// implement business logic
	err := srv.visitToEndpoint(ctx,
		services.NewGreeter(greeterfac.NewDefaultGreeterHdl(rctx)),
	)
	if err != nil {
		zlog.Errorf("apiserver: failed to visit service: %v", err)
		return err
	}

	return srv.visit(ctx,
		services.NewIAM(iamfac.NewDefaultIAMHdl(rctx)),
		services.NewTenant(tenantfac.NewDefaultTenantHdl(rctx)),
	)
}
