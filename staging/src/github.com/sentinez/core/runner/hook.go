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

	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
	"github.com/sentinez/shared/zlog"
	"go.uber.org/fx"
)

func NewContext[T any](appConf *settingpb.Config) *Context[T] {
	options := []fx.Option{}
	if appConf.GetFlag().GetEnvMode() != "dev" {
		options = append(options, fx.NopLogger)
	}

	return &Context[T]{
		opts: options,
	}
}

type Context[T any] struct {
	opts []fx.Option
}

func (c *Context[T]) OnStart(start func(context.Context, *T) error) {
	function := func(lc fx.Lifecycle, server *T) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					if err := start(ctx, server); err != nil {
						if errors.Is(err, http.ErrServerClosed) {
							zlog.Infof("[runner] %+v", err)
						}
					}
				}()
				return nil
			},
		})
	}
	c.Invoke(function)
}

func (c *Context[T]) OnStop(stop any) {

	switch fn := stop.(type) {
	case func(context.Context, *T) error:
		function := func(lc fx.Lifecycle, server *T) {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					return fn(ctx, server)
				},
			})
		}
		c.Invoke(function)
	case func(context.Context) error:
		function := func(lc fx.Lifecycle, _ *T) {
			lc.Append(fx.Hook{
				OnStop: fn,
			})
		}
		c.Invoke(function)
	}

}

func (c *Context[T]) Invoke(fn any) {
	c.opts = append(c.opts, fx.Invoke(fn))
}

func (c *Context[T]) Inject(fn ...any) {
	c.opts = append(c.opts, fx.Provide(fn...))
}
