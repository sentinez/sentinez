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
