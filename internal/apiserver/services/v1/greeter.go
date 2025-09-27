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

	greeterpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	httpgw "github.com/sentinez/sentinez/pkg/core/net/httpx/gw"
)

var _ httpgw.ServiceRegistrar = (*Greeter)(nil)

// NewGreeter creates a new greeter service to register handler to gateway
func NewGreeter(server greeterpb.GreeterServiceServer) *Greeter {
	return &Greeter{server: server}
}

// Greeter represents the Greeter service
type Greeter struct {
	server greeterpb.GreeterServiceServer
}

// AcceptFromEndpoint implements httpgw.ServiceRegistrar.
func (g *Greeter) AcceptFromEndpoint(ctx context.Context,
	server httpgw.Server, appConf *common.AppConfig) error {

	return httpgw.RegisterServiceFromEndpoint(ctx,
		appConf,
		server.RuntimeMux(),
		greeterpb.GetMetaGreeterServiceKey(),
		greeterpb.RegisterGreeterServiceHandlerFromEndpoint,
	)
}

// Accept accepts the greeter service
func (g *Greeter) Accept(ctx context.Context, server httpgw.Server) error {

	return httpgw.RegisterServiceHandlerServer(ctx,
		server.RuntimeMux(),
		g.server,
		greeterpb.RegisterGreeterServiceHandlerServer,
	)
}
