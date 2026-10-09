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
	"context"
	"errors"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

func QuicListen(addr string, opts ...NetworkOption) (*QuicListener, error) {
	ql := &QuicListener{}

	for _, opt := range opts {
		opt(&ql.opt)
	}

	if ql.opt.tlsConf == nil {
		return nil, errors.New("quic listen: tls config required")
	}

	// Advertise the "h3" ALPN, otherwise HTTP/3 clients fail the handshake.
	tlsConf := http3.ConfigureTLSConfig(ql.opt.tlsConf)

	lis, err := quic.ListenAddr(addr, tlsConf, ql.opt.quicConf)
	if err != nil {
		return nil, err
	}

	ql.QUICListener = lis

	return ql, nil
}

type QuicListener struct {
	http3.QUICListener
	opt Option
}

func (l *QuicListener) Accept(ctx context.Context) (*quic.Conn, error) {
	conn, err := l.QUICListener.Accept(ctx)
	if err != nil {
		return nil, err
	}

	wconn := newQuicConn(conn)
	go func() { _ = wconn.Close() }()

	if l.opt.onQuicAccept != nil {
		if err := l.opt.onQuicAccept(conn); err != nil {
			return nil, err
		}
	}

	return conn, nil
}
