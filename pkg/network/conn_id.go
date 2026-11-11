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
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net"

	"github.com/sentinez/pkg/szlog"
)

type connIdKey struct{}

// ConnId returns the connection id that the QUIC listener put on the
// connection context, or "" when ctx has none.
func ConnId(ctx context.Context) string {
	id, _ := ctx.Value(connIdKey{}).(string)

	return id
}

// HelloConnId returns the id of the connection a TLS handshake runs on:
// StdConn.Id (the socket cookie) for TCP, the context id for QUIC, or ""
// for neither.
func HelloConnId(chi *tls.ClientHelloInfo) string {
	if conn, ok := chi.Conn.(*StdConn); ok {
		return conn.Id
	}

	return ConnId(chi.Context())
}

func withConnId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, connIdKey{}, id)
}

// stdConnId returns the kernel socket cookie of a TCP conn as 16 hex
// characters, or a random id when the cookie is unavailable. QUIC keeps
// random ids: all its connections share one UDP socket, hence one cookie.
func stdConnId(conn net.Conn) string {
	cookie, err := sockCookie(conn)
	if err != nil {
		if !errors.Is(err, errors.ErrUnsupported) {
			szlog.Debugf("network: conn id: %v", err)
		}

		return newConnId()
	}

	return hex.EncodeToString(binary.BigEndian.AppendUint64(nil, cookie))
}

func newConnId() string {
	id := make([]byte, 8)
	_, _ = rand.Read(id)

	return hex.EncodeToString(id)
}
