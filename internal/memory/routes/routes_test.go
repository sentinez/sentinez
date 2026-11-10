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

package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchLocation(t *testing.T) {
	tests := []struct {
		name         string
		givePath     string
		giveLocation string
		want         bool
	}{
		{"root", "/", "/", true},
		{"root prefix", "/foo", "/", true},
		{"exact", "/api", "/api", true},
		{"sub path", "/api/x", "/api", true},
		{"no segment boundary", "/apix", "/api", false},
		{"trailing slash location", "/api/x", "/api/", true},
		{"other", "/web", "/api", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchLocation([]byte(tt.givePath), []byte(tt.giveLocation))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestJoinRewrite(t *testing.T) {
	tests := []struct {
		name        string
		giveRewrite string
		giveRest    string
		want        string
	}{
		{"root to root", "/", "", "/"},
		{"root keeps rest", "/", "foo", "/foo"},
		{"no double slash", "/", "/x", "/x"},
		{"prefix rewrite", "/v1", "/users", "/v1/users"},
		{"prefix rewrite exact", "/v1", "", "/v1"},
		{"trailing slash rewrite", "/v1/", "/users", "/v1/users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			joinRewrite(&buf, []byte(tt.giveRewrite), []byte(tt.giveRest))
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestPick(t *testing.T) {
	// Sorted longest first, as Store keeps them.
	routes := newRoutes("/local", "/api", "/")

	tests := []struct {
		name         string
		giveLocation string
		givePath     string
		want         string
	}{
		{"longest prefix", "", "/api/users", "/api"},
		{"root fallback", "", "/login", "/"},
		{"route to", "/local", "/login", "/local"},
		{"route to overrides prefix", "/", "/api/users", "/"},
		{"unknown route to falls back", "/nope", "/api/x", "/api"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pick(routes, tt.giveLocation, []byte(tt.givePath))

			assert.Equal(t, tt.want, got.loc.GetLocation())
		})
	}
}

func TestPickNone(t *testing.T) {
	routes := newRoutes("/api")

	assert.Nil(t, pick(routes, "", []byte("/web")))
}

func TestRouteUpstreamRoundRobin(t *testing.T) {
	r := newRoute(&edgepb.Location{
		Location: "/",
		ProxyPass: []*edgepb.Upstream{
			{Server: "http://a"}, {Server: "http://b"}, {Server: "http://c"},
		},
	})

	want := []string{
		"http://a", "http://b", "http://c", "http://a", "http://b",
	}

	got := make([]string, 0, len(want))
	for range want {
		server, err := r.upstream()
		require.NoError(t, err)
		got = append(got, server)
	}

	assert.Equal(t, want, got)
}

func TestRouteUpstreamNone(t *testing.T) {
	r := newRoute(&edgepb.Location{Location: "/api"})

	_, err := r.upstream()

	assert.Error(t, err)
}

func TestStoreSkipsLocationWithoutUpstream(t *testing.T) {
	router := New()
	router.Store(&edgepb.Server{
		ResourceDomain: "test-no-upstream",
		Locations: []*edgepb.Location{
			{Location: "/api"},
			{
				Location:  "/",
				ProxyPass: []*edgepb.Upstream{{Server: "http://a"}},
			},
		},
	})
	t.Cleanup(func() { router.Delete("test-no-upstream") })

	routes, ok := router.routes.Load("test-no-upstream")
	require.True(t, ok)
	require.Len(t, routes, 1)
	assert.Equal(t, "/", routes[0].loc.GetLocation())
}

func newRoutes(locations ...string) []*route {
	routes := make([]*route, 0, len(locations))
	for _, loc := range locations {
		routes = append(routes, newRoute(&edgepb.Location{Location: loc}))
	}

	return routes
}

func TestForwardRecordsHeaders(t *testing.T) {
	r := newRoute(&edgepb.Location{
		Location:     "/api",
		ProxyRewrite: "/",
		ProxyPass:    []*edgepb.Upstream{{Server: "origin.example.com"}},
		ProxySetHeaders: map[string]string{
			"X-Env":  "prod",
			"X-Host": "$host",
		},
	})
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/a",
		nil)
	ctx := httpx.NewContext(req, httptest.NewRecorder())

	upstream, err := forward(ctx, r)

	require.NoError(t, err)
	assert.NotEmpty(t, upstream)
	want := map[string]string{
		"X-Forwarded-Prefix": "/api",
		"X-Env":              "prod",
		"X-Host":             "example.com",
	}
	assert.Equal(t, want, ctx.X().GetSetHeaders())
	assert.Equal(t, "prod", req.Header.Get("X-Env"))
	assert.Equal(t, "/api", req.Header.Get("X-Forwarded-Prefix"))
}
