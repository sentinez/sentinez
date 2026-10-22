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

package main

import (
	"context"

	"github.com/sentinez/core"
	"github.com/sentinez/core/szrunner"
	"github.com/sentinez/sentinez/pkg/apps/gateway/realtime"
	"github.com/sentinez/sentinez/pkg/apps/gateway/realtime/config"
	wscore "github.com/sentinez/sentinez/pkg/network/wsz"
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

func app(c *szrunner.Context[realtime.Realtime]) {
	c.Inject(wscore.NewServer, realtime.New)

	c.Serve(func(_ context.Context, server *realtime.Realtime) error {
		return server.Start()
	})

	c.OnStop(func(ctx context.Context, server *realtime.Realtime) error {
		return server.Shutdown(ctx)
	})
}
