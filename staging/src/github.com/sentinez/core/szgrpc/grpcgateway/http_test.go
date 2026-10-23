// Copyright 2025 Duc-Hung Ho.
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

package grpcgateway

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/sentinez/core"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/httpconst"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _pong = "pong"

// newTestServer returns a server answering /ping with _pong.
func newTestServer() *XServer {
	srv := newXServer(&settingpb.Config{})
	srv.HTTPMux().HandleFunc("/ping", func(w http.ResponseWriter,
		_ *http.Request) {
		_, _ = io.WriteString(w, _pong)
	})

	return srv
}

// serve runs srv with opts until the test ends.
func serve(t *testing.T, srv *XServer, opts ...szhttp.ServerOption) {
	t.Helper()

	done := make(chan error, 1)
	go func() { done <- srv.ListenAndServe("127.0.0.1:0", opts...) }()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		require.NoError(t, srv.Shutdown(ctx))
		<-done
	})
}

// newTLSConfig returns a TLS config with a self-signed certificate for
// 127.0.0.1.
func newTLSConfig(t *testing.T) *tls.Config {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl,
		&key.PublicKey, key)
	require.NoError(t, err)

	return &tls.Config{
		MinVersion: tls.VersionTLS13,
		Certificates: []tls.Certificate{
			{Certificate: [][]byte{der}, PrivateKey: key},
		},
	}
}

func get(t *testing.T, client *http.Client, url string) *http.Response {
	t.Helper()

	var resp *http.Response
	require.Eventually(t, func() bool {
		r, err := client.Get(url)
		if err != nil {
			return false
		}

		resp = r
		return true
	}, time.Second, 10*time.Millisecond)
	t.Cleanup(func() { _ = resp.Body.Close() })

	return resp
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return string(body)
}

func TestListenAndServeHTTP(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := newTestServer()
	serve(t, srv, szhttp.WithStdListener(ln))

	resp := get(t, http.DefaultClient, "http://"+ln.Addr().String()+"/ping")

	assert.Equal(t, _pong, readBody(t, resp))
	assert.Equal(t, core.Name, resp.Header.Get(httpconst.HeaderServer))
	assert.Empty(t, resp.Header.Get("Alt-Svc"))
}

func TestListenAndServeHTTPS(t *testing.T) {
	tlsConf := newTLSConfig(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	qln, err := quic.ListenAddrEarly("127.0.0.1:0",
		http3.ConfigureTLSConfig(tlsConf), nil)
	require.NoError(t, err)

	srv := newTestServer()
	serve(t, srv,
		szhttp.WithTLSConfig(tlsConf),
		szhttp.WithStdListener(ln),
		szhttp.WithQuicListener(qln),
	)

	clientTLS := &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	tcp := &http.Client{Transport: &http.Transport{TLSClientConfig: clientTLS}}
	h3 := &http.Client{Transport: &http3.Transport{TLSClientConfig: clientTLS}}

	resp := get(t, tcp, "https://"+ln.Addr().String()+"/ping")
	assert.Equal(t, _pong, readBody(t, resp))
	assert.NotEmpty(t, resp.Header.Get("Alt-Svc"))

	resp = get(t, h3, "https://"+qln.Addr().String()+"/ping")
	assert.Equal(t, _pong, readBody(t, resp))
	assert.Equal(t, "HTTP/3.0", resp.Proto)
}

func TestTLSConfig(t *testing.T) {
	_, err := tlsConfig(szhttp.NewOption())
	require.ErrorIs(t, err, errNoCertificate)

	_, err = tlsConfig(szhttp.NewOption(
		szhttp.WithCertificate("missing.pem", "missing-key.pem")))
	require.Error(t, err)
	assert.NotErrorIs(t, err, errNoCertificate)
}
