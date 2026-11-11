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
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"strings"
	"time"

	"github.com/sentinez/pkg/szerror"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	_pemCertificate = "CERTIFICATE"
	_pemBegin       = "-----BEGIN "
	_keyBlockSuffix = "PRIVATE KEY"
)

// errNoCertificate a pem without any certificate.
var errNoCertificate = szerror.StatusInvalidArgumentF(
	"pem holds no certificate")

// parseCaBundle returns one CaCertificate per certificate of bundle, with
// the fields parsed from it (the pem normalized to that one certificate).
// Every certificate must be a CA that has not expired, and the bundle holds
// nothing else, nor the same certificate twice.
func parseCaBundle(
	bundle string, now time.Time) ([]*tenantpb.CaCertificate, error) {

	var certs []*tenantpb.CaCertificate

	seen := make(map[string]struct{})
	rest := bytes.TrimSpace([]byte(bundle))

	for len(rest) != 0 {
		// pem.Decode skips what comes before a block: refuse it
		block, next := pem.Decode(rest)
		if block == nil || !bytes.HasPrefix(rest, []byte(_pemBegin)) {
			return nil, szerror.StatusInvalidArgumentF(
				"pem holds data that is not a certificate")
		}

		cert, err := parseCaBlock(block, now)
		if err != nil {
			return nil, err
		}

		fingerprint := cert.GetFingerprintSha256()
		if _, ok := seen[fingerprint]; ok {
			return nil, szerror.StatusInvalidArgumentF(
				"certificate %q is repeated", cert.GetName())
		}

		seen[fingerprint] = struct{}{}
		certs = append(certs, cert)
		rest = bytes.TrimSpace(next)
	}

	if len(certs) == 0 {
		return nil, errNoCertificate
	}

	return certs, nil
}

func parseCaBlock(
	block *pem.Block, now time.Time) (*tenantpb.CaCertificate, error) {

	if block.Type != _pemCertificate {
		return nil, szerror.StatusInvalidArgumentF(
			"pem block %q is not a certificate", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, szerror.StatusInvalidArgumentF(
			"certificate is malformed: %v", err)
	}

	if !cert.BasicConstraintsValid || !cert.IsCA {
		return nil, szerror.StatusInvalidArgumentF(
			"%q is not a CA certificate", cert.Subject.String())
	}

	if now.After(cert.NotAfter) {
		return nil, szerror.StatusInvalidArgumentF(
			"%q expired at %s", cert.Subject.String(),
			cert.NotAfter.UTC().Format(time.RFC3339))
	}

	return newCaCertificate(cert), nil
}

// newCaCertificate describes cert. Its name is the common name, or else the
// subject.
func newCaCertificate(cert *x509.Certificate) *tenantpb.CaCertificate {
	sum := sha256.Sum256(cert.Raw)

	name := cert.Subject.CommonName
	if name == "" {
		name = cert.Subject.String()
	}

	return &tenantpb.CaCertificate{
		Name: name,
		Pem: string(pem.EncodeToMemory(&pem.Block{
			Type: _pemCertificate, Bytes: cert.Raw,
		})),
		Subject:           cert.Subject.String(),
		Issuer:            cert.Issuer.String(),
		Provider:          providerOf(cert),
		SerialNumber:      cert.SerialNumber.Text(16),
		FingerprintSha256: hex.EncodeToString(sum[:]),
		NotBefore:         timestamppb.New(cert.NotBefore),
		NotAfter:          timestamppb.New(cert.NotAfter),
		SelfSigned:        isSelfSigned(cert),
	}
}

// providerOf is the vendor of the CA: the first organization of the issuer,
// else its common name.
func providerOf(cert *x509.Certificate) string {
	if orgs := cert.Issuer.Organization; len(orgs) != 0 && orgs[0] != "" {
		return orgs[0]
	}

	return cert.Issuer.CommonName
}

// isSelfSigned reports whether cert is signed by its own key, as a root CA.
func isSelfSigned(cert *x509.Certificate) bool {
	return bytes.Equal(cert.RawSubject, cert.RawIssuer) &&
		cert.CheckSignatureFrom(cert) == nil
}

// attachPrivateKey sets the private key of certs, which must be the one of
// the only certificate. Without key it does nothing. The key is stored as
// one normalized PEM block, and no error carries its content.
func attachPrivateKey(
	certs []*tenantpb.CaCertificate, keyPEM string) error {

	if strings.TrimSpace(keyPEM) == "" {
		return nil
	}

	if len(certs) != 1 {
		return szerror.StatusInvalidArgumentF(
			"a private key needs exactly one certificate, got %d",
			len(certs))
	}

	block := privateKeyBlock(keyPEM)
	if block == nil {
		return szerror.StatusInvalidArgumentF(
			"private_key holds no PEM private key")
	}

	normalized := pem.EncodeToMemory(&pem.Block{
		Type: block.Type, Bytes: block.Bytes,
	})

	// it also refuses a key that is encrypted with a passphrase
	_, err := tls.X509KeyPair([]byte(certs[0].GetPem()), normalized)
	if err != nil {
		return szerror.StatusInvalidArgumentF(
			"private_key is not usable or does not match the certificate")
	}

	certs[0].PrivateKey = string(normalized)

	return nil
}

// privateKeyBlock is the first PEM private key block of keyPEM, or nil.
func privateKeyBlock(keyPEM string) *pem.Block {
	// pem.Decode finds a block only at the start of a line
	rest := []byte(strings.TrimSpace(keyPEM))

	for {
		var block *pem.Block

		block, rest = pem.Decode(rest)
		if block == nil {
			return nil
		}

		if strings.HasSuffix(block.Type, _keyBlockSuffix) {
			return block
		}
	}
}
