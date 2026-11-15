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
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertificatesGetCertificate(t *testing.T) {
	exact := &tls.Certificate{}

	c := New()
	t.Cleanup(c.Clear)
	c.Store("api.example.com", exact)

	tests := []struct {
		name     string
		giveName string
		want     *tls.Certificate
	}{
		{"exact", "api.example.com", exact},
		{"case and dot", "Api.Example.Com.", exact},
		{"other domain", "www.example.com", nil},
		{"empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.GetCertificate(
				&tls.ClientHelloInfo{ServerName: tt.giveName},
			)
			require.NoError(t, err)
			assert.Same(t, tt.want, got)
		})
	}
}

func TestCertificatesDelete(t *testing.T) {
	c := New()
	t.Cleanup(c.Clear)
	c.Store("example.com", &tls.Certificate{})

	c.Delete("example.com")

	_, ok := c.Load("example.com")
	assert.False(t, ok)
}
