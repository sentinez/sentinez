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

	coregrpc "github.com/sentinez/core/grpc"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
	edgeapi "github.com/sentinez/sentinez/internal/dmz/edge/api"
)

//
// Package edge implements the core Edge Server component.
//
// The Edge Engine provides remote procedure call (RPC) capabilities over gRPC,
// allowing other services to connect to edge layer, execute remote functions
// and interact with its internal processing pipelines.
//
// Together with the Edge Server (HTTP entrypoint), these components form
// the foundation of the edge platform - responsible for traffic handling,
// routing, proxying, policy enforcement.
//
// This package integrates tightly with the `runner` package to ensure
// controlled startup, graceful shutdown, and consistent lifecycle management
// across the entire edge system.
//

func NewService(conf *settingpb.Config) *Service {
	return &Service{
		Server:      coregrpc.New(coregrpc.WithXMeta(conf.GetMeta())),
		EdgeService: edgeapi.New(),
	}
}

type Service struct {
	*coregrpc.Server
	*edgeapi.EdgeService
}

func (e *Service) Start(_ context.Context, conf *settingpb.Config) error {
	edgepb.RegisterEdgeServiceServer(e.AsServer(), e)

	return e.Serve(conf)
}
