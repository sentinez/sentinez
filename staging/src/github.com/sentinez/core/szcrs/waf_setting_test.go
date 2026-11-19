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

package szcrs

import (
	"testing"

	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
	"github.com/stretchr/testify/assert"
)

func TestEnabled(t *testing.T) {
	tests := []struct {
		give rulesetpb.EngineMode
		want bool
	}{
		{give: rulesetpb.EngineMode_ENGINE_MODE_ON, want: true},
		{give: rulesetpb.EngineMode_ENGINE_MODE_DETECTION_ONLY, want: true},
		{give: rulesetpb.EngineMode_ENGINE_MODE_OFF},
		{give: rulesetpb.EngineMode_ENGINE_MODE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.give.String(), func(t *testing.T) {
			got := Enabled(&rulesetpb.Crs{Mode: tt.give})

			assert.Equal(t, tt.want, got)
		})
	}
}
