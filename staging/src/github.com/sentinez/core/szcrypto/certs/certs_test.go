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

package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _host = "shop.example.com"

// issued a certificate with its key.
type issued struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
}

// issue tmpl, signed by parent, or by itself without one.
func issue(t *testing.T, tmpl *x509.Certificate, parent *issued) *issued {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	signer, signerCert := key, tmpl
	if parent != nil {
		signer, signerCert = parent.key, parent.cert
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, signerCert,
		&key.PublicKey, signer)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	return &issued{cert: cert, key: key}
}

func caTmpl(serial int64, name string) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber:          big.NewInt(serial),
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
}

func leafTmpl(serial int64, notAfter time.Time) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject:      pkix.Name{CommonName: _host},
		DNSNames:     []string{_host},
		NotBefore:    time.Now().Add(-2 * time.Hour),
		NotAfter:     notAfter,
	}
}

// pki a root, an intermediate it signs, and a leaf the intermediate signs.
type pki struct {
	root, inter, leaf *issued
}

func newPki(t *testing.T, notAfter time.Time) pki {
	t.Helper()

	root := issue(t, caTmpl(1, "root"), nil)
	inter := issue(t, caTmpl(2, "inter"), root)

	return pki{root: root, inter: inter,
		leaf: issue(t, leafTmpl(3, notAfter), inter)}
}

func (p pki) pool() []*x509.Certificate {
	return []*x509.Certificate{p.root.cert, p.inter.cert}
}

func pairOf(i *issued) tls.Certificate {
	return tls.Certificate{
		Certificate: [][]byte{i.cert.Raw},
		PrivateKey:  i.key,
		Leaf:        i.cert,
	}
}

func TestPickServesTheChainWithoutTheRoot(t *testing.T) {
	p := newPki(t, time.Now().Add(time.Hour))

	got, err := Pick(_host, []tls.Certificate{pairOf(p.leaf)}, p.pool())

	require.NoError(t, err)
	assert.Equal(t, [][]byte{p.leaf.cert.Raw, p.inter.cert.Raw},
		got.Certificate)
}

func TestPickSkipsUnfitPairs(t *testing.T) {
	expired := newPki(t, time.Now().Add(-time.Minute))
	valid := newPki(t, time.Now().Add(time.Hour))
	pairs := []tls.Certificate{pairOf(expired.leaf), pairOf(valid.leaf)}

	got, err := Pick(_host, pairs, valid.pool())

	require.NoError(t, err)
	assert.Equal(t, valid.leaf.cert, got.Leaf)
}

func TestPickFailsWithoutValidPair(t *testing.T) {
	p := newPki(t, time.Now().Add(time.Hour))
	tests := []struct {
		name      string
		giveHost  string
		givePairs []tls.Certificate
	}{
		{name: "other host", giveHost: "other.example.com",
			givePairs: []tls.Certificate{pairOf(p.leaf)}},
		{name: "no leaf", giveHost: _host,
			givePairs: []tls.Certificate{{}}},
		{name: "no pair", giveHost: _host},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Pick(tt.giveHost, tt.givePairs, p.pool())

			assert.ErrorIs(t, err, ErrNoValidCertificate)
		})
	}
}

func TestPickClientIgnoresTheHost(t *testing.T) {
	p := newPki(t, time.Now().Add(time.Hour))

	got, err := PickClient([]tls.Certificate{pairOf(p.leaf)}, p.pool())

	require.NoError(t, err)
	assert.Equal(t, [][]byte{p.leaf.cert.Raw, p.inter.cert.Raw},
		got.Certificate)
}

func TestPickClientFailsWithoutValidPair(t *testing.T) {
	p := newPki(t, time.Now().Add(-time.Minute))

	_, err := PickClient([]tls.Certificate{pairOf(p.leaf)}, p.pool())

	assert.ErrorIs(t, err, ErrNoValidCertificate)
}

func TestPickKeepsThePairs(t *testing.T) {
	p := newPki(t, time.Now().Add(time.Hour))
	pairs := []tls.Certificate{pairOf(p.leaf)}

	_, err := Pick(_host, pairs, p.pool())

	require.NoError(t, err)
	assert.Len(t, pairs[0].Certificate, 1)
}

func TestParse(t *testing.T) {
	p := newPki(t, time.Now().Add(time.Hour))
	certPem := string(pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: p.leaf.cert.Raw}))
	tests := []struct {
		name    string
		givePem string
		want    *x509.Certificate
	}{
		{name: "certificate", givePem: certPem, want: p.leaf.cert},
		{name: "not pem", givePem: "garbage"},
		{name: "malformed", givePem: string(pem.EncodeToMemory(
			&pem.Block{Type: "CERTIFICATE", Bytes: []byte("x")}))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Parse(tt.givePem))
		})
	}
}

func TestChainOf(t *testing.T) {
	p := newPki(t, time.Now().Add(time.Hour))
	tests := []struct {
		name     string
		givePool []*x509.Certificate
		want     [][]byte
	}{
		{name: "up to the root", givePool: p.pool(),
			want: [][]byte{p.inter.cert.Raw}},
		{name: "issuer missing", givePool: []*x509.Certificate{p.root.cert}},
		{name: "empty pool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ChainOf(p.leaf.cert, tt.givePool))
		})
	}
}

func TestIsSelfSigned(t *testing.T) {
	p := newPki(t, time.Now().Add(time.Hour))
	tests := []struct {
		name     string
		giveCert *x509.Certificate
		want     bool
	}{
		{name: "root", giveCert: p.root.cert, want: true},
		{name: "intermediate", giveCert: p.inter.cert},
		{name: "leaf", giveCert: p.leaf.cert},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsSelfSigned(tt.giveCert))
		})
	}
}
