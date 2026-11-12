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

// Package topvalues counts the most frequent values of the fields of an
// event table over its rows: the field list of the analytic.
package topvalues

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/pkg/szstorage/dbx"
	dbxquery "github.com/sentinez/pkg/szstorage/dbx/query"
)

var (
	errNoFields  = errors.New("no field")
	errFieldName = errors.New("invalid field name")
)

// _fieldName the names a field may have: they are written in the query,
// not bound.
var _fieldName = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// _ranked the query ranking the counts of the fields, %s: their values
// per field, the most frequent first, the empty ones left out but the
// number of rows, its field empty.
const _ranked = `SELECT field, value, total FROM (
SELECT field, value, total,
row_number() OVER (PARTITION BY field ORDER BY total DESC, value) AS pos
FROM (%s) AS counts WHERE field = '' OR value <> '') AS ranked
WHERE pos <= ? ORDER BY field, pos`

// Row a row of a query: the rows holding a value of the field Field, or
// the number of rows when Field is empty.
type Row struct {
	Field string
	Value string
	Total int64
}

// Value the rows holding a value of a field.
type Value struct {
	Value string
	Total int64
}

// FieldValues the most frequent values of a field, the most first.
type FieldValues struct {
	Field  string
	Values []*Value
}

// Result the most frequent values of fields and the number of rows
// counted.
type Result struct {
	Total int64
	// Fields one per field queried, in its order.
	Fields []*FieldValues
}

// query counts the limit most frequent values of each of fields among the
// rows b selects (its table and conditions, no column), the NULL and
// empty ones left out, then the number of rows: a Row each, per field
// then the most frequent first. The scalar fields are counted in one scan
// of the rows, the Array ones in another.
func query(b sq.SelectBuilder, fields []eventfilter.Field,
	limit uint64) (dbxquery.Query, error) {
	if err := check(fields); err != nil {
		return nil, err
	}

	var scalars, arrays []eventfilter.Field
	for _, field := range fields {
		if field.IsArray() {
			arrays = append(arrays, field)
			continue
		}
		scalars = append(scalars, field)
	}

	parts := []sq.SelectBuilder{scalarCounts(b, scalars)}
	if len(arrays) > 0 {
		parts = append(parts, arrayCounts(b, arrays))
	}

	return &rankedQuery{parts: parts, limit: limit}, nil
}

// collect scans the rows of a query.
func collect(rows dbx.Rows) ([]*Row, error) {
	var collected []*Row
	for rows.Next() {
		var row Row
		if err := rows.Scan(&row.Field, &row.Value, &row.Total); err != nil {
			return nil, err
		}
		collected = append(collected, &row)
	}

	return collected, rows.Err()
}

// group the Result of the rows of a query of fields.
func group(rows []*Row, fields []eventfilter.Field) *Result {
	result := &Result{Fields: make([]*FieldValues, 0, len(fields))}
	byName := make(map[string]*FieldValues, len(fields))
	for _, field := range fields {
		values := &FieldValues{Field: field.Name}
		result.Fields = append(result.Fields, values)
		byName[field.Name] = values
	}

	for _, row := range rows {
		if row.Field == "" {
			result.Total = row.Total
			continue
		}
		if values, ok := byName[row.Field]; ok {
			values.Values = append(values.Values,
				&Value{Value: row.Value, Total: row.Total})
		}
	}

	return result
}

// pick the fields of names among the Fields of cols, in their order, each
// once; every one of them when names is empty. A name not among them is
// an eventfilter.ErrUnknownField.
func pick(cols eventfilter.Columns,
	names []string) ([]eventfilter.Field, error) {
	if len(names) == 0 {
		return cols.Fields, nil
	}

	picked := make([]eventfilter.Field, 0, len(names))
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		field, err := cols.Field(name)
		if err != nil {
			return nil, err
		}
		if !seen[name] {
			seen[name] = true
			picked = append(picked, field)
		}
	}

	return picked, nil
}

// rankedQuery the counts of the fields, parts, ranked: the query of
// query.
type rankedQuery struct {
	parts []sq.SelectBuilder
	limit uint64
}

// ToSql implements dbxquery.Query: the parts united, then ranked.
func (q *rankedQuery) ToSql() (string, []any, error) {
	stmts := make([]string, 0, len(q.parts))
	var args []any
	for _, part := range q.parts {
		stmt, partArgs, err := part.ToSql()
		if err != nil {
			return "", nil, err
		}
		stmts = append(stmts, stmt)
		args = append(args, partArgs...)
	}

	stmt := fmt.Sprintf(_ranked, strings.Join(stmts, " UNION ALL "))
	return stmt, append(args, q.limit), nil
}

// scalarCounts counts the values of fields, none Array, in one scan of
// the rows of b, then the rows: one grouping set per field, then the one
// of every row.
func scalarCounts(
	b sq.SelectBuilder, fields []eventfilter.Field,
) sq.SelectBuilder {
	if len(fields) == 0 {
		return b.Columns("'' AS field", "'' AS value", "count(*) AS total")
	}

	whens := make([]string, 0, len(fields))
	casts := make([]string, 0, len(fields)+1)
	sets := make([]string, 0, len(fields)+1)
	for _, field := range fields {
		whens = append(whens, fmt.Sprintf("WHEN GROUPING(%s) = 0 THEN '%s'",
			field.Expr, field.Name))
		casts = append(casts, "("+field.Expr+")::TEXT")
		sets = append(sets, "("+field.Expr+")")
	}

	return b.Columns(
		"CASE "+strings.Join(whens, " ")+" ELSE '' END AS field",
		"COALESCE("+strings.Join(append(casts, "''"), ", ")+") AS value",
		"count(*) AS total",
	).GroupBy("GROUPING SETS (" + strings.Join(append(sets, "()"), ", ") +
		")")
}

// arrayCounts counts, in one scan of the rows of b, the rows holding each
// distinct element of each of the Array fields.
func arrayCounts(
	b sq.SelectBuilder, fields []eventfilter.Field,
) sq.SelectBuilder {
	elements := make([]string, 0, len(fields))
	for _, field := range fields {
		elements = append(elements, fmt.Sprintf("SELECT '%s' AS field, "+
			"element::TEXT AS element FROM (SELECT DISTINCT unnest(%s) "+
			"AS element) AS distinct_elements", field.Name, field.Expr))
	}

	return b.Columns(
		"elements.field",
		"COALESCE(elements.element, '') AS value",
		"count(*) AS total",
	).CrossJoin("LATERAL ("+strings.Join(elements, " UNION ALL ")+
		") AS elements").
		GroupBy("elements.field", "elements.element")
}

// check returns an error unless fields has one field at least, each of a
// valid name.
func check(fields []eventfilter.Field) error {
	if len(fields) == 0 {
		return errNoFields
	}

	for _, field := range fields {
		if !_fieldName.MatchString(field.Name) {
			return fmt.Errorf("%w %q", errFieldName, field.Name)
		}
	}

	return nil
}
