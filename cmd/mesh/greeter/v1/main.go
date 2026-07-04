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

	"github.com/sentinez/core"
	"github.com/sentinez/core/runner"
	"github.com/sentinez/modules/greeter/v1"
	"github.com/sentinez/sentinez/pkg/mesh/greeter/apps/config"
)

func main() {
	app := runner.NewApp[greeter.Greeter](config.Config(), core.Code)
	app.Main(func(c *runner.Context[*greeter.Greeter]) {
		c.Inject(config.Config, greeter.NewService)

		c.OnStart(func(_ context.Context, server *greeter.Greeter) error {
			return server.Start()
		})

		c.OnStop(func(ctx context.Context, server *greeter.Greeter) error {
			return server.Shutdown(ctx)
		})
	})
}
