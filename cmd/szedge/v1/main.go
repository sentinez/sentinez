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
	"github.com/sentinez/core/szrunner"
	"github.com/sentinez/sentinez/cmd/szedge/v1/server"
	"github.com/sentinez/sentinez/pkg/apps/dmz/edge/config"
	"github.com/sentinez/shared/zlog"
)

// main is the entrypoint of the Edge application.
// It initializes configuration, creates the HTTP server and Edge Engine,
// and registers their start/stop hooks with the runner framework.
func main() {
	if err := run(); err != nil {
		zlog.Fatal(err)
	}
}

func run() error {
	r, err := szrunner.New(config.Config, core.Code)
	if err != nil {
		return err
	}

	return r.Run(context.Background(),
		szrunner.NewApp(server.Edge),
		szrunner.NewApp(server.GRPC),
		szrunner.NewApp(server.Pprof),
	)
}
