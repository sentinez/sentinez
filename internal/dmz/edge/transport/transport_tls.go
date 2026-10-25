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

package transport

import (
	"crypto/tls"
	"net"

	"github.com/exaring/ja4plus"
	"github.com/sentinez/sentinez/pkg/network"
	"github.com/sentinez/shared/szlog"
	"github.com/sentinez/shared/szstore/ja4"
)

func TLSConfig(chi *tls.ClientHelloInfo) (*tls.Config, error) {
	szlog.Debugf("SNI: %s", chi.ServerName)

	connId := network.HelloConnId(chi)
	if connId == "" {
		return nil, nil
	}

	fingerprint := ja4plus.JA4(ja4Hello(chi))
	szlog.Debugf("transport: generate fingerprint %s", fingerprint)

	ja4.Set(connId, fingerprint)

	return nil, nil
}

// ja4Hello relabels QUIC handshakes for ja4plus, which reads the protocol
// from LocalAddr().Network(): quic-go reports "udp", read as DTLS ('d'),
// while JA4 marks QUIC with 'q' for "quic". The edge serves no DTLS.
func ja4Hello(chi *tls.ClientHelloInfo) *tls.ClientHelloInfo {
	if chi.Conn == nil || chi.Conn.LocalAddr().Network() != "udp" {
		return chi
	}

	hello := *chi
	hello.Conn = quicConn{chi.Conn}

	return &hello
}

type quicConn struct{ net.Conn }

func (c quicConn) LocalAddr() net.Addr { return quicAddr{c.Conn.LocalAddr()} }

type quicAddr struct{ net.Addr }

func (quicAddr) Network() string { return "quic" }
