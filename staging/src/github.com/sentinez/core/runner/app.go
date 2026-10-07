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

	"go.uber.org/fx"
)

// App is one service of a Runner. Each App has its own dependency
// container, so apps in the same process can provide the same types.
type App struct {
	opts   []fx.Option
	serves []func(context.Context) error
}

// NewApp builds an App whose main component is *T; setup registers its
// providers and lifecycle hooks on the given Context.
func NewApp[T any](setup ...func(*Context[T])) *App {
	app := &App{}
	ctx := &Context[T]{app: app}

	for _, fn := range setup {
		fn(ctx)
	}

	return app
}

// Context registers the providers and lifecycle hooks of an App whose
// main component is *T.
type Context[T any] struct {
	app *App
}

// Inject registers constructors in the App container.
func (c *Context[T]) Inject(constructors ...any) {
	c.app.opts = append(c.app.opts, fx.Provide(constructors...))
}

// Invoke registers a function called with its dependencies when the App
// is built.
func (c *Context[T]) Invoke(fn any) {
	c.app.opts = append(c.app.opts, fx.Invoke(fn))
}

// OnStart registers a hook run before Serve. Hooks run in registration
// order and must not block.
func (c *Context[T]) OnStart(fn func(context.Context, *T) error) {
	c.Invoke(func(lc fx.Lifecycle, t *T) {
		lc.Append(fx.StartHook(func(ctx context.Context) error {
			return fn(ctx, t)
		}))
	})
}

// Serve registers the blocking main loop of the App. It runs once every
// App has started; returning an error other than http.ErrServerClosed
// stops the whole Runner.
func (c *Context[T]) Serve(fn func(context.Context, *T) error) {
	c.Invoke(func(t *T) {
		c.app.serves = append(c.app.serves, func(ctx context.Context) error {
			return fn(ctx, t)
		})
	})
}

// OnStop registers a hook run on shutdown, in reverse registration order.
func (c *Context[T]) OnStop(fn func(context.Context, *T) error) {
	c.Invoke(func(lc fx.Lifecycle, t *T) {
		lc.Append(fx.StopHook(func(ctx context.Context) error {
			return fn(ctx, t)
		}))
	})
}
