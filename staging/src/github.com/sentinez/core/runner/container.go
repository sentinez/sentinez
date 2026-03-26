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

	"go.uber.org/fx"
)

// container represents the container with uber/fx frameworks.
// manage the lifecycle of the application.
type container struct {
	engine *fx.App
}

// Run the app with the given context.
func (ctn *container) Run(ctx context.Context) error {
	err := make(chan error)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// defer close(err)
	// defer close(sig)

	// fork the goroutine 1 for start the app
	go ctn.onStart(ctx, err)

	// fork the goroutine 2 for stop the app
	go ctn.onStop(ctx, sig, err)

	// wait for the error from the goroutine 1 or 2, end the app
	return <-err
}

// onStart the app with the given context.
func (ctn *container) onStart(ctx context.Context, errChan chan<- error) {

	// if the error is not nil, return the error to err channel end goroutine 1
	if err := ctn.engine.Start(ctx); err != nil {
		errChan <- err
	}
}

// onStop the app with the given context.
func (ctn *container) onStop(
	ctx context.Context, sigChan <-chan os.Signal, errChan chan<- error) {

	// wait for the signal interrupt from the OS
	<-sigChan

	// if the error is not nil, return the error to err channel, end goroutine 2
	if err := ctn.engine.Stop(ctx); err != nil {
		errChan <- err
	}

	// if stop the app successfully, return nil to err channel, end goroutine 2
	errChan <- nil
}
