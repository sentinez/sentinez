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
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// ErrUnknownField the error of a field the table of a Columns does not
// have among its Fields.
var ErrUnknownField = errors.New("unknown field")

// Field a field of the events their top values count and a Match filters.
type Field struct {
	// Name of the field, its proto field name.
	Name string
	// Expr the SQL of its value, a column or an expression of the row:
	// written in the query, it must not come from a request.
	Expr string
	// Array Expr is an array whose elements are the values.
	Array bool
}

// Column the Field of a column.
func Column(name string) Field {
	return Field{Name: name, Expr: name}
}

// ArrayColumn the Array Field of an array column.
func ArrayColumn(name string) Field {
	return Field{Name: name, Expr: name, Array: true}
}

// Match the events holding Value in the field Field (an element of an
// Array one), compared as TEXT.
type Match struct {
	Field string
	Value string
}

// Field the field of cols named name, an ErrUnknownField when none.
func (cols Columns) Field(name string) (Field, error) {
	for _, field := range cols.Fields {
		if field.Name == name {
			return field, nil
		}
	}

	return Field{}, fmt.Errorf("%w %q", ErrUnknownField, name)
}

// condition the SQL condition of m on field.
func (m Match) condition(field Field) sq.Sqlizer {
	if field.Array {
		return sq.Expr("? = ANY(("+field.Expr+")::TEXT[])", m.Value)
	}

	return sq.Expr("("+field.Expr+")::TEXT = ?", m.Value)
}
