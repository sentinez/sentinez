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

	"github.com/sentinez/controlplane/apps/greeter/v1"
	"github.com/sentinez/core"
	"github.com/sentinez/core/szgrpc"
	"github.com/sentinez/core/szrunner"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/sentinez/pkg/apps/controlplane/greeter/config"
	"github.com/sentinez/shared/szlog"
)

func main() {
	if err := run(); err != nil {
		szlog.Fatal(err)
	}
}

func run() error {
	r, err := szrunner.New(config.Config, core.Code)
	if err != nil {
		return err
	}

	return r.Run(context.Background(), szrunner.NewApp(app))
}

func app(c *szrunner.Context[greeter.Greeter]) {
	c.Inject(greeter.NewService)

	c.Serve(func(_ context.Context, server *greeter.Greeter) error {
		conf, err := config.Config()
		if err != nil {
			return err
		}

		return server.Start(szgrpc.WithAddress(conf.GetDefault(
			settingpb.Senz_SENZ_ADDRESS, defaults.GreeterAddress)))
	})

	c.OnStop(func(ctx context.Context, server *greeter.Greeter) error {
		return server.Shutdown(ctx)
	})
}
