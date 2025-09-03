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

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"
	greeterhdl "github.com/sentinez/sentinez/internal/core/greeter/v1/handler"
	"github.com/sentinez/sentinez/pkg/common/protobuf"
	grpcgw "github.com/sentinez/sentinez/pkg/core/gateway/grpc"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
)

// make sure Greeter implement runner.Server
// it will start by runner/v1.runner through runner.Server
var _ runner.Engine = (*Greeter)(nil)

type Service struct {
	*grpcgw.Server
	handler greeter.GreeterServiceServer
}

func NewService(runnerCtx *runner.Context) *Service {
	return &Service{
		Server:  grpcgw.NewDefault(runnerCtx),
		handler: greeterhdl.New(),
	}
}

// New creates a new Greeter module.
func New(srv *Service) runner.Engine {

	return &Greeter{
		Service: srv,
	}
}

// Greeter implements GreeterServiceServer.
type Greeter struct {
	*Service
}

// Start implements IGreeter, override runner.Server.Start
func (g *Greeter) Start(_ context.Context) error {
	if err := protobuf.Validate(g.GetConfig()); err != nil {
		return err
	}

	greeter.RegisterGreeterServiceServer(g.AsServer(), g.handler)

	go grpcgw.Register(greeter.GetMetaGreeterServiceKey(), g.GetConfig())
	return g.Serve(g.GetConfig().GetAddress())
}
