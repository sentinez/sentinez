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
	"testing"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/stretchr/testify/assert"
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
	routes := []*edgepb.Location{
		{Location: "/local"}, {Location: "/api"}, {Location: "/"},
	}

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

			assert.Equal(t, tt.want, got.GetLocation())
		})
	}
}

func TestPickNone(t *testing.T) {
	routes := []*edgepb.Location{{Location: "/api"}}

	assert.Nil(t, pick(routes, "", []byte("/web")))
}
