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
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	greeterhdl "github.com/sentinez/sentinez/internal/core/greeter/v1/handler"
	grpcgw "github.com/sentinez/sentinez/pkg/core/gateway/grpc"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
)

type Service struct {
	*grpcgw.Server
	handler greeter.GreeterServiceServer
}

func NewService(meta *common.SntzMeta) *Service {
	return &Service{
		Server:  grpcgw.NewDefault(meta),
		handler: greeterhdl.New(),
	}
}

// New creates a new Greeter module.
func New(srv *Service) *Greeter {

	return &Greeter{
		Service: srv,
	}
}

// Greeter implements GreeterServiceServer.
type Greeter struct {
	*Service
}

// Start implements IGreeter, override runner.Server.Start
func (g *Greeter) Start(ctx context.Context) error {
	greeter.RegisterGreeterServiceServer(g.AsServer(), g.handler)

	appConf := runner.GetAppConfig(ctx)
	return g.Serve(appConf)
}
