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
	"net/http"
	"time"
)

// StandardTransporter the transport of a reverse proxy over TCP; conf,
// when not nil, is the TLS config it dials https with, else the default.
func StandardTransporter(conf *tls.Config) http.RoundTripper {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		base = &http.Transport{}
	}
	transport := base.Clone()

	transport.DialContext = (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext

	if conf != nil {
		// ForceAttemptHTTP2 of the clone keeps h2 with a custom config,
		// which the transport then changes (it adds h2 to NextProtos):
		// a clone keeps conf, shared by several transports, as it is
		transport.TLSClientConfig = conf.Clone()
	}

	return transport
}

func GetInterface(name string) (*net.Interface, error) {
	return net.InterfaceByName(name)
}
