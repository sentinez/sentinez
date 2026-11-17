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

package szhttp

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type Option struct {
	CertFile    string
	CertKeyFile string
	TLSConfig   *tls.Config
	ServerName  []byte

	OnStdConnect func(context.Context, net.Conn) context.Context
	StdListener  net.Listener

	OnQuicConnect func(context.Context, *quic.Conn) context.Context
	QuicListener  http3.QUICListener
}

func NewOption(opts ...ServerOption) *Option {
	var option Option
	for _, opt := range opts {
		opt(&option)
	}

	return &option
}

// HasCertificate reports whether there is a certificate to serve over TLS:
// the files, or those the TLS config gives by itself (GetCertificate).
func (o *Option) HasCertificate() bool {
	if o.CertFile != "" && o.CertKeyFile != "" {
		return true
	}

	return o.TLSConfig != nil && (len(o.TLSConfig.Certificates) != 0 ||
		o.TLSConfig.GetCertificate != nil)
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

func WithOnStdConnect(
	fn func(context.Context, net.Conn) context.Context) ServerOption {
	return func(opt *Option) {
		if opt == nil {
			return
		}

		opt.OnStdConnect = fn
	}
}

func WithStdListener(ln net.Listener) ServerOption {
	return func(opt *Option) {
		if opt == nil {
			return
		}

		opt.StdListener = ln
	}
}

func WithOnQuicConnect(
	fn func(context.Context, *quic.Conn) context.Context) ServerOption {
	return func(opt *Option) {
		if opt == nil {
			return
		}

		opt.OnQuicConnect = fn
	}
}

func WithQuicListener(ln http3.QUICListener) ServerOption {
	return func(opt *Option) {
		if opt == nil {
			return
		}

		opt.QuicListener = ln
	}
}

type Server interface {
	Shutdown(ctx context.Context) error
	ListenAndServe(addr string, opts ...ServerOption) error
	Use(mdw ...func(next RequestHandler) RequestHandler)
	Handle(fn RequestHandler)
	// AcceptReverse a proxy to the upstream target ("http://host" or
	// "https://host"); conf, when not nil, is the TLS config it dials an
	// https target with (CAs, client certificate), else the defaults.
	AcceptReverse(target string, conf *tls.Config) (ReverseProxy, error)
}

type ReverseProxy interface {
	Serve(ctx Context)
}
