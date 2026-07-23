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

package corehttp

import (
	"context"
	"crypto/tls"
	"net"
)

type Option struct {
	CertFile    string
	CertKeyFile string
	TLSConfig   *tls.Config
	ServerName  []byte
	OnAccept    func(conn net.Conn) context.Context
}

type ServerOption func(opt *Option)

func WithCertificate(certFile, certKeyFile string) ServerOption {
	return func(opt *Option) {
		if opt == nil {
			return
		}

		opt.CertFile = certFile
		opt.CertKeyFile = certKeyFile
	}
}

func WithTLSConfig(conf *tls.Config) ServerOption {
	return func(opt *Option) {
		if opt == nil {
			return
		}

		opt.TLSConfig = conf
	}
}

func WithServerName(name []byte) ServerOption {
	return func(opt *Option) {
		if opt == nil {
			return
		}

		opt.ServerName = name
	}
}

func WithOnAccept(fn func(conn net.Conn) context.Context) ServerOption {
	return func(opt *Option) {
		if opt == nil {
			return
		}

		opt.OnAccept = fn
	}
}

type Server interface {
	Shutdown(ctx context.Context) error
	ListenAndServe(addr string, opts ...ServerOption) error
	Use(mdw ...func(next RequestHandler) RequestHandler)
	Handle(fn RequestHandler)
}

type ReverseProxy interface {
	Serve(ctx Context)
}
