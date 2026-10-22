// Copyright 2026 Duc-Hung Ho.
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

package server

import (
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof" // registers the /debug/pprof handlers
	"time"

	"github.com/sentinez/core/szrunner"
	"github.com/sentinez/shared/zlog"
)

const (
	_pprofAddr          = ":6060"
	_pprofHeaderTimeout = 10 * time.Second
)

// Pprof exposes runtime profiling data at :6060.
//
//	go tool pprof http://localhost:6060/debug/pprof/profile
func Pprof(c *szrunner.Context[http.Server]) {
	c.Inject(func() *http.Server {
		return &http.Server{
			Addr:              _pprofAddr,
			ReadHeaderTimeout: _pprofHeaderTimeout,
		}
	})

	c.Serve(func(_ context.Context, server *http.Server) error {
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			// Profiling is optional: the edge keeps serving without it.
			zlog.Warnf("pprof: %v", err)
		}

		return nil
	})

	c.OnStop(func(ctx context.Context, server *http.Server) error {
		return server.Shutdown(ctx)
	})
}
