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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clientCert a self-signed client certificate, and the pool trusting it.
func clientCert(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "edge"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl,
		&key.PublicKey, key)
	require.NoError(t, err)

	leaf, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	pool := x509.NewCertPool()
	pool.AddCert(leaf)

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key,
		Leaf: leaf}, pool
}

// mtlsUpstream an https server asking for a client certificate of clients,
// and the pool trusting its own certificate.
func mtlsUpstream(t *testing.T, clients *x509.CertPool,
) (*httptest.Server, *x509.CertPool) {
	t.Helper()

	srv := httptest.NewUnstartedServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
	srv.TLS = &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  clients,
	}
	srv.StartTLS()
	t.Cleanup(srv.Close)

	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())

	return srv, roots
}

func TestTransportDialsUpstreamWithMTLS(t *testing.T) {
	cert, clients := clientCert(t)
	srv, roots := mtlsUpstream(t, clients)

	tests := []struct {
		name     string
		giveConf *tls.Config
		wantOk   bool
	}{
		{"system roots", nil, false},
		{"own CA, no client certificate",
			&tls.Config{RootCAs: roots}, false},
		{"own CA and client certificate", &tls.Config{
			RootCAs: roots, Certificates: []tls.Certificate{cert},
		}, true},
	}

	transports := map[string]func(*tls.Config) http.RoundTripper{
		"standard": StandardTransporter,
		"alt-svc": func(c *tls.Config) http.RoundTripper {
			return NewAltSvcTransport(c)
		},
	}

	for kind, newTransport := range transports {
		for _, tt := range tests {
			t.Run(kind+"/"+tt.name, func(t *testing.T) {
				err := get(newTransport(tt.giveConf), srv.URL)

				assert.Equal(t, tt.wantOk, err == nil, "err: %v", err)
			})
		}
	}
}

// get sends a GET to url with tr.
func get(tr http.RoundTripper, url string) error {
	resp, err := (&http.Client{Transport: tr}).Get(url)
	if err != nil {
		return err
	}

	return resp.Body.Close()
}
