// Copyright 2025 Sentinez Labs.
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

package httpx1

import (
	"net/http"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/common/color"
	"github.com/sentinez/sentinez/pkg/core/net/httpx"
	"github.com/sentinez/sentinez/pkg/std/version"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var _ Server = (*HTTPServer)(nil)

type Server interface {
	httpx.Server
	Use(mdw ...func(http.Handler) http.Handler)
	Handle(fn func(ctx Context) error)
	GetRunnerCtx() *common.RunnerCtx
}

func NewServer(runnerCtx *common.RunnerCtx) Server {
	return &HTTPServer{
		rctx: runnerCtx,
	}
}

type HTTPServer struct {
	mdw  []func(http.Handler) http.Handler
	rctx *common.RunnerCtx
}

// GetRunnerCtx implements Server.
func (s *HTTPServer) GetRunnerCtx() *common.RunnerCtx {
	return s.rctx
}

func (s *HTTPServer) Use(mdw ...func(http.Handler) http.Handler) {
	s.mdw = append(s.mdw, mdw...)
}

func (s *HTTPServer) Handle(fn func(ctx Context) error) {
	http.Handle("/", chain(http.HandlerFunc(Convert(fn)), s.mdw...))
}

func (s *HTTPServer) ListenAndServe(addr string) error {
	version.INFO(s.rctx.GetMeta().GetServiceName(),
		s.rctx.GetMeta().GetServiceKey())

	zlog.Infof("%s >>> running on %s",
		color.Blue.Add("http"),
		color.Magenta.Add(addr),
	)

	return http.ListenAndServe(addr, nil)
}

func (s *HTTPServer) Shutdown() error {
	return Shutdown()
}

func chain(h http.Handler, m ...func(http.Handler) http.Handler) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}

	return extendHeader(h)
}

func extendHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		w.Header().Set("Server", version.Name)
	})
}
