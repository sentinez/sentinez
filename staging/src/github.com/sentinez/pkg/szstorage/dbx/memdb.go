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

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-memdb"
)

// MemIdIndex the unique index on the Id field of every go-memdb table,
// the one go-memdb requires.
const MemIdIndex = "id"

// MemDBSchema the go-memdb schema of schemas, one table each, validated.
func MemDBSchema(schemas ...Schema) (*memdb.DBSchema, error) {
	db := &memdb.DBSchema{
		Tables: make(map[string]*memdb.TableSchema, len(schemas)),
	}

	for _, schema := range schemas {
		table, err := schema.MemDB()
		if err != nil {
			return nil, err
		}

		if _, ok := db.Tables[table.Name]; ok {
			return nil, fmt.Errorf("duplicate memdb table %s", table.Name)
		}

		db.Tables[table.Name] = table
	}

	if err := db.Validate(); err != nil {
		return nil, fmt.Errorf("memdb schema: %w", err)
	}

	return db, nil
}

// MemDB the go-memdb table of s, named after s.Table: a unique "id" index
// on the Id field plus one index per MemIndex column, named after the
// column, on its Go field (owner_tenant_id: OwnerTenantId). An integer
// lookup takes the field's Go type, ex: int32(2) or an enum, not 2.
func (s Schema) MemDB() (*memdb.TableSchema, error) {
	table := &memdb.TableSchema{
		Name: s.Table.String(),
		Indexes: map[string]*memdb.IndexSchema{
			MemIdIndex: {
				Name:    MemIdIndex,
				Unique:  true,
				Indexer: &memdb.StringFieldIndex{Field: "Id"},
			},
		},
	}

	for _, column := range s.Columns {
		if !column.MemIndex || column.Name == MemIdIndex {
			continue
		}

		indexer := newMemIndexer(column.Type, goFieldName(column.Name))
		if indexer == nil {
			return nil, fmt.Errorf("memdb index %s.%s: unsupported type %s",
				table.Name, column.Name, column.Type)
		}

		table.Indexes[column.Name] = &memdb.IndexSchema{
			Name:         column.Name,
			AllowMissing: true,
			Indexer:      indexer,
		}
	}

	return table, nil
}

// newMemIndexer the go-memdb indexer of a column type, nil when go-memdb
// cannot index it, ex: JSONB.
func newMemIndexer(columnType ColumnType, field string) memdb.Indexer {
	switch columnType {
	case Text:
		return &memdb.StringFieldIndex{Field: field}
	case TextArr:
		return &memdb.StringSliceFieldIndex{Field: field}
	case Int4, Int8:
		return &memdb.IntFieldIndex{Field: field}
	case Bool:
		return &memdb.BoolFieldIndex{Field: field}
	default:
		return nil
	}
}

// goFieldName the protoc-gen-go field name of a snake_case column.
func goFieldName(column string) string {
	parts := strings.Split(column, "_")
	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}

	return strings.Join(parts, "")
}
