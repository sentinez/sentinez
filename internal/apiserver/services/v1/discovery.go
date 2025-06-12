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

// Package services provides all service declare.
package services

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	discoverypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/discovery/v1"
	httpgw "github.com/sentinez/sentinez/pkg/core/gateway/http"
	"github.com/sentinez/sentinez/pkg/std/eventq"
	"github.com/sentinez/sentinez/pkg/std/names"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var _ httpgw.ServiceRegistrar = (*discovery)(nil)

// NewDiscovery creates a new Discovery service to register handler to gateway
func NewDiscovery(
	srv discoverypb.DiscoveryServiceServer) httpgw.ServiceRegistrar {
	return &discovery{server: srv}
}

// discovery represents the discovery service
type discovery struct {
	server discoverypb.DiscoveryServiceServer
}

// AcceptFromEndpoint implements httpgw.ServiceRegistrar.
func (d *discovery) AcceptFromEndpoint(
	ctx context.Context, server httpgw.Server) error {

	eventq.Subscribe(ctx, names.DiscoveryV1.String(),
		func(endpoint string) error {
			opts := []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			}

			zlog.Infof("[visitor.VisitServiceFromEndpoint] %s %s",
				names.DiscoveryV1.String(), "******")

			return discoverypb.RegisterDiscoveryServiceHandlerFromEndpoint(
				ctx, server.RuntimeMux(), endpoint, opts)
		})

	return nil
}

// Accept accepts the Discovery service
func (d *discovery) Accept(ctx context.Context, server httpgw.Server) error {
	return discoverypb.RegisterDiscoveryServiceHandlerServer(
		ctx, server.RuntimeMux(), d.server)
}
