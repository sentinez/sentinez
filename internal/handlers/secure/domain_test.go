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

package secure

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/sentinez/internal/memory/settings"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMisdirected(t *testing.T) {
	tests := []struct {
		name    string
		giveSNI string // none: a plain conn, or no SNI
		giveNs  string
		want    bool
	}{
		{"plain", "", "news.example.com", false},
		{"same domain", "news.example.com", "news.example.com", false},
		{"case and dot", "News.Example.com.", "news.example.com", false},
		{"coalesced", "badcheese.example.com", "news.example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "https://x/", nil)
			req.TLS = &tls.ConnectionState{ServerName: tt.giveSNI}

			ctx := httpx.NewContext(req, httptest.NewRecorder())

			assert.Equal(t, tt.want, misdirected(ctx, tt.giveNs))
		})
	}
}

func TestDomainRefusesPassthroughRequest(t *testing.T) {
	const ns = "pass.example.com"
	store := settings.New()
	store.Store(&edgepb.Setting{Ingress: &edgepb.Ingress{
		Resource: &edgepb.Resource{
			Domain:  ns,
			TlsMode: edgepb.TlsMode_TLS_MODE_PASSTHROUGH,
		},
	}})
	t.Cleanup(func() { store.Delete(ns) })

	rec := httptest.NewRecorder()
	ctx := httpx.NewContext(
		httptest.NewRequest(http.MethodGet, "https://"+ns+"/", nil), rec)
	next := chains.HandlerFunc(func(szhttp.Context) error {
		t.Fatal("a passthrough domain reaches no stage")
		return nil
	})

	require.NoError(t, (&DomainBased{settings: store}).Handle(ctx, next))
	assert.Equal(t, http.StatusMisdirectedRequest, rec.Code)
}
