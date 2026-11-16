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
	"fmt"

	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/pkg/szlog"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/memory/certificates"
)

// newSealer the sealer of SENZ_DATA_KEY, nil when it is not set: the edge
// then serves no certificate of centraldata.
func newSealer(conf *settingpb.Config) (*crypto.Sealer, error) {
	key := conf.Get(settingpb.Senz_SENZ_DATA_KEY)
	if key == "" {
		return nil, nil
	}

	return crypto.NewSealer(key)
}

// LoadCertificate stores the certificate s serves for its domain; a
// setting without one, or whose certificates fail, drops the previous one.
func (m *MemStore) LoadCertificate(s *edgepb.Setting) error {
	ns := s.GetIngress().GetResource().GetDomain()
	certs := s.GetIngress().GetTls().GetCertificates()
	if len(certs) == 0 {
		m.certificates.Delete(ns)
		return nil
	}

	cert, err := m.openCertificate(ns, certs)
	if err != nil {
		m.certificates.Delete(ns)
		return err
	}

	m.certificates.Store(ns, cert)
	szlog.Debugf("edge: certificate of %s loaded", ns)

	return nil
}

func (m *MemStore) openCertificate(ns string, certs []*edgepb.Certificate,
) (*tls.Certificate, error) {
	if m.sealer == nil {
		return nil, crypto.ErrNoDataKey
	}

	pairs, pool, err := openAll(m.sealer, certs)
	if err != nil {
		return nil, err
	}

	return certificates.Pick(ns, pairs, pool)
}

// openAll opens certs: the key pairs, and the certificates without a key.
func openAll(sealer *crypto.Sealer, certs []*edgepb.Certificate,
) ([]tls.Certificate, []*x509.Certificate, error) {
	var (
		pairs []tls.Certificate
		pool  []*x509.Certificate
	)

	for _, c := range certs {
		certPem, err := sealer.Open(c.GetPem(), c.GetId())
		if err != nil {
			return nil, nil, fmt.Errorf("%s: pem: %w", c.GetId(), err)
		}

		if c.GetPrivateKey() == "" {
			if x := certificates.Parse(certPem); x != nil {
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
