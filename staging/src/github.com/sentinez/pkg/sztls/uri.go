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

package sztls

import (
	"crypto/tls"
	"fmt"

	"github.com/sentinez/pkg/szuri"
)

// Endpoint is where a client connects.
type Endpoint struct {
	Hosts []string
	// TLS is nil for plaintext.
	TLS *tls.Config
}

// ParseURI returns the endpoint of uri, of scheme plain, or secure for
// TLS: the mTLS of f when it is set, else TLS trusting the system roots.
// It needs a host, e.g. ParseURI(uri, "kafka", "kafkas", f).
func ParseURI(uri, plain, secure string, f Files) (Endpoint, error) {
	u, err := szuri.Parse(uri)
	if err != nil {
		return Endpoint{}, fmt.Errorf("%s uri: %w", plain, err)
	}

	if err := u.Require(); err != nil {
		return Endpoint{}, fmt.Errorf("%s uri: %w", plain, err)
	}

	isSecure, err := u.Secure(plain, secure)
	if err != nil {
		return Endpoint{}, fmt.Errorf("%s uri: %w", plain, err)
	}

	if !isSecure {
		return Endpoint{Hosts: u.Hosts}, nil
	}

	conf, err := f.ClientConfig()
	if err != nil {
		return Endpoint{}, fmt.Errorf("%s uri: %w", plain, err)
	}

	return Endpoint{Hosts: u.Hosts, TLS: conf}, nil
}
