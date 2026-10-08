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

package rand

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Returned IDs must not alias the pooled buffer: later calls would
// overwrite them in place.
func TestIDStableAfterNextCall(t *testing.T) {
	tests := []struct {
		name    string
		give    func() string
		wantPfx string
	}{
		{
			name:    "xid",
			give:    func() string { return NewXID([]byte("senz:req:")) },
			wantPfx: "senz:req:",
		},
		{
			name:    "time id",
			give:    func() string { return NewTimeID([]byte("senz:t:"), 1) },
			wantPfx: "senz:t:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first := tt.give()
			snapshot := strings.Clone(first)

			for range 100 {
				_ = tt.give()
			}

			assert.Equal(t, snapshot, first)
			assert.True(t, strings.HasPrefix(first, tt.wantPfx))
		})
	}
}
