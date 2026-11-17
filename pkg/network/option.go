// Copyright 2026 Sentinéz Labs.
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

package network

import (
	"crypto/tls"
	"net"

	"github.com/quic-go/quic-go"
)

type Network string

type Option struct {
	network     string
	onStdAccept func(conn net.Conn) error
	passthrough PassthroughFunc

	tlsConf  *tls.Config
	quicConf *quic.Config
}

type NetworkOption func(*Option)

func WithTCP() NetworkOption {
	return func(o *Option) {
		o.network = "tcp"
	}
}

func WithOnAccept(fn func(conn net.Conn) error) NetworkOption {
	return func(o *Option) {
		o.onStdAccept = fn
	}
}

// WithPassthrough makes the listener read the client hello of each
// connection and forward the ones fn resolves to their upstream untouched
// (TLS passthrough); Accept returns the others only.
func WithPassthrough(fn PassthroughFunc) NetworkOption {
	return func(o *Option) {
		o.passthrough = fn
	}
}

func WithTLSConfig(tls *tls.Config) NetworkOption {
	return func(o *Option) {
		o.tlsConf = tls
	}
}
