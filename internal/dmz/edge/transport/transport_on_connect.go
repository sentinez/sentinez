// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transport

import (
	"context"
	"crypto/tls"
	"net"
	"reflect"

	netstd "github.com/sentinez/contrib/httphz/net/std"
	"github.com/sentinez/sentinez/pkg/network"
	"github.com/sentinez/shared/zlog"
)

func OnHertzConnect(ctx context.Context, conn net.Conn) context.Context {
	netTlsConn, ok := conn.(*netstd.TLSConn)
	if !ok {
		zlog.Debugf("transport: type: %s", reflect.TypeOf(conn).String())
		return ctx
	}

	tlsConn, ok := netTlsConn.Unwrap().(*tls.Conn)
	if !ok {
		zlog.Debugf("transport: type: %s",
			reflect.TypeOf(netTlsConn.Conn.Unwrap()).String())
		return ctx
	}

	netConn, ok := tlsConn.NetConn().(*network.Conn)
	if !ok {
		zlog.Debugf("transport: type: %s",
			reflect.TypeOf(tlsConn.NetConn()).String())
		return ctx
	}

	zlog.Debugf("transport: connection id: %s", netConn.Id)

	return context.WithValue(ctx, network.ConnectionId, netConn.Id)
}

func OnStandardConnect(ctx context.Context, conn net.Conn) context.Context {
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		zlog.Debugf("transport: type: %s", reflect.TypeOf(conn).String())
		return ctx
	}

	netConn, ok := tlsConn.NetConn().(*network.Conn)
	if !ok {
		zlog.Debugf("transport: type: %s",
			reflect.TypeOf(tlsConn.NetConn()).String())
		return ctx
	}

	zlog.Debugf("transport: connection id: %s", netConn.Id)

	return context.WithValue(ctx, network.ConnectionId, netConn.Id)
}
