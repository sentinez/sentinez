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

// Package certs picks the certificate a TLS endpoint presents out of key
// pairs, and builds the chain of its issuers.
package certs

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"
)

// ErrNoValidCertificate no key pair is valid now (for the host).
var ErrNoValidCertificate = errors.New("no valid certificate")

// Pick the first key pair of pairs valid now for host, with the
// chain of its issuers out of pool.
func Pick(host string, pairs []tls.Certificate,
	pool []*x509.Certificate) (*tls.Certificate, error) {

	cert := pick(pairs, pool, func(leaf *x509.Certificate) bool {
		return leaf.VerifyHostname(host) == nil
	})
	if cert == nil {
		return nil, fmt.Errorf("%s: %w", host, ErrNoValidCertificate)
	}

	return cert, nil
}

// PickClient the first key pair of pairs valid now, with the
// chain of its issuers out of pool: the client certificate of a TLS
// connection.
func PickClient(pairs []tls.Certificate,
	pool []*x509.Certificate) (*tls.Certificate, error) {

	cert := pick(pairs, pool, func(*x509.Certificate) bool { return true })
	if cert == nil {
		return nil, fmt.Errorf("client: %w", ErrNoValidCertificate)
	}

	return cert, nil
}

// Parse the certificate of the first PEM block of certPem, nil
// if none.
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

// ChainOf the DER chain from the issuer of leaf up, out of pool. The root
// is left out: the peer holds it.
func ChainOf(leaf *x509.Certificate, pool []*x509.Certificate) [][]byte {
	var ders [][]byte
	for cur := leaf; len(ders) < len(pool); {
		next := issuerOf(cur, pool)
		if next == nil || IsSelfSigned(next) {
			break
		}

		ders = append(ders, next.Raw)
		cur = next
	}

	return ders
}

// IsSelfSigned reports whether cert is signed by its own key, as a root CA.
func IsSelfSigned(cert *x509.Certificate) bool {
	return bytes.Equal(cert.RawSubject, cert.RawIssuer) &&
		cert.CheckSignatureFrom(cert) == nil
}

// pick the first key pair of pairs valid now whose leaf is ok, with the
// chain of its issuers out of pool; nil if none.
func pick(pairs []tls.Certificate, pool []*x509.Certificate,
	ok func(leaf *x509.Certificate) bool) *tls.Certificate {

	now := time.Now()
	for i := range pairs {
		leaf := pairs[i].Leaf
		if leaf == nil || now.Before(leaf.NotBefore) ||
			now.After(leaf.NotAfter) || !ok(leaf) {
			continue
		}

		cert := pairs[i]
		cert.Certificate = append(cert.Certificate[:1:1],
			ChainOf(leaf, pool)...)

		return &cert
	}

	return nil
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
