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

package services

import (
	"context"

	iampb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	httpgw "github.com/sentinez/sentinez/pkg/core/gateway/http"
	"github.com/sentinez/sentinez/pkg/std/eventq"
	"github.com/sentinez/sentinez/pkg/std/names"
	"github.com/sentinez/sentinez/pkg/std/zlog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ httpgw.ServiceRegistrar = (*identityAccessManagement)(nil)

// NewIAM creates a new iam service registrar.
func NewIAM(
	srv iampb.IdentityAccessManagementServiceServer) httpgw.ServiceRegistrar {
	return &identityAccessManagement{server: srv}
}

// identityAccessManagement is the identityAccessManagement service registrar.
type identityAccessManagement struct {
	server iampb.IdentityAccessManagementServiceServer
}

// AcceptFromEndpoint implements httpgw.ServiceRegistrar.
func (i *identityAccessManagement) AcceptFromEndpoint(ctx context.Context,
	server httpgw.Server) error {

	eventq.Subscribe(ctx, names.IAMV1.String(), func(endpoint string) error {
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		zlog.Infof("[visitor.VisitServiceFromEndpoint] %s %s",
			names.IAMV1.String(), "******")

		return iampb.RegisterIdentityAccessManagementServiceHandlerFromEndpoint(
			ctx, server.RuntimeMux(), endpoint, opts)
	})

	return nil
}

// Accept to visit the iam service.
func (i *identityAccessManagement) Accept(ctx context.Context,
	server httpgw.Server) error {

	return iampb.RegisterIdentityAccessManagementServiceHandlerServer(ctx,
		server.RuntimeMux(), i.server)
}
