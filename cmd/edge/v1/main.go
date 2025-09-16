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

	edgev1 "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	edgeflags "github.com/sentinez/sentinez/cmd/edge/v1/apps/flags"
	edgeyaml "github.com/sentinez/sentinez/cmd/edge/v1/apps/yaml"
	"github.com/sentinez/sentinez/internal/edge/v1"
	httpxf1 "github.com/sentinez/sentinez/pkg/core/net/httpx/f1"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/std/config"

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
	flag := edgeflags.Parse()
	conf := config.Load(flag.GetEnvFile())

	yamlconf := edgeyaml.LoadRoutesFromYAML(flag.GetProxyConfig())

	rctx := &common.RunnerCtx{
		Meta:   edgev1.GetMetaEdge(),
		Config: conf,
		Flag:   flag,
	}

	runner.Main(func(_ context.Context) error {
		httpSrv := httpxf1.NewServer(edgev1.GetMetaEdge())

		edgeServer := edge.New(httpSrv, yamlconf)

		runner.OnStart(func(ctx context.Context) error {
			return edgeServer.Start(ctx)
		})

		runner.OnStop(func(ctx context.Context) error {
			return edgeServer.Shutdown(ctx)
		})

		return nil
	}, runner.WithContextValue(rctx))
}
