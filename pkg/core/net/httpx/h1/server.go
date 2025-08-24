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
}

func NewServer() Server {
	return &HTTPServer{}
}

type HTTPServer struct {
	mdw      []func(http.Handler) http.Handler
	Metadata *common.SentinezMetadata
}

func (s *HTTPServer) uses(h http.Handler,
	middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func (s *HTTPServer) Use(mdw ...func(http.Handler) http.Handler) {
	s.mdw = append(s.mdw, mdw...)
}

func (s *HTTPServer) Handle(fn func(ctx Context) error) {
	http.Handle("/", s.uses(http.HandlerFunc(Convert(fn)), s.mdw...))
}

func (s *HTTPServer) ListenAndServe(addr string) error {
	version.INFO(s.Metadata.GetServiceName(), s.Metadata.GetServiceKey())
	zlog.Infof("%s >>> running on %s",
		color.Blue.Add("http"),
		color.Magenta.Add(addr),
	)

	return http.ListenAndServe(addr, nil)
}

func (s *HTTPServer) Shutdown() error {
	return Shutdown()
}
