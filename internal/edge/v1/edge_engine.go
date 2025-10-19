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
	"github.com/sentinez/sentinez/internal/edge/engine/v1"
	grpcgw "github.com/sentinez/sentinez/pkg/network/grpc"
	"github.com/sentinez/sentinez/pkg/runner/v1"
)

func NewEngine(meta *common.SntzMeta) *Engine {
	return &Engine{
		Server: grpcgw.NewDefault(meta),
		Engine: engine.New(),
	}
}

type Engine struct {
	*grpcgw.Server
	*engine.Engine
}

func (e *Engine) Start(ctx context.Context) error {
	edgepb.RegisterEdgeEngineServiceServer(e.AsServer(), e)

	appConf := runner.GetAppConfig(ctx)
	return e.Serve(appConf)
}
