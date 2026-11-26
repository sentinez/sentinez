// Copyright 2025 Duc-Hung Ho.
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

// Package connectiontest issues the certificates of a test of the internal
// mTLS.
package connectiontest

import (
	"testing"

	"github.com/sentinez/core/szgrpc/client/connection"
	"github.com/sentinez/pkg/sztls/sztlstest"
)

// CA signs the certificates of a test.
type CA struct {
	ca *sztlstest.CA
}

// NewCA returns a CA writing its files under the temporary directory of t.
func NewCA(t *testing.T) *CA {
	t.Helper()

	return &CA{ca: sztlstest.NewCA(t)}
}

// Issue writes a certificate of the common name cn, valid for localhost
// as server and as client, and returns the TLS files of it and of the CA.
func (c *CA) Issue(t *testing.T, cn string) connection.TLS {
	t.Helper()

	return connection.TLS(c.ca.Issue(t, cn))
}
