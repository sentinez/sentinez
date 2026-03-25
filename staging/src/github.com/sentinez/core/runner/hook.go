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
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
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
	conf *confpb.Config
}

func (a *App) Run(start func(conf *confpb.Config) error) {
	start(a.conf)
	serve(context.Background(), a)
}

func OnStart(start any) {
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

func OnStop(stop func(ctx context.Context) error) {
	function := func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStop: stop,
		})
	}

	internal.Invoke(function)
}

// Register adds a paired OnStart + OnStop lifecycle hook within a single fx.Hook.
// This ensures that OnStop is always called by fx during shutdown,
// because fx only invokes OnStop for hooks whose OnStart has run.
//
// Use this instead of calling OnStart and OnStop separately.
func Register(
	start any,
	stop func(ctx context.Context) error,
) {
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
						}
					}()
					return nil
				},
				OnStop: stop,
			})
		}
		internal.Invoke(function)
	default:
		// For fx-injectable start functions (e.g. func(conf *confpb.Config) error),
		// we wrap them in an lc.Append so they run in a goroutine during OnStart
		// and OnStop is paired in the same fx.Hook — ensuring fx calls it on shutdown.
		function := func(lc fx.Lifecycle, conf *confpb.Config) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					if typedFn, ok := fn.(func(*confpb.Config) error); ok {
						go func() {
							if err := typedFn(conf); err != nil {
								if errors.Is(err, http.ErrServerClosed) {
									zlog.Infof("[runner] %+v", err)
								}
							}
						}()
					}
					return nil
				},
				OnStop: stop,
			})
		}
		internal.Invoke(function)
	}
}

func Invoke(fn any) {
	internal.Invoke(fn)
}

func Inject(fn ...any) {
	internal.Provide(fn...)
}
