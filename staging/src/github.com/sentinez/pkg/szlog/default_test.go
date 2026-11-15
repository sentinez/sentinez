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

package szlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The level is LevelDebug until SetLogLevel, which runs once per process.

func TestEnabled(t *testing.T) {
	tests := []struct {
		name        string
		giveLevel   Level
		wantEnabled bool
	}{
		{"debug", LevelDebug, true},
		{"error", LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantEnabled, Enabled(tt.giveLevel))
		})
	}
}

func TestDebugFnCallsFnWhenDebugIsOn(t *testing.T) {
	called := false

	DebugFn(func() string {
		called = true

		return "message"
	})

	assert.True(t, called)
}
