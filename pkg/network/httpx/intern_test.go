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

package httpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderKey(t *testing.T) {
	tests := []struct {
		name string
		give string
		want string
	}{
		{"interned", "Cache-Control", "Cache-Control"},
		{"interned, canonical", "Etag", "Etag"},
		{"interned, as declared", "ETag", "Etag"},
		{"unknown, copied as is", "x-custom", "x-custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, headerKey([]byte(tt.give)))
		})
	}
}

func TestHeaderValue(t *testing.T) {
	tests := []struct {
		name string
		give string
	}{
		{"interned", "HIT"},
		{"unknown", "max-age=60"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.give, headerValue([]byte(tt.give)))
		})
	}
}
