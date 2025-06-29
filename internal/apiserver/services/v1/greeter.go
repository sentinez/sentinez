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
	"time"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/discovery/v1"
	greeterpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"
	"github.com/sentinez/sentinez/pkg/common/cron"
	httpgw "github.com/sentinez/sentinez/pkg/core/gateway/http"
	"github.com/sentinez/sentinez/pkg/std/grpc/client"
	"github.com/sentinez/sentinez/pkg/std/names"
	"github.com/sentinez/sentinez/pkg/std/zlog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ httpgw.ServiceRegistrar = (*greeter)(nil)

// NewGreeter creates a new greeter service to register handler to gateway
func NewGreeter(server greeterpb.GreeterServiceServer) httpgw.ServiceRegistrar {
	return &greeter{server: server}
}

// greeter represents the greeter service
type greeter struct {
	server greeterpb.GreeterServiceServer
}

// AcceptFromEndpoint implements httpgw.ServiceRegistrar.
func (g *greeter) AcceptFromEndpoint(ctx context.Context,
	server httpgw.Server) error {

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	cron.Start(ctx, time.Second*10, func() {
		disc, err := client.NewDiscoveryClient()
		if err != nil {
			return
		}

		resp, err := disc.Discover(ctx,
			&discovery.DiscoverRequest{Name: names.GreeterV1.String()})
		if err != nil {
			return
		}

		err = greeterpb.RegisterGreeterServiceHandlerFromEndpoint(
			ctx, server.RuntimeMux(), resp.GetAddress(), opts)
		if err == nil {
			zlog.Debug("[apiserver] greeter service: ", resp.GetAddress())
		}

	})

	return nil
}

// Accept accepts the greeter service
func (g *greeter) Accept(ctx context.Context, server httpgw.Server) error {
	return greeterpb.
		RegisterGreeterServiceHandlerServer(ctx, server.RuntimeMux(), g.server)
}
