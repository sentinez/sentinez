// Copyright 2026 Duc-Hung Ho.
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

// Package main provides the entry point for the collector service.
package main

import (
	"context"

	"github.com/sentinez/controlplane/apps/collector/v1"
	collectorfac "github.com/sentinez/controlplane/apps/collector/v1/factory"
	"github.com/sentinez/controlplane/apps/collector/v1/ingest"
	"github.com/sentinez/core"
	"github.com/sentinez/core/szgrpc"
	"github.com/sentinez/core/szrunner"
	"github.com/sentinez/pkg/szlog"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/sentinez/pkg/apps/controlplane/collector/config"
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

	return r.Run(context.Background(),
		szrunner.NewApp(app),
		szrunner.NewApp(ingestApp),
	)
}

func app(c *szrunner.Context[collector.Collector]) {
	c.Inject(collector.NewService)

	c.Serve(func(_ context.Context, server *collector.Collector) error {
		conf, err := config.Config()
		if err != nil {
			return err
		}

		return server.Start(szgrpc.WithAddress(conf.GetDefault(
			settingpb.Senz_SENZ_ADDRESS, defaults.CollectorAddress)))
	})

	c.OnStop(func(ctx context.Context, server *collector.Collector) error {
		return server.Shutdown(ctx)
	})
}

// ingestApp inserts the edge events of kafka in the analytic tables.
func ingestApp(c *szrunner.Context[ingest.Ingest]) {
	c.Inject(collectorfac.NewDefaultIngest)

	c.Serve(func(ctx context.Context, in *ingest.Ingest) error {
		return in.Run(ctx)
	})

	c.OnStop(func(_ context.Context, in *ingest.Ingest) error {
		in.Close()
		return nil
	})
}
