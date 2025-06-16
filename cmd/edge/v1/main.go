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

	edgeconfig "github.com/sentinez/sentinez/cmd/edge/v1/apps/config"
	edgeflags "github.com/sentinez/sentinez/cmd/edge/v1/apps/flags"

	"github.com/sentinez/sentinez/internal/edge/v1"
	"github.com/sentinez/sentinez/pkg/core/sentinez/v1"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func loadYaml() *edgeconfig.Routes {
	return edgeconfig.LoadRoutesFromYAML("./cmd/edge/v1/proxy.yaml")
}

func main() {
	if err := flags.Validate(edgeflags.ParseFlag()); err != nil {
		zlog.Fatal(err)
	}

	app := sentinez.Build(edge.New, edgeflags.ParseFlag, loadYaml)
	if err := app.Run(context.Background()); err != nil {
		zlog.Fatal(err)
	}
}
