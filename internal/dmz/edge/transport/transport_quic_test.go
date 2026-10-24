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

package transport

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/sentinez/core/szcontext"
	"github.com/sentinez/sentinez/pkg/network"
	"github.com/sentinez/shared/szstore/ja4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuicStoresJA4(t *testing.T) {
	conn, client := dialQuic(t)

	tr := szcontext.GetTransport(OnQuicConnect(context.Background(), conn))
	require.NotNil(t, tr)
	assert.Equal(t, "localhost", tr.GetServerName())

	fingerprint := ja4.Get(tr.GetConnId())
	assert.True(t, strings.HasPrefix(fingerprint, "q13"), fingerprint)

	require.NoError(t, client.CloseWithError(0, ""))
	assert.Eventually(t, func() bool {
		return ja4.Get(tr.GetConnId()) == ""
	}, time.Second, 10*time.Millisecond)
}

// dialQuic returns the server and client sides of one QUIC connection.
func dialQuic(t *testing.T) (*quic.Conn, *quic.Conn) {
	t.Helper()

	lis, err := network.QuicListen("127.0.0.1:0",
		network.WithTLSConfig(newServerTLSConfig(t)))
	require.NoError(t, err)
	t.Cleanup(func() { _ = lis.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	client, err := quic.DialAddr(ctx, lis.Addr().String(), &tls.Config{
		ServerName:         "localhost",
		NextProtos:         []string{http3.NextProtoH3},
		InsecureSkipVerify: true, //nolint:gosec // self-signed test cert
	}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.CloseWithError(0, "") })

	conn, err := lis.Accept(ctx)
	require.NoError(t, err)

	return conn, client
}

func newServerTLSConfig(t *testing.T) *tls.Config {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl,
		&key.PublicKey, key)
	require.NoError(t, err)

	return &tls.Config{
		GetConfigForClient: TLSConfig,
		MinVersion:         tls.VersionTLS13,
		Certificates: []tls.Certificate{
			{Certificate: [][]byte{der}, PrivateKey: key},
		},
	}
}
