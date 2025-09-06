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

package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/core/runner/v1/internal"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"google.golang.org/grpc/grpclog"

	"go.uber.org/fx"
)

var (
	onceWhenStart    sync.Once
	onceWhenShutdown sync.Once

	logging zlog.Sugard

	options map[OptionType]any

	start func(context.Context) error
	stop  func(context.Context) error
)

// runner functions called by fx.Invoke.
// when the application starts, it will start the server
// nolint:funlen
func runner(lc fx.Lifecycle) error {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			errChan := make(chan error, 1)
			go func() {
				if start == nil {
					errChan <- fmt.Errorf("[runner]: missing start function")
					return
				}
				if err := start(ctx); err != nil {
					if errors.Is(err, errors.ErrServerClosed) {
						logging.Infof("[runner] %+v", err)
					} else {
						logging.Errorf("[runner] %+v", err)
					}

					errChan <- err
				}
			}()

			select {
			case err := <-errChan:
				return err
			default:
				return nil
			}
		},
		OnStop: func(ctx context.Context) error {
			_ = logging.Sync()
			if stop == nil {
				return nil
			}

			return stop(ctx)
		},
	})

	return nil
}

func Main(fn func(ctx context.Context) error, opts ...Option) {
	onceWhenStart.Do(func() {
		logging = zlog.NewDefaultConsole(zlog.LevelError)

		grpclog.SetLoggerV2(logging)

		start = fn

		options = make(map[OptionType]any)
		for _, opt := range opts {
			options[opt.Type()] = opt.Value()
		}
	})

	ctn := container{}

	opt, ok := options[RunnerCtx]
	if !ok || opt == nil {
		zlog.Fatal("runner: missing runner context value")
	}

	rctx := opt.(*common.RunnerCtx)
	zlog.SetLogLevel(rctx.GetFlag().GetLogLevel())

	// disable log: use fx.NopLogger
	if rctx.GetFlag().GetEnvMode() != "dev" {
		ctn.engine = fx.New(internal.Option(), fx.Invoke(runner), fx.NopLogger)
	} else {
		ctn.engine = fx.New(internal.Option(), fx.Invoke(runner))
	}

	_ = ctn.Run(newContext(rctx))
}

func Shutdown(fn func(ctx context.Context) error) {
	onceWhenShutdown.Do(func() {
		stop = fn
	})
}
