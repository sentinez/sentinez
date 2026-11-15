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

package szhttp

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionHasCertificate(t *testing.T) {
	getCert := func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		return nil, nil
	}

	tests := []struct {
		name    string
		giveOpt Option
		want    bool
	}{
		{"files", Option{CertFile: "c", CertKeyFile: "k"}, true},
		{"cert file only", Option{CertFile: "c"}, false},
		{
			"GetCertificate",
			Option{TLSConfig: &tls.Config{GetCertificate: getCert}},
			true,
		},
		{
			"certificates",
			Option{TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{{}},
			}},
			true,
		},
		{"empty TLS config", Option{TLSConfig: &tls.Config{}}, false},
		{"nothing", Option{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.giveOpt.HasCertificate())
		})
	}
}
