// Copyright 2026 Duc-Hung Ho.
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

package edge

import (
	"crypto/tls"
	"sync/atomic"
	"testing"

	"github.com/sentinez/sentinez/api/client/connection/connectiontest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _testDomain = "tls.example.com"

// issued a certificate of domain from a throwaway CA.
func issued(t *testing.T, domain string) *tls.Certificate {
	t.Helper()

	files := connectiontest.NewCA(t).Issue(t, domain)
	cert, err := tls.LoadX509KeyPair(files.CertFile, files.KeyFile)
	require.NoError(t, err)

	return &cert
}

// serveTLS accepts connections with conf until the test ends, writing one
// byte on each so the client reads its session ticket.
func serveTLS(t *testing.T, conf *tls.Config) string {
	t.Helper()

	ln, err := tls.Listen("tcp", "127.0.0.1:0", conf)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			_, _ = conn.Write([]byte{1})
			_ = conn.Close()
		}
	}()

	return ln.Addr().String()
}

// dial connects with the session cache of the client, and reports
// whether it resumed a session.
func dial(addr string, cache tls.ClientSessionCache) (bool, error) {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName:         _testDomain,
		InsecureSkipVerify: true,
		ClientSessionCache: cache,
	})
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()

	if _, err = conn.Read(make([]byte, 1)); err != nil {
		return false, err
	}

	return conn.ConnectionState().DidResume, nil
}

func TestTLSConfigResumesWhileServed(t *testing.T) {
	cert := issued(t, _testDomain)
	conf, err := newTLSConfig("", "",
		func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			return cert, nil
		})
	require.NoError(t, err)
	addr := serveTLS(t, conf)
	cache := tls.NewLRUClientSessionCache(1)

	_, err = dial(addr, cache)
	require.NoError(t, err)
	resumed, err := dial(addr, cache)

	require.NoError(t, err)
	assert.True(t, resumed)
}

func TestTLSConfigRefusesResumingRemovedCertificate(t *testing.T) {
	var served atomic.Pointer[tls.Certificate]
	served.Store(issued(t, _testDomain))
	conf, err := newTLSConfig("", "",
		func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			return served.Load(), nil
		})
	require.NoError(t, err)
	addr := serveTLS(t, conf)
	cache := tls.NewLRUClientSessionCache(1)
	_, err = dial(addr, cache)
	require.NoError(t, err)

	served.Store(nil)
	_, err = dial(addr, cache)

	require.ErrorContains(t, err, "unrecognized name",
		"a full handshake, without certificate")
}
