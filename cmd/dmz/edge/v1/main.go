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

// Package main provides the entry point for the Edge Service.
package main

import (
	"context"

	"github.com/sentinez/core"
	"github.com/sentinez/core/runner"
	edge "github.com/sentinez/sentinez/pkg/dmz/edge"
	"github.com/sentinez/sentinez/pkg/dmz/edge/apps/config"
	edgeyaml "github.com/sentinez/sentinez/pkg/dmz/edge/apps/yaml"
	stdhttpx "github.com/sentinez/sentinez/pkg/network/httpx/std"
	stdproxy "github.com/sentinez/sentinez/pkg/network/httpx/std/proxy"

	"net/http"
	_ "net/http/pprof"
)

//
// The main package is the entrypoint for the Sentinez Edge Proxy service.
// It initializes the DMZ HTTP proxy server, the Edge Engine (gRPC handler),
// and manages their lifecycles using the internal runner framework.
//

// func init enables the pprof HTTP server for profiling purposes.
// Uncomment this block to expose runtime profiling data at :6060.
//
// Example:
//
//	go tool pprof http://localhost:6060/debug/pprof/profile
func init() {
	go func() {
		_ = http.ListenAndServe(":6060", nil)
	}()
}

// main is the entrypoint of the Edge application.
// It initializes configuration, creates the HTTP server and Edge Engine,
// and registers their start/stop hooks with the runner framework.
func main() {
	app := runner.NewApp[edge.Server](config.Config(), core.Code)
	app.Main(func(c *runner.Context[*edge.Server]) {
		c.Inject(
			config.Config,
			edgeyaml.LoadSetting,
			stdhttpx.NewServer,
			edge.New,
		)

		c.OnStart(func(_ context.Context, server *edge.Server) error {
			server.SetReverseProxyConstructor(stdproxy.NewReverseProxy)

			return server.Start()
		})

		c.OnStop(func(ctx context.Context, server *edge.Server) error {
			return server.Shutdown(ctx)
		})
	})
}
