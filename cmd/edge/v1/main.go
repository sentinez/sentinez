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

	"github.com/sentinez/sentinez/cmd/edge/v1/apps/config"
	edgeyaml "github.com/sentinez/sentinez/cmd/edge/v1/apps/yaml"
	"github.com/sentinez/sentinez/internal/edge/v1"
	httpxdmz "github.com/sentinez/sentinez/pkg/dmz/httpx"
	"github.com/sentinez/sentinez/pkg/runner"

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
//   go tool pprof http://localhost:6060/debug/pprof/profile
//
// func init() {
// 	go func() {
// 		_ = http.ListenAndServe(":6060", nil)
// 	}()
// }

// main is the entrypoint of the Edge application.
// It initializes configuration, creates the HTTP server and Edge Engine,
// and registers their start/stop hooks with the runner framework.
func main() {
	// runner.Main handles the service lifecycle, including initialization,
	// start, graceful shutdown, and signal handling.
	runner.Main(config.Config(), func(ctx context.Context) error {
		// Retrieve the current application configuration from context.
		conf := runner.GetAppConfig(ctx)

		// Load YAML-based proxy
		// configuration (routes, backends, policies, etc.).
		setting := edgeyaml.LoadSetting(conf.GetFlag().GetProxyConfig())

		// Initialize the DMZ HTTP server.
		// This server is the public-facing entrypoint that forwards requests
		// to the internal Edge Engine for processing.
		httpSrv := httpxdmz.NewServer(conf.GetMeta())

		// Create the Edge server
		// instance using the HTTP layer and proxy settings.
		edgeServer := edge.New(httpSrv, setting)

		// Register startup and shutdown hooks for the Edge HTTP server.
		runner.OnStart(edgeServer.Start)
		runner.OnStop(edgeServer.Shutdown)

		// Initialize the Edge Engine (gRPC handler) responsible for
		// handling internal communication, WAF logic, and routing control.
		// engine := edge.NewEngine(conf.GetMeta())

		// Register startup and shutdown hooks for the Edge Engine.
		// runner.OnStart(engine.Start)
		// runner.OnStop(engine.Shutdown)

		return nil
	})
}
