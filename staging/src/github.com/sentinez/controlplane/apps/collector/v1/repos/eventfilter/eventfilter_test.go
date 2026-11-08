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

var _cols = Columns{
	RequestId:   "request_id",
	ResourceId:  "resource_id",
	MatchedTime: "matched_time",
	Fields: []Field{
		Column("status"),
		ArrayColumn("rule_ids"),
		{Name: "action", Expr: "action->>'type'"},
	},
}

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
				RequestIds:  []string{"req_1"},
				ResourceIds: []string{"r1", "r2"},
				From:        from,
				To:          to,
				Page:        &typepb.Pages{Size: 10, Index: 3},
			},
			wantSQL: "SELECT id FROM t WHERE request_id IN (?) " +
				"AND resource_id IN (?,?) " +
				"AND matched_time >= ? AND matched_time < ? " +
				"ORDER BY matched_time DESC LIMIT 10 OFFSET 20",
			wantArgs: []any{"req_1", "r1", "r2", int64(1000), int64(2000)},
		},
		{
			name:    "page without index",
			give:    &Filter{Page: &typepb.Pages{Size: 10}},
			wantSQL: "SELECT id FROM t ORDER BY matched_time DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := tt.give.Apply(sq.Select("id").From("t"), _cols)
			require.NoError(t, err)

			gotSQL, gotArgs, err := b.ToSql()
			require.NoError(t, err)
			assert.Equal(t, tt.wantSQL, gotSQL)
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}

func TestWhereSkipsOrderAndPage(t *testing.T) {
	give := &Filter{
		ResourceIds: []string{"r1"},
		From:        time.UnixMilli(1000),
		Page:        &typepb.Pages{Size: 10, Index: 3},
	}

	b, err := give.Where(sq.Select("id").From("t"), _cols)
	require.NoError(t, err)

	gotSQL, gotArgs, err := b.ToSql()
	require.NoError(t, err)
	assert.Equal(t,
		"SELECT id FROM t WHERE resource_id IN (?) AND matched_time >= ?",
		gotSQL)
	assert.Equal(t, []any{"r1", int64(1000)}, gotArgs)
}

func TestWhereMatches(t *testing.T) {
	give := &Filter{Matches: []Match{
		{Field: "status", Value: "429"},
		{Field: "rule_ids", Value: "941100"},
		{Field: "action", Value: "8"},
	}}

	b, err := give.Where(sq.Select("id").From("t"), _cols)
	require.NoError(t, err)

	gotSQL, gotArgs, err := b.ToSql()
	require.NoError(t, err)
	assert.Equal(t, "SELECT id FROM t WHERE (status)::TEXT = ? "+
		"AND ? = ANY((rule_ids)::TEXT[]) "+
		"AND (action->>'type')::TEXT = ?", gotSQL)
	assert.Equal(t, []any{"429", "941100", "8"}, gotArgs)
}

func TestWhereRejectsUnknownField(t *testing.T) {
	give := &Filter{Matches: []Match{{Field: "headers", Value: "x"}}}

	_, err := give.Where(sq.Select("id").From("t"), _cols)

	assert.ErrorIs(t, err, ErrUnknownField)
}
