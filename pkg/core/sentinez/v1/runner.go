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

package sentinez

import (
	"context"
	"time"

	"github.com/sentinez/sentinez/pkg/common/protobuf"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/sentinez/sentinez/pkg/std/zlog"

	"go.uber.org/fx"
)

const timeout = 500 * time.Millisecond // 500 milliseconds

// runner functions called by fx.Invoke.
// when the application starts, it will start the server
func runner(lc fx.Lifecycle, srv Server) {
	// init log
	log := zlog.NewSystemLog()

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			errChan := make(chan error, 1)
			go func() {
				if err := srv.Start(ctx); err != nil {
					if errors.Is(err, errors.ErrServerClosed) {
						log.Infof("[runner] %+v", err)
					} else {
						log.Errorf("[runner] %+v", err)
					}

					errChan <- err
				}
			}()

			select {
			case err := <-errChan:
				return err
			case <-time.After(timeout):
				return protobuf.Validate(flags.Parse())
			}

		},
		OnStop: func(ctx context.Context) error {
			_ = log.Sync()
			return srv.Shutdown(ctx)
		},
	})
}
