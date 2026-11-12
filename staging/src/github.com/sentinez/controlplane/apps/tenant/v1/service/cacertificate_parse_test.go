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

package tenantsvc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _now = time.Date(2026, time.November, 11, 12, 0, 0, 0, time.UTC)

// testCa is a generated certificate, with its key.
type testCa struct {
	cert    *x509.Certificate
	key     *ecdsa.PrivateKey
	certPEM string
	keyPEM  string
}

type caOptions struct {
	subject  pkix.Name
	notAfter time.Time
	notCa    bool
	parent   *testCa
}

// newTestCa generates a certificate; self-signed unless opts has a parent.
func newTestCa(t *testing.T, opts caOptions) *testCa {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := caTemplate(t, opts)
	parent, signer := tmpl, key
	if opts.parent != nil {
		parent, signer = opts.parent.cert, opts.parent.key
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, parent,
		&key.PublicKey, signer)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	return &testCa{
		cert: cert,
		key:  key,
		certPEM: string(pem.EncodeToMemory(
			&pem.Block{Type: "CERTIFICATE", Bytes: der})),
		keyPEM: string(pem.EncodeToMemory(
			&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})),
	}
}

// caTemplate is the certificate to create for opts.
func caTemplate(t *testing.T, opts caOptions) *x509.Certificate {
	t.Helper()

	serial, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	require.NoError(t, err)

	notAfter := opts.notAfter
	if notAfter.IsZero() {
		notAfter = _now.Add(24 * time.Hour)
	}

	return &x509.Certificate{
		SerialNumber:          serial,
		Subject:               opts.subject,
		NotBefore:             _now.Add(-time.Hour),
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		IsCA:                  !opts.notCa,
		KeyUsage: x509.KeyUsageCertSign |
			x509.KeyUsageDigitalSignature,
	}
}

func rootNamed(t *testing.T, org, cn string) *testCa {
	t.Helper()

	name := pkix.Name{CommonName: cn}
	if org != "" {
		name.Organization = []string{org}
	}

	return newTestCa(t, caOptions{subject: name})
}

func TestParseCaBundleRoot(t *testing.T) {
	ca := rootNamed(t, "Acme Trust", "Acme Root CA")

	certs, err := parseCaBundle("\n"+ca.certPEM+"\n", _now, false)

	require.NoError(t, err)
	require.Len(t, certs, 1)

	sum := sha256.Sum256(ca.cert.Raw)
	got := certs[0]
	assert.Equal(t, "Acme Root CA", got.GetName())
	assert.Equal(t, ca.cert.Subject.String(), got.GetSubject())
	assert.Equal(t, ca.cert.Issuer.String(), got.GetIssuer())
	assert.Equal(t, "Acme Trust", got.GetProvider())
	assert.Equal(t, ca.cert.SerialNumber.Text(16), got.GetSerialNumber())
	assert.Equal(t, hex.EncodeToString(sum[:]), got.GetFingerprintSha256())
	assert.True(t, got.GetSelfSigned())
	assert.True(t, ca.cert.NotBefore.Equal(got.GetNotBefore().AsTime()))
	assert.True(t, ca.cert.NotAfter.Equal(got.GetNotAfter().AsTime()))
	assert.Equal(t, ca.certPEM, got.GetPem())
}

func TestParseCaBundleProviderFallsBackToIssuerName(t *testing.T) {
	ca := rootNamed(t, "", "Lone Root")

	certs, err := parseCaBundle(ca.certPEM, _now, false)

	require.NoError(t, err)
	assert.Equal(t, "Lone Root", certs[0].GetProvider())
}

func TestParseCaBundleSplitsTheBundle(t *testing.T) {
	root := rootNamed(t, "Acme Trust", "Acme Root CA")
	inter := newTestCa(t, caOptions{
		subject: pkix.Name{CommonName: "Acme Issuing CA"},
		parent:  root,
	})

	certs, err := parseCaBundle(root.certPEM+inter.certPEM, _now, false)

	require.NoError(t, err)
	require.Len(t, certs, 2)
	assert.Equal(t, "Acme Root CA", certs[0].GetName())
	assert.True(t, certs[0].GetSelfSigned())
	assert.Equal(t, "Acme Issuing CA", certs[1].GetName())
	assert.False(t, certs[1].GetSelfSigned())
	assert.Equal(t, "Acme Trust", certs[1].GetProvider())
	assert.NotEqual(t,
		certs[0].GetFingerprintSha256(), certs[1].GetFingerprintSha256())
}

func leafNamed(t *testing.T, cn string) *testCa {
	t.Helper()

	return newTestCa(t, caOptions{
		subject: pkix.Name{CommonName: cn}, notCa: true,
	})
}

func TestParseCaBundleLeafWhenAllowed(t *testing.T) {
	leaf := leafNamed(t, "www.example.com")

	certs, err := parseCaBundle(leaf.certPEM, _now, true)

	require.NoError(t, err)
	require.Len(t, certs, 1)
	assert.Equal(t, "www.example.com", certs[0].GetName())
	assert.False(t, certs[0].GetSelfSigned())
}

