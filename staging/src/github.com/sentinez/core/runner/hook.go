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

// Package runner provides a way to add hooks to the application lifecycle.
package runner

import (
	"context"
	"errors"
	"net/http"

	"github.com/sentinez/core/runner/internal"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/shared/zlog"
	"go.uber.org/fx"
	"google.golang.org/grpc/grpclog"
)

func NewApp(appConf *confpb.Config, scopeName string) *App {
	logging := zlog.NewConsole(scopeName, zlog.LevelError)
	grpclog.SetLoggerV2(logging)

	level := zlog.ToLevel(appConf.GetFlag().GetLogLevel())
	zlog.SetScopeLogLevel(scopeName, level)

	if appConf.GetFlag().GetEnvMode() != "dev" {
		internal.AppendOption(fx.NopLogger)
	}

	return &App{conf: appConf}
}

type App struct {
	conf  *confpb.Config
	start func(conf *confpb.Config) error
}

func (a *App) Handle(start func(conf *confpb.Config) error) {
	a.start = start
}

func (a *App) OnStart(start any) {
	switch fn := start.(type) {
	case func(context.Context) error:
		function := func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {

					go func() {
						if err := fn(ctx); err != nil {
							if errors.Is(err, http.ErrServerClosed) {
								zlog.Infof("[runner] %+v", err)
							}
							//else {
							//	logging.Fatalf("[runner] %+v", err)
							//}
						}
					}()

					return nil
				},
			})
		}
		internal.Invoke(function)
	default:
		internal.Invoke(start)
	}
}

func (a *App) OnStop(stop func(ctx context.Context) error) {
	function := func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStop: stop,
		})
	}

	internal.Invoke(function)
}

func (a *App) Invoke(fn any) {
	internal.Invoke(fn)
}

func (a *App) Inject(fn ...any) {
	internal.Provide(fn...)
}
