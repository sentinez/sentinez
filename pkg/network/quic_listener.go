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

	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("quic listen: %w", err)
	}

	ql.tr = &quic.Transport{Conn: pc, ConnContext: onQuicConnContext}

	lis, err := ql.tr.Listen(tlsConf, ql.opt.quicConf)
	if err != nil {
		return nil, errors.Join(err, ql.Close())
	}

	ql.QUICListener = lis

	return ql, nil
}

// QuicListener owns its quic.Transport so that every connection gets an id
// (see ConnId) before the TLS handshake starts.
type QuicListener struct {
	http3.QUICListener
	tr  *quic.Transport
	opt Option
}

// Close stops the transport, which also closes the listener, and its UDP
// socket: quic-go only closes sockets it created itself.
func (l *QuicListener) Close() error {
	return errors.Join(l.tr.Close(), l.tr.Conn.Close())
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
