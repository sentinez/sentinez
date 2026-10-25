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
	"encoding/hex"
)

type connIdKey struct{}

// ConnId returns the connection id that the QUIC listener put on the
// connection context, or "" when ctx has none.
func ConnId(ctx context.Context) string {
	id, _ := ctx.Value(connIdKey{}).(string)

	return id
}

// HelloConnId returns the id of the connection a TLS handshake runs on:
// StdConn.Id for TCP, the context id for QUIC, or "" for neither.
func HelloConnId(chi *tls.ClientHelloInfo) string {
	if conn, ok := chi.Conn.(*StdConn); ok {
		return conn.Id
	}

	return ConnId(chi.Context())
}

func withConnId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, connIdKey{}, id)
}

func newConnId() string {
	id := make([]byte, 8)
	_, _ = rand.Read(id)

	return hex.EncodeToString(id)
}
