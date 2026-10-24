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
	"fmt"
	"net"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/sentinez/shared/szstore/ja4"
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

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("quic listen: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("quic listen: %w", err)
	}

	ql.udp = udpConn
	ql.tr = &quic.Transport{Conn: udpConn, ConnContext: onQuicConnContext}

	lis, err := ql.tr.Listen(tlsConf, ql.opt.quicConf)
	if err != nil {
		return nil, errors.Join(err, ql.tr.Close(), udpConn.Close())
	}

	ql.QUICListener = lis

	return ql, nil
}

// QuicListener owns its quic.Transport so that every connection gets an id
// (see ConnId) before the TLS handshake starts.
type QuicListener struct {
	http3.QUICListener
	tr  *quic.Transport
	udp *net.UDPConn
	opt Option
}

func (l *QuicListener) Accept(ctx context.Context) (*quic.Conn, error) {
	conn, err := l.QUICListener.Accept(ctx)
	if err != nil {
		return nil, err
	}

	if l.opt.onQuicAccept != nil {
		if err := l.opt.onQuicAccept(conn); err != nil {
			return nil, err
		}
	}

	return conn, nil
}

// Close stops the listener, then the transport and its UDP socket.
func (l *QuicListener) Close() error {
	return errors.Join(l.QUICListener.Close(), l.tr.Close(), l.udp.Close())
}

// onQuicConnContext tags the connection context with a fresh id. The same
// context reaches tls.ClientHelloInfo.Context and quic.Conn.Context, so the
// JA4 fingerprint stored during the handshake can be found again per request.
func onQuicConnContext(
	ctx context.Context, _ *quic.ClientInfo) (context.Context, error) {

	id := newConnId()
	// quic-go cancels ctx when the connection closes or the handshake fails.
	context.AfterFunc(ctx, func() { ja4.Delete(id) })

	return withConnId(ctx, id), nil
}
