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

package corehttp

import (
	"context"
	"fmt"
	"net"

	"github.com/sentinez/core/console"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
)

var _ Server = (*server)(nil)

func DecoreServer(conf *confpb.Config, s Server) Server {
	return &server{
		meta: conf.GetMeta(),
		s:    s,
	}
}

type server struct {
	meta *typepb.XMeta
	s    Server
}

// Handle implements [Server].
func (s *server) Handle(fn RequestHandler) {
	s.s.Handle(fn)
}

// ListenAndServe implements [Server].
func (s *server) ListenAndServe(addr string) error {
	host, port, _ := net.SplitHostPort(addr)
	console.INFO(s.meta.GetServiceName(), s.meta.GetServiceKey(),
		fmt.Sprintf("running on http %s:%s", host, port))

	return s.s.ListenAndServe(addr)
}

// ListenAndServeTLS implements [Server].
func (s *server) ListenAndServeTLS(addr string,
	certFile string, keyFile string) error {

	host, port, _ := net.SplitHostPort(addr)
	console.INFO(s.meta.GetServiceName(), s.meta.GetServiceKey(),
		fmt.Sprintf("running on https %s:%s", host, port))

	return s.s.ListenAndServeTLS(addr, certFile, keyFile)
}

// Shutdown implements [Server].
func (s *server) Shutdown(ctx context.Context) error {
	return s.s.Shutdown(ctx)
}

// Use implements [Server].
func (s *server) Use(mdw ...func(next RequestHandler) RequestHandler) {
	s.s.Use(mdw...)
}
