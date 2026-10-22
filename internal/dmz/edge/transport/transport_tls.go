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

	conn, ok := chi.Conn.(*network.StdConn)
	if ok {
		fingerprint := ja4plus.JA4(chi)
		szlog.Debugf("transport: generate fingerprint %s", fingerprint)

		ja4.Set(conn.Id, fingerprint)

		szlog.Infof("transport: found fingerprint %s", fingerprint)
	}

	return nil, nil
}
