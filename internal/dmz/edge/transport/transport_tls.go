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

	"github.com/exaring/ja4plus"
	"github.com/sentinez/sentinez/pkg/network"
	"github.com/sentinez/shared/szlog"
	"github.com/sentinez/shared/szstore/ja4"
)

func TLSConfig(chi *tls.ClientHelloInfo) (*tls.Config, error) {
	szlog.Infof("SNI: %s", chi.ServerName)

	connId, isQuic := fingerprintKey(chi)
	if connId == "" {
		return nil, nil
	}

	fingerprint := ja4plus.JA4(chi)
	if isQuic && len(fingerprint) > 0 {
		// ja4plus reads "udp" as DTLS ('d'); JA4 marks QUIC with 'q'.
		fingerprint = "q" + fingerprint[1:]
	}
	szlog.Debugf("transport: generate fingerprint %s", fingerprint)

	ja4.Set(connId, fingerprint)

	return nil, nil
}

// fingerprintKey returns the id the fingerprint is stored under: the
// StdConn id for TCP, the id QuicListener put on the handshake context for
// QUIC.
func fingerprintKey(chi *tls.ClientHelloInfo) (connId string, isQuic bool) {
	if conn, ok := chi.Conn.(*network.StdConn); ok {
		return conn.Id, false
	}

	connId = network.ConnId(chi.Context())

	return connId, connId != ""
}
