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

// Package main provides the entry point for the greeter service.
package main

import (
	"context"

	configspb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/configs/v1"
	"github.com/sentinez/sentinez/cmd/greeter/v1/apps/config"
	"github.com/sentinez/sentinez/internal/core/greeter/v1"
	"github.com/sentinez/sentinez/pkg/runner"
)

func main() {
	app := runner.NewApp(config.Config())
	app.Handle(func(conf *configspb.AppConfig) error {
		grpc := greeter.NewService(conf.GetMeta())

		app.OnStart(grpc.Start)
		app.OnStop(grpc.Shutdown)

		return nil
	})

	runner.Serve(context.Background(), app)
}
