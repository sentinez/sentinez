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

package edge

import (
	"context"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	configspb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/configs/v1"
	"github.com/sentinez/sentinez/internal/edge/v1/engine"
	grpc "github.com/sentinez/sentinez/pkg/network/grpc"
)

//
// Package edge implements the core Edge Server component.
//
// The Edge Engine provides remote procedure call (RPC) capabilities over gRPC,
// allowing other services to connect to edge layer, execute remote functions
// and interact with its internal processing pipelines.
//
// Togeter with the Edge Server (HTTP entrypoint), these components form
// the foundation of the edge platform - responsible for traffic handling,
// routing, proxing, policy enforcement.
//
// This package integrates tightly with the `runner` package to ensure
// controlled startup, graceful shutdown, and consistent lifecycle management
// across the entire edge system.
//

func NewEngine(meta *common.XMeta) *Engine {
	return &Engine{
		Server: grpc.NewDefault(meta),
		Engine: engine.New(),
	}
}

type Engine struct {
	*grpc.Server
	*engine.Engine
}

func (e *Engine) Start(_ context.Context, conf *configspb.AppConfig) error {
	edgepb.RegisterEdgeEngineServiceServer(e.AsServer(), e)

	return e.Serve(conf)
}
