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

	"github.com/sentinez/core/szrunner"
	"github.com/sentinez/sentinez/pkg/apps/dmz/edge"
	"github.com/sentinez/sentinez/pkg/cluster/leader"
)

// Edge serves the traffic and campaigns for the leadership of the zone:
// the leader keeps the zone Status in Consul and follows centraldata for
// the cluster (edge.Server.Lead). The elector lives in this app, as the
// tasks of the leader need the stores of the edge.
func Edge(c *szrunner.Context[edge.Server]) {
	c.Inject(edge.NewServer, leader.New)

	c.Serve(func(_ context.Context, server *edge.Server) error {
		return server.Start()
	})

	c.Serve(func(ctx context.Context, server *edge.Server) error {
		return server.Lead(ctx)
	})

	c.OnStop(func(ctx context.Context, server *edge.Server) error {
		return server.Shutdown(ctx)
	})
}

func GRPC(c *szrunner.Context[edge.Service]) {
	c.Inject(edge.NewService)

	c.Serve(func(_ context.Context, server *edge.Service) error {
		return server.Start()
	})

	c.OnStop(func(ctx context.Context, server *edge.Service) error {
		return server.Shutdown(ctx)
	})
}
