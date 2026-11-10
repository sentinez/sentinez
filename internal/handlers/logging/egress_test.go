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

package logging

import (
	"testing"

	"github.com/sentinez/core/szhttp"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/sentinez/pkg/pools/egress"
	"github.com/stretchr/testify/assert"
)

// httpContext names the embedded szhttp.Context of fakeContext, whose
// Context method a field named Context would hide.
type httpContext = szhttp.Context

// fakeContext answers the calls of newEgressEvent; the others panic.
type fakeContext struct {
	httpContext
	x       *edgepb.ContextExtra
	headers [][2]string
}

func (c *fakeContext) X() *edgepb.ContextExtra { return c.x }
func (*fakeContext) RequestId() string         { return "req_1" }
func (*fakeContext) Method() []byte            { return []byte("GET") }
func (*fakeContext) Protocol() string          { return "HTTP/2.0" }
func (*fakeContext) Path() []byte              { return []byte("/a") }
func (*fakeContext) QueryStr() []byte          { return []byte("q=1") }
func (*fakeContext) StatusCode() int           { return 201 }
func (*fakeContext) ResourceId() string        { return "res_1" }
func (*fakeContext) TenantId() string          { return "ten_1" }

func (c *fakeContext) VisitResponseHeaders(visitor func(k, v []byte)) {
	for _, h := range c.headers {
		visitor([]byte(h[0]), []byte(h[1]))
	}
}

func TestNewEgressEvent(t *testing.T) {
	ctx := &fakeContext{
		x: &edgepb.ContextExtra{
			Upstream:              "http://origin:8080",
			UpstreamTimestamp:     5000,
			UpstreamLatencyMicros: 1500,
		},
		headers: [][2]string{
			{"content-type", "text/html"},
			{"Content-Length", "42"},
			{"X-Other", "x"},
		},
	}

	event := newEgressEvent(ctx, ctx.x)
	defer egress.Release(event)

	assert.Equal(t, "req_1", event.Id)
	assert.Equal(t, "http://origin:8080", event.Upstream)
	assert.Equal(t, "GET", event.Method)
	assert.Equal(t, "HTTP/2.0", event.Protocol)
	assert.Equal(t, "http", event.Scheme)
	assert.Equal(t, "/a", event.Path)
	assert.Equal(t, "q=1", event.Queries)
	assert.Equal(t, int32(201), event.Status)
	assert.Equal(t, "text/html", event.ContentType)
	assert.Equal(t, int64(42), event.ContentLength)
	assert.Equal(t, int64(1500), event.LatencyMicros)
	assert.Equal(t, int64(5000), event.Timestamp)
	assert.Equal(t, "res_1", event.ResourceId)
	assert.Equal(t, "ten_1", event.TenantId)
}

func TestNewEgressEventNoContentLength(t *testing.T) {
	ctx := &fakeContext{x: &edgepb.ContextExtra{Upstream: "http://o"}}

	event := newEgressEvent(ctx, ctx.x)
	defer egress.Release(event)

	assert.Equal(t, int64(_unknownLength), event.ContentLength)
}

func TestContentLength(t *testing.T) {
	tests := []struct {
		give string
		want int64
	}{
		{give: "0", want: 0},
		{give: "1024", want: 1024},
		{give: "-3", want: _unknownLength},
		{give: "abc", want: _unknownLength},
		{give: "", want: _unknownLength},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, contentLength([]byte(tt.give)))
		})
	}
}

func TestSetHeadersOf(t *testing.T) {
	tests := []struct {
		name string
		give map[string]string
		want string
	}{
		{"none", nil, ""},
		{
			"sorted by name",
			map[string]string{
				"X-Real-IP":          "1.2.3.4",
				"X-Forwarded-Prefix": "/api",
			},
			"X-Forwarded-Prefix: /api\nX-Real-IP: 1.2.3.4\n",
		},
		{
			"credentials left out",
			map[string]string{
				"Authorization": "Bearer secret",
				"cookie":        "a=b",
				"X-Env":         "prod",
			},
			"X-Env: prod\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := &edgepb.ContextExtra{SetHeaders: tt.give}

			assert.Equal(t, tt.want, setHeadersOf(x))
		})
	}
}

func TestNewEgressEventHeaders(t *testing.T) {
	ctx := &fakeContext{x: &edgepb.ContextExtra{
		Upstream:   "http://o",
		SetHeaders: map[string]string{"X-Env": "prod"},
	}}

	event := newEgressEvent(ctx, ctx.x)
	defer egress.Release(event)

	assert.Equal(t, "X-Env: prod\n", event.Headers)
}

func TestEgressEventReleaseClearsHeaders(t *testing.T) {
	event := egress.Acquire()
	event.Headers = "X-Env: prod\n"

	egress.Release(event)

	assert.Empty(t, egress.Acquire().Headers)
}
