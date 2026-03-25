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

	"github.com/sentinez/core/runner"
	"github.com/sentinez/sentinez"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	"github.com/sentinez/sentinez/cmd/edge/v1/apps/config"
	edgeyaml "github.com/sentinez/sentinez/cmd/edge/v1/apps/yaml"
	"github.com/sentinez/sentinez/internal/edge/v1"
	"github.com/sentinez/sentinez/pkg/common/protobuf"
	stdhttpx "github.com/sentinez/sentinez/pkg/network/httpx/std"
	"github.com/sentinez/shared/zlog"

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
	app := runner.NewApp(config.Config(), sentinez.Code)
	app.Run(func(conf *confpb.Config) error {
		var (
			setting    = edgeyaml.LoadSetting(conf.GetFlag().GetProxyConfig())
			httpSrv    = stdhttpx.NewServer(conf.GetMeta())
			edgeServer = edge.New(httpSrv, setting)
		)

		if err := protobuf.Validate(setting); err != nil {
			return err
		}

		// s, _ := jsonx.Marshal(setting)
		// zlog.Debugf("setting: %s", s)

		runner.Register(edgeServer.Start, edgeServer.Shutdown)
		runner.OnStart(func(_ context.Context) error {
			zlog.Debug("[main] start server")
			return nil
		})
		runner.OnStop(func(_ context.Context) error {
			zlog.Debug("[main] stop server")
			return nil
		})

		return nil
	})
}
