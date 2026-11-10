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
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/pkg/szstorage/dbx"
)

// ErrUnknownField the error of a field the table of a Columns does not
// have among its Fields.
var ErrUnknownField = errors.New("unknown field")

// ErrInvalidValue the error of the value of a Match that is not one of
// the type of its field, e.g. a status that is not a number.
var ErrInvalidValue = errors.New("invalid value")

// Field a field of the events their top values count and a Match filters.
type Field struct {
	// Name of the field, its proto field name.
	Name string
	// Expr the SQL of its value, a column or an expression of the row:
	// written in the query, it must not come from a request.
	Expr string
	// Type the type of Expr, dbx.Text when empty: a Match compares its
	// value in that type, so the index of the column serves. dbx.IntArr
	// and dbx.TextArr are arrays of the integers and texts they hold.
	Type dbx.ColumnType
}

// IsArray reports whether Expr is an array whose elements are the values.
func (f Field) IsArray() bool {
	return f.Type == dbx.TextArr || f.Type == dbx.IntArr
}

// Column the Field of a TEXT column.
func Column(name string) Field {
	return Field{Name: name, Expr: name}
}

// TypedColumn the Field of a column of type typ.
func TypedColumn(name string, typ dbx.ColumnType) Field {
	return Field{Name: name, Expr: name, Type: typ}
}

// Match the events holding Value in the field Field (an element of an
// Array one), compared in the type of the field.
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

// value m.Value in the type of field, or of its elements: an int32 for
// INTEGER, an int64 for BIGINT, a bool for BOOLEAN, else the string; an
// ErrInvalidValue when it is not one of that type.
func (m Match) value(field Field) (any, error) {
	var (
		n   int64
		err error
	)

	switch field.Type {
	case dbx.Int4, dbx.IntArr:
		n, err = strconv.ParseInt(m.Value, 10, 32)
		if err == nil {
			return int32(n), nil
		}
	case dbx.Int8:
		n, err = strconv.ParseInt(m.Value, 10, 64)
		if err == nil {
			return n, nil
		}
	case dbx.Bool:
		var b bool
		b, err = strconv.ParseBool(m.Value)
		if err == nil {
			return b, nil
		}
	default:
		return m.Value, nil
	}

	return nil, fmt.Errorf("%w %q of field %s: want %s",
		ErrInvalidValue, m.Value, field.Name, field.Type)
}

// condition the SQL condition of m on field: its value, typed, is the one
// of the column, or one of its elements.
func (m Match) condition(field Field) (sq.Sqlizer, error) {
	value, err := m.value(field)
	if err != nil {
		return nil, err
	}

	if field.IsArray() {
		return sq.Expr("? = ANY("+field.Expr+")", value), nil
	}

	return sq.Expr("("+field.Expr+") = ?", value), nil
}
