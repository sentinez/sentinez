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

package network

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// _backendName a name the certificate of httptest covers.
const _backendName = "example.com"

// tlsBackend an https server answering 204, and the pool trusting it.
func tlsBackend(t *testing.T) (*httptest.Server, *x509.CertPool) {
	t.Helper()

	srv := httptest.NewTLSServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
	t.Cleanup(srv.Close)

	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())

	return srv, roots
}

func listenSNI(t *testing.T, fn PassthroughFunc) *StdListener {
	t.Helper()

	lis, err := StdListen("127.0.0.1:0", WithTCP(), WithPassthrough(fn))
	require.NoError(t, err)
	t.Cleanup(func() { _ = lis.Close() })

	return lis
}

// getOver sends a GET over a TLS conn to addr for name, verified by roots.
func getOver(t *testing.T, addr, name string, roots *x509.CertPool,
) (int, error) {
	t.Helper()

	conn, err := tls.Dial("tcp", addr,
		&tls.Config{ServerName: name, RootCAs: roots})
	if err != nil {
		return 0, err
	}
	defer func() { _ = conn.Close() }()

	req, err := http.NewRequest(http.MethodGet, "https://"+name+"/", nil)
	require.NoError(t, err)
	require.NoError(t, req.Write(conn))

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, resp.Body.Close()
}

func TestPassthroughForwardsTLSUntouched(t *testing.T) {
	backend, roots := tlsBackend(t)
	lis := listenSNI(t, func(name string) ([]string, bool) {
		return []string{backend.Listener.Addr().String()},
			name == _backendName
	})

	code, err := getOver(t, lis.Addr().String(), _backendName, roots)

	require.NoError(t, err, "the backend certificate reaches the client")
	assert.Equal(t, http.StatusNoContent, code)
}

func TestPassthroughFailsOverToNextUpstream(t *testing.T) {
	backend, roots := tlsBackend(t)
	down, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	require.NoError(t, down.Close())
	lis := listenSNI(t, func(string) ([]string, bool) {
		return []string{down.Addr().String(),
			backend.Listener.Addr().String()}, true
	})

	code, err := getOver(t, lis.Addr().String(), _backendName, roots)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)
}

func TestPassthroughHandsTerminatedConnToAccept(t *testing.T) {
	backend, roots := tlsBackend(t)
	lis := listenSNI(t, func(string) ([]string, bool) { return nil, false })
	// the edge terminates: it serves the hello it is handed again
	go func() {
		_ = (&http.Server{
			Handler: backend.Config.Handler,
			TLSConfig: &tls.Config{
				Certificates: backend.TLS.Certificates,
			},
		}).ServeTLS(lis, "", "")
	}()

	code, err := getOver(t, lis.Addr().String(), _backendName, roots)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)
}

func TestPassthroughHandsPlainConnToAccept(t *testing.T) {
	lis := listenSNI(t, func(string) ([]string, bool) { return nil, true })
	const sent = "GET / HTTP/1.1\r\nHost: a\r\n\r\n"

	client, err := net.Dial("tcp", lis.Addr().String())
	require.NoError(t, err)
	defer func() { _ = client.Close() }()
	_, err = io.WriteString(client, sent)
	require.NoError(t, err)

	conn, err := lis.Accept()
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()
	got := make([]byte, len(sent))
	_, err = io.ReadFull(conn, got)

	require.NoError(t, err)
	assert.Equal(t, sent, string(got), "the bytes peeked are replayed")
}

func TestPassthroughAcceptFailsOnceClosed(t *testing.T) {
	lis := listenSNI(t, func(string) ([]string, bool) { return nil, false })

	require.NoError(t, lis.Close())
	_, err := lis.Accept()

	assert.ErrorIs(t, err, net.ErrClosed)
}
