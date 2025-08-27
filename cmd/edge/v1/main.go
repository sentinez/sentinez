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
	edgeflags "github.com/sentinez/sentinez/cmd/edge/v1/apps/flags"
	edgeyaml "github.com/sentinez/sentinez/cmd/edge/v1/apps/yaml"
	"github.com/sentinez/sentinez/internal/edge/v1"
	httpxf1 "github.com/sentinez/sentinez/pkg/core/net/httpx/f1"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/sentinez/sentinez/pkg/std/zlog"

	_ "net/http/pprof"
)

// expose pprof
//
// func init() {
// 	go func() {
// 		_ = http.ListenAndServe(":6060", nil)
// 	}()
// }

func loadYaml(confPath string) *edgeyaml.Config {
	return edgeyaml.LoadRoutesFromYAML(confPath)
}

func main() {
	if err := flags.Validate(edgeflags.ParseFlag()); err != nil {
		zlog.Fatal(err)
	}

	proxyConf := loadYaml(edgeflags.ParseFlag().GetProxyConfig())

	app := runner.New(httpxf1.NewHTTPServer).
		Build(func(srv *httpxf1.HTTPServer) (runner.Server, error) {
			srv.Metadata = edgev1.Metadata_edge
			return edge.New(srv, edgeflags.ParseFlag(), proxyConf), nil
		})

	_ = app.Run(context.Background())
}
