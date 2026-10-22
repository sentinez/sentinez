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

package controlplane

import (
	"context"

	"github.com/sentinez/core/szgrpc/szgrpcgateway"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var _ szgrpcgateway.ServiceRegistrar = (*identityAccessManagement)(nil)

// NewIAM creates a new iam service registrar.
func NewIAM(
	srv iampb.IdentityAccessManagementServiceServer,
) szgrpcgateway.ServiceRegistrar {
	return &identityAccessManagement{server: srv}
}

// identityAccessManagement is the identityAccessManagement service registrar.
type identityAccessManagement struct {
	server iampb.IdentityAccessManagementServiceServer
}

// AcceptFromEndpoint implements httpgw.ServiceRegistrar.
func (i *identityAccessManagement) AcceptFromEndpoint(ctx context.Context,
	server szgrpcgateway.Server, appConf *settingpb.Config) error {

	return szgrpcgateway.RegisterServiceFromEndpoint(ctx,
		appConf,
		server.RuntimeMux(),
		iampb.GetMetaIamServiceKey(),
		iampb.RegisterIdentityAccessManagementServiceHandlerFromEndpoint,
	)
}

// Accept to visit the iam service.
func (i *identityAccessManagement) Accept(ctx context.Context,
	server szgrpcgateway.Server) error {

	// a local conn, unlike the generated HandlerServer registration, runs
	// interceptors, so requests are validated at the gateway.
	conn := szgrpcgateway.NewLocalConn()
	iampb.RegisterIdentityAccessManagementServiceServer(conn, i.server)

	return iampb.RegisterIdentityAccessManagementServiceHandlerClient(ctx,
		server.RuntimeMux(),
		iampb.NewIdentityAccessManagementServiceClient(conn),
	)
}
