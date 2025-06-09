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

// Package greeter implements the Greeter service server.
package greeter

import (
	"context"
	"time"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"
	sentinezpb "github.com/sentinez/sentinez/api/gen/go/sentinez/v1"
	greeterdomain "github.com/sentinez/sentinez/internal/core/greeter/v1/domain"
	greeterhandler "github.com/sentinez/sentinez/internal/core/greeter/v1/handler"
	"github.com/sentinez/sentinez/pkg/common/protobuf"
	grpcgw "github.com/sentinez/sentinez/pkg/core/gateway/grpc"
	"github.com/sentinez/sentinez/pkg/core/sentinez/v1"
	"github.com/sentinez/sentinez/pkg/std/names"
)

// make sure Greeter implement sentinez.Server
// it will start by sentinez.runner through sentinez.Server
var _ sentinez.Server = (*Greeter)(nil)

// inject all dependencies to the greeter
// This is a dependency injection pattern.
var (
	_ = sentinez.Inject(greeterhandler.New)
	_ = sentinez.Inject(greeterdomain.New)
)

// New creates a new Greeter module.
func New(srv greeter.GreeterServiceServer, conf *sentinezpb.Config,
	flag *sentinezpb.FlagGRPCService) sentinez.Server {

	return &Greeter{
		Server: grpcgw.NewDefault(),
		srv:    srv,
		config: conf,
		flag:   flag,
	}
}

// Greeter implements GreeterServiceServer.
type Greeter struct {
	*grpcgw.Server // inherit grpc.Server
	config         *sentinezpb.Config
	flag           *sentinezpb.FlagGRPCService
	srv            greeter.GreeterServiceServer
}

// Start implements IGreeter, override sentinez.Server.Start
func (g *Greeter) Start(_ context.Context) error {
	greeter.PrintASCII()
	if err := protobuf.Validate(g.config); err != nil {
		return err
	}

	greeter.RegisterGreeterServiceServer(g.AsServer(), g.srv)

	return g.Serve(&grpcgw.ServiceInfo{
		Config:      g.config,
		Addr:        g.flag.GetAddress(),
		GatewayAddr: g.flag.GetGatewayAddress(),
		Name:        names.GreeterV1.String(),
		TTL:         time.Minute * 5,
	})
}
