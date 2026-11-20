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
	"testing"
	"unsafe"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/stretchr/testify/assert"
)

// httpContext names the embedded Context of fakeContext, whose Context
// method a field named Context would hide.
type httpContext = Context

// fakeContext records the request headers SetUpstreamHeader sets.
type fakeContext struct {
	httpContext
	x       *edgepb.ContextExtra
	headers map[string]string
}

func (c *fakeContext) X() *edgepb.ContextExtra { return c.x }

func (c *fakeContext) SetHeader(k, v []byte) {
	c.headers[string(k)] = string(v)
}

func newFakeContext() *fakeContext {
	return &fakeContext{
		x:       &edgepb.ContextExtra{},
		headers: map[string]string{},
	}
}

func TestSetUpstreamHeader(t *testing.T) {
	ctx := newFakeContext()

	SetUpstreamHeader(ctx, "X-Real-IP", "1.2.3.4")
	SetUpstreamHeader(ctx, "X-Forwarded-Prefix", "/api")

	want := map[string]string{
		"X-Real-IP":          "1.2.3.4",
		"X-Forwarded-Prefix": "/api",
	}
	assert.Equal(t, want, ctx.headers)
	assert.Equal(t, want, ctx.x.GetSetHeaders())
}

func TestSetUpstreamHeaderLastWins(t *testing.T) {
	ctx := newFakeContext()

	SetUpstreamHeader(ctx, "X-Tier", "free")
	SetUpstreamHeader(ctx, "X-Tier", "gold")

	assert.Equal(t, map[string]string{"X-Tier": "gold"},
		ctx.x.GetSetHeaders())
	assert.Equal(t, "gold", ctx.headers["X-Tier"])
}

// immutableContext a context whose getters return immutable strings.
type immutableContext struct{ httpContext }

func (immutableContext) ImmutableRequest() {}

func TestKeep(t *testing.T) {
	tests := []struct {
		name     string
		give     Request
		wantSame bool // the string is b itself, not a copy
	}{
		{"immutable request", immutableContext{}, true},
		{"reused buffers", newFakeContext(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := []byte("a.example.com")

			got := Keep(tt.give, b)

			assert.Equal(t, "a.example.com", got)
			assert.Equal(t, tt.wantSame, unsafe.StringData(got) == &b[0])
		})
	}
}
