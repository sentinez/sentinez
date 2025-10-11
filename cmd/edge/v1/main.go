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
	httpxhz "github.com/sentinez/sentinez/pkg/network/httpx/hz"
	"github.com/sentinez/sentinez/pkg/runner/v1"

	_ "net/http/pprof"
)

// expose pprof
//
// func init() {
// 	go func() {
// 		_ = http.ListenAndServe(":6060", nil)
// 	}()
// }

func main() {
	runner.Main(config.Config(), func(ctx context.Context) error {
		conf := runner.GetAppConfig(ctx)
		yamlconf := edgeyaml.LoadRouteConfig(conf.GetFlag().GetProxyConfig())

		httpSrv := httpxhz.NewServer(conf.GetMeta())
		edgeServer := edge.New(httpSrv, yamlconf)

		runner.OnStart(edgeServer.Start)
		runner.OnStop(edgeServer.Shutdown)

		return nil
	})
}
