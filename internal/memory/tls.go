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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/core/szcrypto/certs"
	"github.com/sentinez/pkg/szlog"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// newSealer the sealer of SZ_DATA_KEY, nil when it is not set: the edge
// then serves no certificate of centraldata.
func newSealer(conf *settingpb.Config) (*crypto.Sealer, error) {
	key := conf.Get(settingpb.Sz_SZ_DATA_KEY)
	if key == "" {
		return nil, nil
	}

	return crypto.NewSealer(key)
}

// LoadCertificate stores the certificate s serves for its domain; a
// setting without one, or whose certificates fail, drops the previous one.
func (m *MemStore) LoadCertificate(s *edgepb.Setting) error {
	ns := s.GetIngress().GetResource().GetDomain()
	sealed := s.GetIngress().GetTls().GetCertificates()
	if len(sealed) == 0 {
		m.certificates.Delete(ns)
		return nil
	}

	cert, err := m.openCertificate(ns, sealed)
	if err != nil {
		m.certificates.Delete(ns)
		return err
	}

	m.certificates.Store(ns, cert)
	szlog.Debugf("edge: certificate of %s loaded", ns)

	return nil
}

func (m *MemStore) openCertificate(ns string, sealed []*edgepb.Certificate,
) (*tls.Certificate, error) {
	if m.sealer == nil {
		return nil, crypto.ErrNoDataKey
	}

	pairs, pool, err := openAll(m.sealer, sealed)
	if err != nil {
		return nil, err
	}

	return certs.Pick(ns, pairs, pool)
}

// openAll opens sealed: the key pairs, and the certificates without a key.
func openAll(sealer *crypto.Sealer, sealed []*edgepb.Certificate,
) ([]tls.Certificate, []*x509.Certificate, error) {
	var (
		pairs []tls.Certificate
		pool  []*x509.Certificate
	)

	for _, c := range sealed {
		certPem, err := sealer.Open(c.GetPem(), c.GetId())
		if err != nil {
			return nil, nil, fmt.Errorf("%s: pem: %w", c.GetId(), err)
		}

		if c.GetPrivateKey() == "" {
			if x := certs.Parse(certPem); x != nil {
				pool = append(pool, x)
			}

			continue
		}

		pair, err := openPair(sealer, c, certPem)
		if err != nil {
			return nil, nil, err
		}

		pairs = append(pairs, pair)
	}

	return pairs, pool, nil
}

func openPair(sealer *crypto.Sealer, c *edgepb.Certificate, certPem string,
) (tls.Certificate, error) {
	keyPem, err := sealer.Open(c.GetPrivateKey(), c.GetId())
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("%s: key: %w", c.GetId(), err)
	}

	pair, err := tls.X509KeyPair([]byte(certPem), []byte(keyPem))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("%s: %w", c.GetId(), err)
	}

	return pair, nil
}

// _upstreamMinVersion the oldest TLS the edge dials an upstream with.
const _upstreamMinVersion = tls.VersionTLS12

// upstreamTLS the TLS config the edge dials the HTTPS upstreams of s with:
// its CAs in place of the system ones, and its client certificate (mTLS).
// nil when s has neither, or dials its upstreams in plain HTTP
// (terminate): the defaults of the transport.
func (m *MemStore) upstreamTLS(s *edgepb.Setting) (*tls.Config, error) {
	t := s.GetIngress().GetTls()
	cas, clients := t.GetUpstreamCas(), t.GetUpstreamCertificates()
	if len(cas) == 0 && len(clients) == 0 || s.GetIngress().GetResource().
		GetTlsMode() == edgepb.TlsMode_TLS_MODE_TERMINATE {
		return nil, nil
	}

	if m.sealer == nil {
		return nil, crypto.ErrNoDataKey
	}

	conf := &tls.Config{MinVersion: _upstreamMinVersion}

	var err error
	if len(cas) != 0 {
		if conf.RootCAs, err = openRoots(m.sealer, cas); err != nil {
			return nil, err
		}
	}

	if len(clients) != 0 {
		cert, err := openClientCertificate(m.sealer, clients)
		if err != nil {
			return nil, err
		}

		conf.Certificates = []tls.Certificate{*cert}
	}

	return conf, nil
}

// openRoots the pool of the CAs cas.
func openRoots(sealer *crypto.Sealer, cas []*edgepb.Certificate,
) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, c := range cas {
		certPem, err := sealer.Open(c.GetPem(), c.GetId())
		if err != nil {
			return nil, fmt.Errorf("%s: pem: %w", c.GetId(), err)
		}

		ca := certs.Parse(certPem)
		if ca == nil {
			return nil, fmt.Errorf("%s: %w", c.GetId(), errNoCertificate)
		}

		pool.AddCert(ca)
	}

	return pool, nil
}

// errNoCertificate a pem without a certificate.
var errNoCertificate = errors.New("no certificate in pem")

// openClientCertificate the client certificate out of sealed, with the
// chain of its issuers.
func openClientCertificate(sealer *crypto.Sealer,
	sealed []*edgepb.Certificate) (*tls.Certificate, error) {

	pairs, pool, err := openAll(sealer, sealed)
	if err != nil {
		return nil, err
	}

	return certs.PickClient(pairs, pool)
}
