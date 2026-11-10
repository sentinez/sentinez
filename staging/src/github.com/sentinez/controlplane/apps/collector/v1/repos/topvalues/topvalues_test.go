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

package topvalues

import (
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_method = eventfilter.Field{Name: "method", Expr: "method"}
	_action = eventfilter.Field{Name: "action", Expr: "action->>'type'"}
	_rules  = eventfilter.TypedColumn("rule_ids", dbx.TextArr)
	_tags   = eventfilter.TypedColumn("tags", dbx.TextArr)
)

// rankedOf the SQL of a query of the parts.
func rankedOf(parts string) string {
	return fmt.Sprintf(_ranked, parts)
}

func TestQueryScalars(t *testing.T) {
	b := sq.Select().From("t").Where(sq.GtOrEq{"ts": 1000})

	got, err := query(b, []eventfilter.Field{_method, _action}, 5)
	require.NoError(t, err)

	gotSQL, gotArgs, err := got.ToSql()
	require.NoError(t, err)
	assert.Equal(t, rankedOf(
		"SELECT CASE WHEN GROUPING(method) = 0 THEN 'method' "+
			"WHEN GROUPING(action->>'type') = 0 THEN 'action' "+
			"ELSE '' END AS field, "+
			"COALESCE((method)::TEXT, (action->>'type')::TEXT, '') AS value, "+
			"count(*) AS total FROM t WHERE ts >= ? "+
			"GROUP BY GROUPING SETS ((method), (action->>'type'), ())"),
		gotSQL)
	assert.Equal(t, []any{1000, uint64(5)}, gotArgs)
}

func TestQueryArrays(t *testing.T) {
	b := sq.Select().From("t").Where(sq.GtOrEq{"ts": 1000})

	got, err := query(b, []eventfilter.Field{_rules, _tags}, 3)
	require.NoError(t, err)

	gotSQL, gotArgs, err := got.ToSql()
	require.NoError(t, err)
	assert.Equal(t, rankedOf(
		"SELECT '' AS field, '' AS value, count(*) AS total "+
			"FROM t WHERE ts >= ? UNION ALL "+
			"SELECT elements.field, COALESCE(elements.element, '') AS value, "+
			"count(*) AS total FROM t CROSS JOIN LATERAL ("+
			"SELECT 'rule_ids' AS field, element::TEXT AS element FROM "+
			"(SELECT DISTINCT unnest(rule_ids) AS element) AS distinct_elements"+
			" UNION ALL "+
			"SELECT 'tags' AS field, element::TEXT AS element FROM "+
			"(SELECT DISTINCT unnest(tags) AS element) AS distinct_elements"+
			") AS elements WHERE ts >= ? "+
			"GROUP BY elements.field, elements.element"),
		gotSQL)
	assert.Equal(t, []any{1000, 1000, uint64(3)}, gotArgs)
}

func TestQueryRejectsFields(t *testing.T) {
	tests := []struct {
		name    string
		give    []eventfilter.Field
		wantErr error
	}{
		{"none", nil, errNoFields},
		{
			"injection",
			[]eventfilter.Field{{Name: "method'; DROP TABLE t", Expr: "method"}},
			errFieldName,
		},
		{"quoted", []eventfilter.Field{{Name: "'method'", Expr: "method"}}, errFieldName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := query(sq.Select().From("t"), tt.give, 5)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestPick(t *testing.T) {
	fields := []eventfilter.Field{_method, _action, _rules}
	cols := eventfilter.Columns{Fields: fields}

	tests := []struct {
		name string
		give []string
		want []eventfilter.Field
	}{
		{"all", nil, fields},
		{"in order", []string{"rule_ids", "method"}, []eventfilter.Field{_rules, _method}},
		{"each once", []string{"action", "action"}, []eventfilter.Field{_action}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pick(cols, tt.give)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPickRejectsUnknown(t *testing.T) {
	_, err := pick(eventfilter.Columns{
		Fields: []eventfilter.Field{_method},
	}, []string{"headers"})

	assert.ErrorIs(t, err, eventfilter.ErrUnknownField)
}

func TestGroup(t *testing.T) {
	rows := []*Row{
		{Value: "", Total: 10},
		{Field: "method", Value: "GET", Total: 7},
		{Field: "method", Value: "POST", Total: 3},
	}

	got := group(rows, []eventfilter.Field{_rules, _method})

	assert.Equal(t, int64(10), got.Total)
	assert.Equal(t, []*FieldValues{
		{Field: "rule_ids"},
		{Field: "method", Values: []*Value{
			{Value: "GET", Total: 7},
			{Value: "POST", Total: 3},
		}},
	}, got.Fields)
}
