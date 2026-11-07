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

package dbx

import "github.com/sentinez/pkg/sztable"

// Schema describes a table. On startup the table is created, then the
// missing columns are added in order, then the indexes are created; the
// id, created_at and updated_at columns are always present.
type Schema struct {
	Table   sztable.Table
	Columns []Column
	Indexes []Index
}

// ColumnNames the names of the columns of s, in order.
func (s Schema) ColumnNames() []string {
	names := make([]string, len(s.Columns))
	for i, column := range s.Columns {
		names[i] = column.Name
	}

	return names
}

// Column a typed column of a Schema; MemIndex also indexes it in the
// go-memdb table derived by Schema.MemDB.
type Column struct {
	Name     string
	Type     ColumnType
	MemIndex bool
}

// ColumnType SQL type of a column, ex: TEXT.
type ColumnType string

const (
	Timestamp ColumnType = "TIMESTAMPTZ"
	Text      ColumnType = "TEXT"
	TextArr   ColumnType = "TEXT[]"
	IntArr    ColumnType = "INTEGER[]"
	ByteA     ColumnType = "BYTEA"
	Int4      ColumnType = "INTEGER"
	Int8      ColumnType = "BIGINT"
	Float4    ColumnType = "REAL"
	Float8    ColumnType = "DOUBLE PRECISION"
	Bool      ColumnType = "BOOLEAN"
	JSONB     ColumnType = "JSONB"
)

// Index a single column index of a Schema.
type Index struct {
	Column string
	Method IndexMethod
}

// IndexMethod access method of an index, ex: GIN.
type IndexMethod string

const (
	BTree IndexMethod = "BTREE"
	// GIN indexes the keys and values of a JSONB column, it serves the
	// containment operator (@>).
	GIN IndexMethod = "GIN"
)
