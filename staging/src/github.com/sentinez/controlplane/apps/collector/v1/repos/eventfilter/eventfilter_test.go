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

package eventfilter

import (
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _cols = Columns{ResourceId: "resource_id", MatchedTime: "matched_time"}

func TestApply(t *testing.T) {
	from := time.UnixMilli(1000)
	to := time.UnixMilli(2000)

	tests := []struct {
		name     string
		give     *Filter
		wantSQL  string
		wantArgs []any
	}{
		{
			name:    "nil",
			wantSQL: "SELECT id FROM t ORDER BY matched_time DESC",
		},
		{
			name: "every field",
			give: &Filter{
				ResourceIds: []string{"r1", "r2"},
				From:        from,
				To:          to,
				Page:        &typepb.Pages{Size: 10, Index: 3},
			},
			wantSQL: "SELECT id FROM t WHERE resource_id IN (?,?) " +
				"AND matched_time >= ? AND matched_time < ? " +
				"ORDER BY matched_time DESC LIMIT 10 OFFSET 20",
			wantArgs: []any{"r1", "r2", int64(1000), int64(2000)},
		},
		{
			name:    "page without index",
			give:    &Filter{Page: &typepb.Pages{Size: 10}},
			wantSQL: "SELECT id FROM t ORDER BY matched_time DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.give.Apply(sq.Select("id").From("t"), _cols)

			gotSQL, gotArgs, err := b.ToSql()
			require.NoError(t, err)
			assert.Equal(t, tt.wantSQL, gotSQL)
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}
