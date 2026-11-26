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

package sztls_test

import (
	"crypto/tls"
	"testing"

	"github.com/sentinez/pkg/sztls"
	"github.com/sentinez/pkg/sztls/sztlstest"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesOf(t *testing.T) {
	conf := &settingpb.Config{Env: map[string]string{
		"SZ_TLS_CA":   "ca.pem",
		"SZ_TLS_CERT": "cert.pem",
		"SZ_TLS_KEY":  "key.pem",
	}}

	want := sztls.Files{
		CAFile:   "ca.pem",
		CertFile: "cert.pem",
		KeyFile:  "key.pem",
	}
	assert.Equal(t, want, sztls.FilesOf(conf))
}

func TestFilesEnabled(t *testing.T) {
	tests := []struct {
		name string
		give sztls.Files
		want bool
	}{
		{name: "none", give: sztls.Files{}, want: false},
		{name: "ca only", give: sztls.Files{CAFile: "ca.pem"}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.give.Enabled())
		})
	}
}

func TestFilesConfig(t *testing.T) {
	files := sztlstest.NewCA(t).Issue(t, "edge")

	conf, err := files.Config()
	require.NoError(t, err)

	assert.Len(t, conf.Certificates, 1)
	assert.NotNil(t, conf.RootCAs)
	assert.Equal(t, uint16(tls.VersionTLS13), conf.MinVersion)
}

func TestFilesConfigError(t *testing.T) {
	tests := []struct {
		name string
		give sztls.Files
	}{
		{name: "partial", give: sztls.Files{CAFile: "ca.pem"}},
		{
			name: "missing files",
			give: sztls.Files{
				CAFile:   "ca.pem",
				CertFile: "cert.pem",
				KeyFile:  "key.pem",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.give.Config()
			require.Error(t, err)
		})
	}
}

func TestFilesClientConfigSystemRoots(t *testing.T) {
	conf, err := sztls.Files{}.ClientConfig()
	require.NoError(t, err)

	assert.Empty(t, conf.Certificates)
	assert.Nil(t, conf.RootCAs)
	assert.Equal(t, uint16(tls.VersionTLS12), conf.MinVersion)
}

func TestFilesClientConfigMutual(t *testing.T) {
	files := sztlstest.NewCA(t).Issue(t, "centraldata")

	conf, err := files.ClientConfig()
	require.NoError(t, err)

	assert.Len(t, conf.Certificates, 1)
}
