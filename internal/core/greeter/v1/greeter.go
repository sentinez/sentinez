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
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	configspb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/configs/v1"
	greeterhdl "github.com/sentinez/sentinez/internal/core/greeter/v1/handler"
	netgrpc "github.com/sentinez/sentinez/pkg/network/grpc"
)

func NewService(meta *common.XMeta) *Greeter {
	return &Greeter{
		Server:  netgrpc.NewDefault(meta),
		handler: greeterhdl.New(),
	}
}

// Greeter implements GreeterServiceServer.
type Greeter struct {
	*netgrpc.Server
	handler greeter.GreeterServiceServer
}

func (g *Greeter) Start(conf *configspb.AppConfig) error {
	greeter.RegisterGreeterServiceServer(g.AsServer(), g.handler)

	return g.Serve(conf)
}
