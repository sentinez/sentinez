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

package szprotobuf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatches(t *testing.T) {
	tests := []struct {
		name      string
		giveA     []string
		giveB     []string
		wantPairs [][2]int
	}{
		{"same", []string{"a", "b"}, []string{"a", "b"},
			[][2]int{{0, 0}, {1, 1}, {2, 2}}},
		{"changed", []string{"a", "b", "c"}, []string{"a", "x", "c"},
			[][2]int{{0, 0}, {2, 2}, {3, 3}}},
		{"moved", []string{"a", "b", "c"}, []string{"b", "c", "a"},
			[][2]int{{1, 0}, {2, 1}, {3, 3}}},
		{"empty", nil, []string{"a"}, [][2]int{{0, 1}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantPairs, matches(tt.giveA, tt.giveB))
		})
	}
}
