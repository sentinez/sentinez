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

package szkafka

import (
	"testing"

	"github.com/sentinez/pkg/sztls"
	"github.com/sentinez/pkg/sztls/sztlstest"
	"github.com/sentinez/pkg/szuri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseURI(t *testing.T) {
	tests := []struct {
		name        string
		giveURI     string
		wantBrokers []string
		wantTLS     bool
	}{
		{
			name:        "plain",
			giveURI:     "kafka://localhost:9094",
			wantBrokers: []string{"localhost:9094"},
		},
		{
			name:        "several brokers",
			giveURI:     "kafka://a:9092, b:9092",
			wantBrokers: []string{"a:9092", "b:9092"},
		},
		{
			name:        "tls",
			giveURI:     "kafkas://a:9093",
			wantBrokers: []string{"a:9093"},
			wantTLS:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep, err := ParseURI(tt.giveURI, sztls.Files{})
			require.NoError(t, err)

			assert.Equal(t, tt.wantBrokers, ep.Brokers)
			assert.Equal(t, tt.wantTLS, ep.TLS != nil)
		})
	}
}

func TestParseURIMutualTLS(t *testing.T) {
	files := sztlstest.NewCA(t).Issue(t, "edge")

	ep, err := ParseURI("kafkas://a:9093", files)
	require.NoError(t, err)

	require.NotNil(t, ep.TLS)
	assert.Len(t, ep.TLS.Certificates, 1)
}

func TestParseURIError(t *testing.T) {
	tests := []struct {
		name    string
		giveURI string
		wantErr error
	}{
		{
			name:    "bare brokers",
			giveURI: "a:9092,b:9092",
			wantErr: szuri.ErrNoScheme,
		},
		{name: "no broker", giveURI: "kafka://", wantErr: szuri.ErrNoHost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseURI(tt.giveURI, sztls.Files{})
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestParseURIOtherScheme(t *testing.T) {
	_, err := ParseURI("https://a:9093", sztls.Files{})

	var schemeErr *szuri.SchemeError
	require.ErrorAs(t, err, &schemeErr)
}

func TestParseURIPartialFiles(t *testing.T) {
	_, err := ParseURI("kafkas://a:9093", sztls.Files{CAFile: "ca.pem"})

	require.Error(t, err)
}
