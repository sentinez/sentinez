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

// Package connectiontest issues the certificates of a test of the internal
// mTLS.
package connectiontest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sentinez/sentinez/api/client/connection"
	"github.com/stretchr/testify/require"
)

// CA signs the certificates of a test.
type CA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	dir  string
}

// NewCA returns a CA writing its files under the temporary directory of t.
func NewCA(t *testing.T) *CA {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl,
		&key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	return &CA{cert: cert, key: key, dir: t.TempDir()}
}

// Issue writes a certificate of the common name cn, valid for localhost
// as server and as client, and returns the TLS files of it and of the CA.
func (c *CA) Issue(t *testing.T, cn string) connection.TLS {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn},
		DNSNames:     []string{"localhost", cn},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth,
		},
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, c.cert,
		&key.PublicKey, c.key)
	require.NoError(t, err)

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	return connection.TLS{
		CAFile:   c.write(t, cn+"-ca.pem", "CERTIFICATE", c.cert.Raw),
		CertFile: c.write(t, cn+".pem", "CERTIFICATE", der),
		KeyFile:  c.write(t, cn+"-key.pem", "EC PRIVATE KEY", keyDER),
	}
}

func (c *CA) write(t *testing.T, name, kind string, der []byte) string {
	t.Helper()

	path := filepath.Join(c.dir, name)
	data := pem.EncodeToMemory(&pem.Block{Type: kind, Bytes: der})
	require.NoError(t, os.WriteFile(path, data, 0o600))

	return path
}
