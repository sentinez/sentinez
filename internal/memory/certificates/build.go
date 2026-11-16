// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package certificates

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"
)

// Pick the first key pair of pairs valid now for ns, with the chain of its
// issuers out of pool.
func Pick(ns string, pairs []tls.Certificate, pool []*x509.Certificate,
) (*tls.Certificate, error) {
	now := time.Now()
	for i := range pairs {
		leaf := pairs[i].Leaf
		if leaf == nil || now.Before(leaf.NotBefore) ||
			now.After(leaf.NotAfter) || leaf.VerifyHostname(ns) != nil {
			continue
		}

		cert := pairs[i]
		cert.Certificate = append(cert.Certificate[:1:1],
			issuers(leaf, pool)...)

		return &cert, nil
	}

	return nil, fmt.Errorf("%s: no valid certificate", ns)
}

// Parse the certificate of the first PEM block of certPem, nil if none.
func Parse(certPem string) *x509.Certificate {
	block, _ := pem.Decode([]byte(certPem))
	if block == nil {
		return nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil
	}

	return cert
}

// issuers the DER chain from the issuer of leaf up, out of pool; the root
// is left out, the client holds it.
func issuers(leaf *x509.Certificate, pool []*x509.Certificate) [][]byte {
	var ders [][]byte
	for cur := leaf; len(ders) < len(pool); {
		next := issuerOf(cur, pool)
		if next == nil || bytes.Equal(next.RawSubject, next.RawIssuer) {
			break
		}

		ders = append(ders, next.Raw)
		cur = next
	}

	return ders
}

func issuerOf(cert *x509.Certificate, pool []*x509.Certificate,
) *x509.Certificate {
	for _, c := range pool {
		if bytes.Equal(c.RawSubject, cert.RawIssuer) &&
			cert.CheckSignatureFrom(c) == nil {
			return c
		}
	}

	return nil
}
