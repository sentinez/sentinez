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

// Package certificates keeps the TLS certificates the edge serves, keyed by
// namespace (the resource domain, as sznormalize.ResourceDomain gives it).
package certificates

import (
	"crypto/tls"
	"sync"

	"github.com/sentinez/pkg/sznormalize"
	"github.com/sentinez/pkg/szsync"
)

var (
	_inst *Certificates
	_once sync.Once
)

// New returns the process-wide certificate store.
func New() *Certificates {
	_once.Do(func() {
		_inst = &Certificates{
			space: szsync.NewMap[string, *tls.Certificate](),
		}
	})

	return _inst
}

// Certificates is a concurrent map of namespace to certificate.
type Certificates struct {
	space *szsync.Map[string, *tls.Certificate]
}

// Store sets the certificate of namespace ns.
func (c *Certificates) Store(ns string, cert *tls.Certificate) {
	c.space.Store(ns, cert)
}

// Load returns the certificate of namespace ns.
func (c *Certificates) Load(ns string) (*tls.Certificate, bool) {
	return c.space.Load(ns)
}

// Delete drops the certificate of namespace ns.
func (c *Certificates) Delete(ns string) {
	c.space.Delete(ns)
}

// Clear drops every certificate.
func (c *Certificates) Clear() {
	c.space.Clear()
}

// GetCertificate implements tls.Config.GetCertificate with SNI lookup. No
// match returns nil: the handshake falls back to tls.Config.Certificates.
func (c *Certificates) GetCertificate(
	hello *tls.ClientHelloInfo,
) (*tls.Certificate, error) {
	cert, _ := c.space.Load(sznormalize.ResourceDomain(hello.ServerName))

	return cert, nil
}
