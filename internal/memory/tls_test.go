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

package memory

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/sentinez/pkg/szseal"
	"github.com/sentinez/pkg/sztable"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// issued a certificate and its key, PEM encoded.
type issued struct {
	cert    *x509.Certificate
	key     *ecdsa.PrivateKey
	certPem string
	keyPem  string
}

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

	keyDer, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	return &issued{
		cert: cert, key: key,
		certPem: string(pem.EncodeToMemory(
			&pem.Block{Type: "CERTIFICATE", Bytes: der})),
		keyPem: string(pem.EncodeToMemory(
			&pem.Block{Type: "PRIVATE KEY", Bytes: keyDer})),
	}
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

func leafTmpl(serial int64, domain string, notAfter time.Time,
) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject:      pkix.Name{CommonName: domain},
		DNSNames:     []string{domain},
		NotBefore:    time.Now().Add(-2 * time.Hour),
		NotAfter:     notAfter,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
}

func newTestSealer(t *testing.T) *szseal.Sealer {
	t.Helper()

	raw := make([]byte, 32)
	_, err := rand.Read(raw)
	require.NoError(t, err)

	sealer, err := szseal.NewSealer(base64.StdEncoding.EncodeToString(raw))
	require.NoError(t, err)

	return sealer
}

func sealed(t *testing.T, sealer *szseal.Sealer, id string,
	c *issued, withKey bool) *edgepb.Certificate {
	t.Helper()

	certPem, err := sealer.Seal(c.certPem, id)
	require.NoError(t, err)

	out := &edgepb.Certificate{Id: id, Pem: certPem}
	if withKey {
		out.PrivateKey, err = sealer.Seal(c.keyPem, id)
		require.NoError(t, err)
	}

	return out
}

func tlsSetting(ns string, certs ...*edgepb.Certificate) *edgepb.Setting {
	st := newSetting("tnt_1", ns)
	if len(certs) > 0 {
		st.Ingress.Tls = &edgepb.Tls{Certificates: certs}
	}

	return st
}

// chain a root, an intermediate it signs and a leaf of domain the
// intermediate signs.
func chain(t *testing.T, domain string, notAfter time.Time,
) (root, inter, leaf *issued) {
	t.Helper()

	root = issue(t, caTmpl(1, "root"), nil)
	inter = issue(t, caTmpl(2, "inter"), root)
	leaf = issue(t, leafTmpl(3, domain, notAfter), inter)

	return root, inter, leaf
}

func TestLoadCertificateServesChain(t *testing.T) {
	sealer := newTestSealer(t)
	root, inter, leaf := chain(t, "tls.example.com", time.Now().Add(time.Hour))
	m := newSyncStore()
	m.sealer = sealer
	t.Cleanup(m.certificates.Clear)

	err := m.LoadCertificate(tlsSetting("tls.example.com",
		sealed(t, sealer, "cacrt_a", root, false),
		sealed(t, sealer, "cacrt_b", inter, false),
		sealed(t, sealer, "cacrt_c", leaf, true),
	))
	require.NoError(t, err)

	got, ok := m.certificates.Load("tls.example.com")
	require.True(t, ok)
	assert.Equal(t, [][]byte{leaf.cert.Raw, inter.cert.Raw},
		got.Certificate, "leaf then intermediate, no root")
}

func TestLoadCertificateDropsInvalid(t *testing.T) {
	sealer := newTestSealer(t)
	_, _, expired := chain(t, "tls.example.com", time.Now().Add(-time.Hour))
	_, _, other := chain(t, "other.example.com", time.Now().Add(time.Hour))
	_, _, valid := chain(t, "tls.example.com", time.Now().Add(time.Hour))
	// sealed for another record: it does not open as cacrt_a
	moved := sealed(t, sealer, "cacrt_b", valid, true)
	moved.Id = "cacrt_a"

	tests := []struct {
		name     string
		giveCert *edgepb.Certificate
	}{
		{"expired", sealed(t, sealer, "cacrt_a", expired, true)},
		{"other domain", sealed(t, sealer, "cacrt_a", other, true)},
		{"no key", sealed(t, sealer, "cacrt_a", valid, false)},
		{"other aad", moved},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newSyncStore()
			m.sealer = sealer
			t.Cleanup(m.certificates.Clear)
			m.certificates.Store("tls.example.com", nil)

			err := m.LoadCertificate(tlsSetting("tls.example.com",
				tt.giveCert))

			require.Error(t, err)
			_, ok := m.certificates.Load("tls.example.com")
			assert.False(t, ok, "the previous certificate is dropped")
		})
	}
}