type bundleCase struct {
	name       string
	giveBundle string
}

// invalidBundles are the bundles parseCaBundle refuses.
func invalidBundles(t *testing.T) []bundleCase {
	t.Helper()

	root := rootNamed(t, "Acme Trust", "Acme Root CA")
	leaf := newTestCa(t, caOptions{
		subject: pkix.Name{CommonName: "www.example.com"},
		notCa:   true,
	})
	expired := newTestCa(t, caOptions{
		subject:  pkix.Name{CommonName: "Old CA"},
		notAfter: _now.Add(-time.Minute),
	})
	badDER := string(pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not der")}))

	return []bundleCase{
		{"empty", ""},
		{"blank", " \n\t "},
		{"not a ca", leaf.certPEM},
		{"expired", expired.certPEM},
		{"a private key", root.keyPEM},
		{"a certificate then a key", root.certPEM + root.keyPEM},
		{"text before", "subject=Acme\n" + root.certPEM},
		{"text after", root.certPEM + "trailing"},
		{"not pem", "hello"},
		{"malformed der", badDER},
		{"the same twice", root.certPEM + root.certPEM},
	}
}

func TestParseCaBundleInvalid(t *testing.T) {
	for _, tt := range invalidBundles(t) {
		t.Run(tt.name, func(t *testing.T) {
			certs, err := parseCaBundle(tt.giveBundle, _now, false)

			assert.Empty(t, certs)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestAttachPrivateKey(t *testing.T) {
	ca := rootNamed(t, "Acme Trust", "Acme Root CA")
	certs, err := parseCaBundle(ca.certPEM, _now, false)
	require.NoError(t, err)

	err = attachPrivateKey(certs, keyBlock(t, ca.keyPEM))

	require.NoError(t, err)
	assert.Equal(t, ca.keyPEM, certs[0].GetPrivateKey())
}

func TestAttachPrivateKeyWithoutKey(t *testing.T) {
	ca := rootNamed(t, "Acme Trust", "Acme Root CA")
	certs, err := parseCaBundle(ca.certPEM, _now, false)
	require.NoError(t, err)

	err = attachPrivateKey(certs, nil)

	require.NoError(t, err)
	assert.Empty(t, certs[0].GetPrivateKey())
}

func TestAttachPrivateKeyInvalid(t *testing.T) {
	ca := rootNamed(t, "Acme Trust", "Acme Root CA")
	other := rootNamed(t, "Other", "Other Root CA")
	inter := newTestCa(t, caOptions{
		subject: pkix.Name{CommonName: "Acme Issuing CA"},
		parent:  ca,
	})
	one, err := parseCaBundle(ca.certPEM, _now, false)
	require.NoError(t, err)
	two, err := parseCaBundle(ca.certPEM+inter.certPEM, _now, false)
	require.NoError(t, err)

	tests := []struct {
		name      string
		giveCerts []*tenantpb.CaCertificate
		giveKey   string
	}{
		{"the key of another certificate", one, other.keyPEM},
		{"a bundle without its key", two, other.keyPEM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := attachPrivateKey(tt.giveCerts, keyBlock(t, tt.giveKey))

			require.Error(t, err)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.NotContains(t, err.Error(), "PRIVATE KEY")
			assert.Empty(t, tt.giveCerts[0].GetPrivateKey())
		})
	}
}

func TestAttachPrivateKeyToTheMatchingCertificateOfABundle(t *testing.T) {
	ca := rootNamed(t, "Acme Trust", "Acme Root CA")
	inter := newTestCa(t, caOptions{
		subject: pkix.Name{CommonName: "Acme Issuing CA"},
		parent:  ca,
	})
	certs, err := parseCaBundle(ca.certPEM+inter.certPEM, _now, false)
	require.NoError(t, err)

	require.NoError(t, attachPrivateKey(certs, keyBlock(t, inter.keyPEM)))

	assert.Empty(t, certs[0].GetPrivateKey())
	assert.NotEmpty(t, certs[1].GetPrivateKey())
}

func TestSplitPrivateKeys(t *testing.T) {
	ca := rootNamed(t, "Acme Trust", "Acme Root CA")
	inter := newTestCa(t, caOptions{
		subject: pkix.Name{CommonName: "Acme Issuing CA"},
		parent:  ca,
	})

	rest, keys := splitPrivateKeys(inter.certPEM + inter.keyPEM + ca.certPEM)

	certs, err := parseCaBundle(rest, _now, false)
	require.NoError(t, err)
	assert.Len(t, certs, 2)
	require.Len(t, keys, 1)
	assert.Contains(t, keys[0].Type, "PRIVATE KEY")
}

// keyBlock is the first PEM block of text.
func keyBlock(t *testing.T, text string) *pem.Block {
	t.Helper()

	block, _ := pem.Decode([]byte(text))
	require.NotNil(t, block)

	return block
}
