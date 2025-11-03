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

package httpxstd

import (
	"context"
	"net/http"

	"github.com/sentinez/sentinez"
	configspb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/configs/v1"
	"github.com/sentinez/sentinez/pkg/common/color"
	"github.com/sentinez/sentinez/pkg/common/protobuf"
	httpxbase "github.com/sentinez/sentinez/pkg/network/httpx/base"
	"github.com/sentinez/sentinez/pkg/runner"
	"github.com/sentinez/sentinez/pkg/zlog"
)

var _ Server = (*HTTPServer)(nil)

type Server interface {
	httpxbase.Server
	Use(mdw ...func(http.Handler) http.Handler)
	Handle(fn func(ctx Context) error)
}

func NewServer(ctx context.Context) Server {
	runneappConf := runner.GetAppConfig(ctx)
	return &HTTPServer{
		appConf: runneappConf,
	}
}

type HTTPServer struct {
	mdw     []func(http.Handler) http.Handler
	appConf *configspb.AppConfig
}

func (s *HTTPServer) Use(mdw ...func(http.Handler) http.Handler) {
	s.mdw = append(s.mdw, mdw...)
}

func (s *HTTPServer) Handle(fn func(ctx Context) error) {
	http.Handle("/", chain(http.HandlerFunc(Convert(fn)), s.mdw...))
}

func (s *HTTPServer) ListenAndServe(addr string) error {
	if err := protobuf.Validate(s.appConf); err != nil {
		return err
	}

	sentinez.INFO(
		s.appConf.GetMeta().GetServiceName(),
		s.appConf.GetMeta().GetServiceKey(),
	)

	zlog.Infof("%s >>> running on %s",
		color.Blue.Add("http"),
		color.Magenta.Add(addr),
	)

	return http.ListenAndServe(addr, nil)
}

func (s *HTTPServer) Shutdown(_ context.Context) error {
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

		w.Header().Set("Server", sentinez.Name)
	})
}
