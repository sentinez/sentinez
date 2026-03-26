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

	options := []fx.Option{}
	if appConf.GetFlag().GetEnvMode() != "dev" {
		options = append(options, fx.NopLogger)
	}

	return &App{
		conf: appConf,
		opts: options,
	}
}

type App struct {
	conf *confpb.Config
	opts []fx.Option
}

func (a *App) Run(ctx context.Context) {
	a.Inject(func() *confpb.Config {
		return a.conf
	})

	ctn := container{engine: fx.New(a.opts...)}
	if err := ctn.Run(ctx); err != nil {
		zlog.Fatal(err)
	}
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
						}
					}()
					return nil
				},
			})
		}
		a.Invoke(function)
	default:
		a.Invoke(start)
	}
}

func (a *App) OnStop(stop func(ctx context.Context) error) {
	function := func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStop: stop,
		})
	}
	a.Invoke(function)
}

// nolint
// Register adds a paired OnStart + OnStop lifecycle hook within a single fx.Hook.
func (a *App) Register(start any, stop func(ctx context.Context) error) {
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
		a.Invoke(function)
	default:
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
		a.Invoke(function)
	}
}

func (a *App) Invoke(fn any) {
	a.opts = append(a.opts, fx.Invoke(fn))
}

func (a *App) Inject(fn ...any) {
	a.opts = append(a.opts, fx.Provide(fn...))
}
