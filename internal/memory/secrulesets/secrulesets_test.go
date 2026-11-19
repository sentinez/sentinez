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

package secrulesets

import (
	"testing"

	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
)

// nolint:funlen
func TestRunsCrs(t *testing.T) {
	on := &rulesetpb.Crs{Mode: rulesetpb.EngineMode_ENGINE_MODE_ON}
	active := typepb.Status_STATUS_ACTIVE

	tests := []struct {
		name string
		give *rulesetpb.Setting
		want bool
	}{
		{
			name: "owasp",
			give: &rulesetpb.Setting{
				Status:   active,
				Provider: rulesetpb.Provider_PROVIDER_OWASP,
				Crs:      on,
			},
			want: true,
		},
		{
			name: "unset provider is owasp",
			give: &rulesetpb.Setting{Status: active, Crs: on},
			want: true,
		},
		{
			name: "managed provider",
			give: &rulesetpb.Setting{
				Status:   active,
				Provider: rulesetpb.Provider_PROVIDER_MANAGED,
				Crs:      on,
			},
		},
		{
			name: "not active",
			give: &rulesetpb.Setting{
				Status: typepb.Status_STATUS_DISABLE,
				Crs:    on,
			},
		},
		{
			name: "engine off",
			give: &rulesetpb.Setting{Status: active, Crs: &rulesetpb.Crs{}},
		},
		{name: "no setting"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, runsCrs(tt.give))
		})
	}
}
