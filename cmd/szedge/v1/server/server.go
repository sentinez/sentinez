// Copyright 2026 Duc-Hung Ho.
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

package server

import (
	"context"

	"github.com/sentinez/core/runner"
	"github.com/sentinez/sentinez/pkg/apps/controlplane/greeter/config"
	"github.com/sentinez/sentinez/pkg/apps/dmz/edge"
	edgeyaml "github.com/sentinez/sentinez/pkg/apps/dmz/edge/yaml"
)

func Edge(c *runner.Context[edge.Server]) {
	c.Inject(
		config.Config,
		edgeyaml.LoadSetting,
		edge.NewServer,
	)

	c.Serve(func(_ context.Context, server *edge.Server) error {
		return server.Start()
	})

	c.OnStop(func(ctx context.Context, server *edge.Server) error {
		return server.Shutdown(ctx)
	})
}

func GRPC(c *runner.Context[edge.Service]) {
	c.Inject(
		config.Config,
		edge.NewService,
	)

	c.Serve(func(_ context.Context, server *edge.Service) error {
		return server.Start()
	})

	c.OnStop(func(ctx context.Context, server *edge.Service) error {
		return server.Shutdown(ctx)
	})
}
