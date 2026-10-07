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

// Package runner runs one or more apps in the same process and manages
// their lifecycle: start, serve, and graceful shutdown on SIGINT/SIGTERM
// or when any app fails.
package runner

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	httpconst "github.com/sentinez/core/http/const"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/zlog"
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
func New(appConf *settingpb.Config, scopeName string) *Runner {
	logging := zlog.NewConsole(scopeName, zlog.LevelError)
	grpclog.SetLoggerV2(logging)

	level := zlog.ToLevel(appConf.GetFlag().GetLogLevel())
	zlog.SetScopeLogLevel(scopeName, level)

	r := &Runner{}
	if appConf.GetFlag().GetEnvMode() != "dev" {
		r.opts = append(r.opts, fx.NopLogger)
	}

	r.closers = append(r.closers, setupOTLP(appConf))

	return r
}

// Main runs apps until a signal or a failure, and exits on error. It is
// meant to be the whole body of main().
func (r *Runner) Main(apps ...*App) {
	if err := r.Run(context.Background(), apps...); err != nil {
		zlog.Fatal(err)
	}
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
	zlog.Infof("[runner] shutting down")

	stopErr := stop()

	return errors.Join(g.Wait(), stopErr)
}

func stop(apps []*fx.App) error {
	ctx, cancel := context.WithTimeout(context.Background(), _stopTimeout)
	defer cancel()

	var errs []error
	for i := len(apps) - 1; i >= 0; i-- {
		if err := apps[i].Stop(ctx); err != nil {
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

func setupOTLP(appConf *settingpb.Config) func(context.Context) error {
	inSecure := true
	endpoint := appConf.GetDefault(
		settingpb.Senz_SENZ_OTLP_ENDPOINT, "localhost:4317")

	if strings.HasPrefix(endpoint, httpconst.SchemeSecure) {
		inSecure = false
	}

	shutdown, err := zlog.SetupOTLP(context.Background(), zlog.OTLPConfig{
		Endpoint:    endpoint,
		Insecure:    inSecure,
		ServiceName: appConf.GetMeta().GetServiceKey(),
	})
	if err != nil {
		zlog.Fatalf("runner.setupOTLP: %v", err)
	}

	// Export fails without a collector; only ensure shutdown does not hang.
	return shutdown
}
