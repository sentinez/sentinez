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

// Package szrunner runs one or more apps in the same process and manages
// their lifecycle: start, serve, and graceful shutdown on SIGINT/SIGTERM
// or when any app fails.
package szrunner

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"slices"
	"syscall"
	"time"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/szlog"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/grpclog"
)

const _stopTimeout = 15 * time.Second

// Runner runs a set of Apps sharing the process-wide logging and OTLP
// setup.
type Runner struct {
	opts    []fx.Option
	closers []func(context.Context) error
}

// New sets up logging and OTLP export for the process.
func New(confConstr func() (*settingpb.Config, error),
	scopeName string) (*Runner, error) {
	appConf, err := confConstr()
	if err != nil {
		return nil, err
	}

	logging := szlog.NewConsole(scopeName, szlog.LevelError)
	grpclog.SetLoggerV2(logging)

	level := szlog.ToLevel(appConf.GetFlag().GetLogLevel())
	szlog.SetScopeLogLevel(scopeName, level)

	r := &Runner{}
	if appConf.GetFlag().GetEnvMode() != "dev" {
		r.opts = append(r.opts, fx.NopLogger)
	}

	r.opts = append(r.opts, fx.Provide(confConstr))

	shutdown, err := setupOTLP(appConf)
	if err != nil {
		return nil, err
	}

	r.closers = append(r.closers, shutdown)

	return r, nil
}

// Run starts apps in order, serves them until ctx is done, a signal is
// received or a Serve fails, then stops them in reverse order.
func (r *Runner) Run(ctx context.Context, apps ...*App) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	started, err := r.start(ctx, apps)
	if err == nil {
		err = serve(ctx, apps, func() error { return stop(started) })
	} else {
		err = errors.Join(err, stop(started))
	}

	return errors.Join(err, r.close())
}

func (r *Runner) start(ctx context.Context, apps []*App) ([]*fx.App, error) {
	started := make([]*fx.App, 0, len(apps))

	for i, app := range apps {
		fxApp := fx.New(append(r.opts[:len(r.opts):len(r.opts)],
			app.opts...)...)
		if err := fxApp.Start(ctx); err != nil {
			return started, fmt.Errorf("runner: start app %d: %w", i, err)
		}

		started = append(started, fxApp)
	}

	return started, nil
}

// serve runs every Serve loop, calls stop once ctx is done or a loop
// fails, then waits for the loops to return.
func serve(ctx context.Context, apps []*App, stop func() error) error {
	g, gctx := errgroup.WithContext(ctx)

	for _, app := range apps {
		for _, fn := range app.serves {
			g.Go(func() error {
				if err := fn(gctx); !errors.Is(err, http.ErrServerClosed) {
					return err
				}

				return nil
			})
		}
	}

	<-gctx.Done()
	szlog.Infof("[runner] shutting down")

	stopErr := stop()

	return errors.Join(g.Wait(), stopErr)
}

func stop(apps []*fx.App) error {
	ctx, cancel := context.WithTimeout(context.Background(), _stopTimeout)
	defer cancel()

	var errs []error
	for i, app := range slices.Backward(apps) {
		if err := app.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("runner: stop app %d: %w", i, err))
		}
	}

	return errors.Join(errs...)
}

func (r *Runner) close() error {
	ctx, cancel := context.WithTimeout(context.Background(), _stopTimeout)
	defer cancel()

	var errs []error
	for _, closeFn := range r.closers {
		errs = append(errs, closeFn(ctx))
	}

	return errors.Join(errs...)
}

// setupOTLP exports the system logs to SENZ_OTLP_INTERNAL_ENDPOINT. When
// it is unset nothing is exported and the returned closer does nothing.
func setupOTLP(appConf *settingpb.Config,
) (func(context.Context) error, error) {
	endpoint := appConf.Get(settingpb.Senz_SENZ_OTLP_INTERNAL_ENDPOINT)
	if endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}

	shutdown, err := szlog.SetupOTLP(context.Background(), szlog.OTLPConfigOf(
		endpoint, appConf.GetMeta().GetServiceKey()))
	if err != nil {
		return nil, fmt.Errorf("runner: setup otlp: %w", err)
	}

	// Export fails without a collector; only ensure shutdown does not hang.
	return shutdown, nil
}