func TestLoadCertificateNeedsDataKey(t *testing.T) {
	sealer := newTestSealer(t)
	_, _, leaf := chain(t, "tls.example.com", time.Now().Add(time.Hour))
	m := newSyncStore()
	t.Cleanup(m.certificates.Clear)

	err := m.LoadCertificate(tlsSetting("tls.example.com",
		sealed(t, sealer, "cacrt_a", leaf, true)))

	require.ErrorIs(t, err, szseal.ErrNoDataKey)
}

func TestUnloadDropsCertificate(t *testing.T) {
	sealer := newTestSealer(t)
	_, _, leaf := chain(t, "tls.example.com", time.Now().Add(time.Hour))
	m := newSyncStore()
	m.sealer = sealer
	t.Cleanup(m.certificates.Clear)

	m.load("rsrc_1", tlsSetting("tls.example.com",
		sealed(t, sealer, "cacrt_a", leaf, true)))
	_, ok := m.certificates.Load("tls.example.com")
	require.True(t, ok)

	m.unload("rsrc_1", "tls.example.com")

	_, ok = m.certificates.Load("tls.example.com")
	assert.False(t, ok)
}

func TestCentralSettingCarriesCertificatesByUsage(t *testing.T) {
	m := newSyncStore()
	sink := m.CentralSink()
	cert := func(id string, usage tenantpb.CaCertificateUsage,
	) *centraldatapb.Change {
		return caCertRow(t, id, usage)
	}

	const (
		server   = tenantpb.CaCertificateUsage_CA_CERTIFICATE_USAGE_SERVER
		client   = tenantpb.CaCertificateUsage_CA_CERTIFICATE_USAGE_CLIENT
		upstream = tenantpb.CaCertificateUsage_CA_CERTIFICATE_USAGE_UPSTREAM
		mtls     = tenantpb.
				CaCertificateUsage_CA_CERTIFICATE_USAGE_UPSTREAM_CLIENT
	)

	require.NoError(t, sink.Replace([]*centraldatapb.Change{
		resourceRow(t, "rsrc_1", "tnt_1", "tls.example.com"),
		cert("cacrt_b", server),
		cert("cacrt_a", server),
		cert("cacrt_c", client),
		cert("cacrt_d", upstream),
		cert("cacrt_e", mtls),
	}))

	st, ok := m.central.Load("rsrc_1")
	require.True(t, ok)
	tls := st.GetIngress().GetTls()
	assert.Equal(t, []string{"cacrt_a", "cacrt_b"},
		idsOf(tls.GetCertificates()))
	assert.Equal(t, "v1:pem-cacrt_a", tls.GetCertificates()[0].GetPem(),
		"still sealed")
	assert.Equal(t, []string{"cacrt_d"}, idsOf(tls.GetUpstreamCas()))
	assert.Equal(t, []string{"cacrt_e"},
		idsOf(tls.GetUpstreamCertificates()))
}

// caCertRow the change upserting the certificate id of rsrc_1, of usage.
func caCertRow(t *testing.T, id string, usage tenantpb.CaCertificateUsage,
) *centraldatapb.Change {
	t.Helper()

	return upsertOf(t, sztable.TenantCaCertificates, id,
		&tenantpb.CaCertificate{
			Id: id, Owner: &typepb.Owner{ResourceId: "rsrc_1"},
			Usage: usage, Pem: "v1:pem-" + id, PrivateKey: "v1:key",
		})
}

func idsOf(certs []*edgepb.Certificate) []string {
	ids := make([]string, 0, len(certs))
	for _, c := range certs {
		ids = append(ids, c.GetId())
	}

	return ids
}
