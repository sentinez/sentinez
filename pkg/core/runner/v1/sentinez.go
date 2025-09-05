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
	"os"
	"os/signal"
	"syscall"

	"github.com/sentinez/sentinez/pkg/core/runner/v1/internal"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"go.uber.org/fx"
)

// Runner represents the application when all constructor was build
// by runner.Build() start the app, it will start the server and provide all
// constructor needed
type Runner[srv any] interface {
	Build(start func(srv) (Engine, error)) Runner[srv]
	Run(ctx context.Context) error
}

func New[srv any](fn func(ctx context.Context) srv) Runner[srv] {
	internal.Provide(fn)
	return &sentinez[srv]{}
}

// sentinez represents the container with uber/fx frameworks.
// manage the lifecycle of the application.
type sentinez[srv any] struct {
	engine *fx.App
}

// Build builds the application.
// The application is built by providing the constructors.
func (s *sentinez[srv]) Build(start func(srv) (Engine, error)) Runner[srv] {

	internal.Provide(start)
	return &sentinez[srv]{}
}

// Run the app with the given context.
func (s *sentinez[srv]) Run(ctx context.Context) error {
	runnerCtx := GetContext(ctx)
	runnerCtxConstructor := func() context.Context {
		return ctx
	}

	zlog.SetLogLevel(runnerCtx.Flag.GetLogLevel())
	internal.Provide(runnerCtxConstructor)

	// disable log: use fx.NopLogger
	if runnerCtx.Flag.GetEnvMode() != "dev" {
		s.engine = fx.New(internal.Option(), fx.Invoke(runner), fx.NopLogger)
	} else {
		s.engine = fx.New(internal.Option(), fx.Invoke(runner))
	}

	err := make(chan error)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// defer close(err)
	// defer close(sig)

	// fork the goroutine 1 for start the app
	go s.onStart(ctx, err)

	// fork the goroutine 2 for stop the app
	go s.onStop(ctx, sig, err)

	// show memory usage
	// if flags.Get().LogLevel == zlog.LevelDebug.String() {
	// 	time.AfterFunc(1*time.Second, memory.PrintUsage)
	// }

	// wait for the error from the goroutine 1 or 2, end the app
	return <-err
}

// onStart the app with the given context.
func (s *sentinez[srv]) onStart(ctx context.Context, errChan chan<- error) {
	if s.engine == nil {
		s.engine = fx.New(internal.Option())
	}

	// if the error is not nil, return the error to err channel end goroutine 1
	if err := s.engine.Start(ctx); err != nil {
		errChan <- err
	}
}

// onStop the app with the given context.
func (s *sentinez[srv]) onStop(
	ctx context.Context, sigChan <-chan os.Signal, errChan chan<- error) {

	// wait for the signal interrupt from the OS
	<-sigChan

	// if the error is not nil, return the error to err channel, end goroutine 2
	if err := s.engine.Stop(ctx); err != nil {
		errChan <- err
	}

	// if stop the app successfully, return nil to err channel, end goroutine 2
	errChan <- nil
}
