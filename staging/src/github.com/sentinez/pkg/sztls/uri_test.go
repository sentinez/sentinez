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
	"testing"

	"github.com/sentinez/pkg/sztls"
	"github.com/sentinez/pkg/sztls/sztlstest"
	"github.com/sentinez/pkg/szuri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseURI(t *testing.T) {
	tests := []struct {
		name      string
		giveURI   string
		wantHosts []string
		wantTLS   bool
	}{
		{
			name:      "plain",
			giveURI:   "olric://e1:3320",
			wantHosts: []string{"e1:3320"},
		},
		{
			name:      "several hosts",
			giveURI:   "olric://e1:3320, e2:3320",
			wantHosts: []string{"e1:3320", "e2:3320"},
		},
		{
			name:      "tls",
			giveURI:   "olrics://e1:3321",
			wantHosts: []string{"e1:3321"},
			wantTLS:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep, err := sztls.ParseURI(tt.giveURI, "olric", "olrics",
				sztls.Files{})
			require.NoError(t, err)

			assert.Equal(t, tt.wantHosts, ep.Hosts)
			assert.Equal(t, tt.wantTLS, ep.TLS != nil)
		})
	}
}

func TestParseURIMutualTLS(t *testing.T) {
	files := sztlstest.NewCA(t).Issue(t, "centraldata")

	ep, err := sztls.ParseURI("olrics://e1:3321", "olric", "olrics", files)
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
		{name: "bare", giveURI: "e1:3320", wantErr: szuri.ErrNoScheme},
		{name: "no host", giveURI: "olric://", wantErr: szuri.ErrNoHost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sztls.ParseURI(tt.giveURI, "olric", "olrics",
				sztls.Files{})
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestParseURIOtherScheme(t *testing.T) {
	_, err := sztls.ParseURI("redis://e1:3320", "olric", "olrics",
		sztls.Files{})

	var schemeErr *szuri.SchemeError
	require.ErrorAs(t, err, &schemeErr)
}
