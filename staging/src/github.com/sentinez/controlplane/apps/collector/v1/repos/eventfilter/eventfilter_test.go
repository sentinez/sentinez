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
	"github.com/sentinez/pkg/szstorage/dbx"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _cols = Columns{
	RequestId:   "request_id",
	ResourceId:  "resource_id",
	MatchedTime: "matched_time",
	Fields: []Field{
		TypedColumn("status", dbx.Int4),
		TypedColumn("rule_ids", dbx.TextArr),
		TypedColumn("rule_nums", dbx.IntArr),
		{Name: "action", Expr: "action->>'type'"},
	},
}

var (
	_applyFrom = time.UnixMilli(1000)
	_applyTo   = time.UnixMilli(2000)
)

var _applyCases = []struct {
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
			From:        _applyFrom,
			To:          _applyTo,
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

func TestApply(t *testing.T) {
	for _, tt := range _applyCases {
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
		{Field: "rule_nums", Value: "7"},
		{Field: "action", Value: "8"},
	}}

	b, err := give.Where(sq.Select("id").From("t"), _cols)
	require.NoError(t, err)

	gotSQL, gotArgs, err := b.ToSql()
	require.NoError(t, err)
	assert.Equal(t, "SELECT id FROM t WHERE (status) = ? "+
		"AND ? = ANY(rule_ids) AND ? = ANY(rule_nums) "+
		"AND (action->>'type') = ?", gotSQL)
	assert.Equal(t, []any{int32(429), "941100", int32(7), "8"}, gotArgs)
}

func TestMatchValue(t *testing.T) {
	tests := []struct {
		name string
		give Field
		want any
	}{
		{"text", Column("c"), "12"},
		{"int4", TypedColumn("c", dbx.Int4), int32(12)},
		{"int4 array", TypedColumn("c", dbx.IntArr), int32(12)},
		{"int8", TypedColumn("c", dbx.Int8), int64(12)},
		{"text array", TypedColumn("c", dbx.TextArr), "12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Match{Field: "c", Value: "12"}.value(tt.give)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWhereRejectsValue(t *testing.T) {
	tests := []struct {
		name string
		give Match
	}{
		{"not a number", Match{Field: "status", Value: "abc"}},
		{"overflow", Match{Field: "status", Value: "99999999999"}},
		{"array element", Match{Field: "rule_nums", Value: "x"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			give := &Filter{Matches: []Match{tt.give}}

			_, err := give.Where(sq.Select("id").From("t"), _cols)

			assert.ErrorIs(t, err, ErrInvalidValue)
		})
	}
}

func TestCond(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		got, err := (*Filter)(nil).Cond(_cols)

		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("empty", func(t *testing.T) {
		got, err := (&Filter{}).Cond(_cols)

		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("conditions", func(t *testing.T) {
		give := &Filter{
			RequestIds: []string{"req_1"},
			From:       time.UnixMilli(1000),
		}

		cond, err := give.Cond(_cols)
		require.NoError(t, err)

		gotSQL, gotArgs, err := cond.ToSql()
		require.NoError(t, err)
		assert.Equal(t, "(request_id IN (?) AND matched_time >= ?)", gotSQL)
		assert.Equal(t, []any{"req_1", int64(1000)}, gotArgs)
	})
}

func TestCondInvalid(t *testing.T) {
	give := &Filter{Matches: []Match{{Field: "status", Value: "x"}}}

	_, err := give.Cond(_cols)

	assert.ErrorIs(t, err, ErrInvalidValue)
}

func TestWhereRejectsUnknownField(t *testing.T) {
	give := &Filter{Matches: []Match{{Field: "headers", Value: "x"}}}

	_, err := give.Where(sq.Select("id").From("t"), _cols)

	assert.ErrorIs(t, err, ErrUnknownField)
}
