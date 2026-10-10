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

package catalog

import (
	"sort"
	"testing"

	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRules(t *testing.T) {
	tests := []struct {
		name     string
		give     corerulesetpb.Version
		wantRuby bool // RESPONSE-956 only exists since 4.17.0
	}{
		{name: "4.16.0", give: corerulesetpb.Version_VERSION_V4_16_0},
		{
			name: "4.17.0", give: corerulesetpb.Version_VERSION_V4_17_0,
			wantRuby: true,
		},
		{
			name: "unspecified is the latest",
			give: corerulesetpb.Version_VERSION_UNSPECIFIED, wantRuby: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := Rules(tt.give)
			require.NotEmpty(t, rules)

			ids := make([]int, 0, len(rules))
			hasRuby := false
			for _, rule := range rules {
				ids = append(ids, int(rule.GetId()))
				hasRuby = hasRuby || rule.GetCategory() ==
					corerulesetpb.Category_CATEGORY_DATA_LEAKAGES_RUBY
			}

			assert.True(t, sort.IntsAreSorted(ids))
			assert.Equal(t, tt.wantRuby, hasRuby)
		})
	}
}

func TestRule(t *testing.T) {
	rule, ok := Rule(corerulesetpb.Version_VERSION_V4_17_0, 942100)
	require.True(t, ok)
	assert.Equal(t, corerulesetpb.Category_CATEGORY_SQLI, rule.GetCategory())
	assert.False(t, rule.GetSystem())

	_, ok = Rule(corerulesetpb.Version_VERSION_V4_17_0, 1)
	assert.False(t, ok)
}
