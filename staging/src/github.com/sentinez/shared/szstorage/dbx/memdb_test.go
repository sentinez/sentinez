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
	"maps"
	"slices"
	"testing"

	"github.com/hashicorp/go-memdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memRow struct {
	Id            string
	OwnerTenantId string
	Status        int32
	Content       []byte
}

func memTestSchema() Schema {
	return Schema{
		Table: "security.rows",
		Columns: []Column{
			{Name: "id", Type: Text, MemIndex: true},
			{Name: "owner_tenant_id", Type: Text, MemIndex: true},
			{Name: "status", Type: Int4, MemIndex: true},
			{Name: "content", Type: JSONB},
		},
	}
}

func TestGoFieldName(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{give: "id", want: "Id"},
		{give: "owner_tenant_id", want: "OwnerTenantId"},
		{give: "password_hash", want: "PasswordHash"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, goFieldName(tt.give))
		})
	}
}

func TestSchemaMemDB(t *testing.T) {
	table, err := memTestSchema().MemDB()

	require.NoError(t, err)
	assert.Equal(t, "security.rows", table.Name)
	assert.ElementsMatch(t, []string{"id", "owner_tenant_id", "status"},
		slices.Collect(maps.Keys(table.Indexes)))
	assert.True(t, table.Indexes["id"].Unique)
	assert.Equal(t, &memdb.StringFieldIndex{Field: "OwnerTenantId"},
		table.Indexes["owner_tenant_id"].Indexer)
}

func TestMemDBSchemaErrors(t *testing.T) {
	unsupported := Schema{
		Table:   "t",
		Columns: []Column{{Name: "content", Type: JSONB, MemIndex: true}},
	}

	tests := []struct {
		name    string
		give    []Schema
		wantErr string
	}{
		{
			name:    "unsupported type",
			give:    []Schema{unsupported},
			wantErr: "unsupported type JSONB",
		},
		{
			name:    "duplicate table",
			give:    []Schema{memTestSchema(), memTestSchema()},
			wantErr: "duplicate memdb table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MemDBSchema(tt.give...)

			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestMemDBSchemaLookup(t *testing.T) {
	schema, err := MemDBSchema(memTestSchema())
	require.NoError(t, err)

	db, err := memdb.NewMemDB(schema)
	require.NoError(t, err)

	txn := db.Txn(true)
	require.NoError(t, txn.Insert("security.rows",
		&memRow{Id: "1", OwnerTenantId: "t1", Status: 2}))
	require.NoError(t, txn.Insert("security.rows", &memRow{Id: "2"}))
	txn.Commit()

	tests := []struct {
		giveIndex string
		giveValue any
		wantId    string
	}{
		{giveIndex: "id", giveValue: "2", wantId: "2"},
		{giveIndex: "owner_tenant_id", giveValue: "t1", wantId: "1"},
		{giveIndex: "status", giveValue: int32(2), wantId: "1"},
	}

	for _, tt := range tests {
		t.Run(tt.giveIndex, func(t *testing.T) {
			got, err := db.Txn(false).
				First("security.rows", tt.giveIndex, tt.giveValue)

			require.NoError(t, err)
			require.IsType(t, &memRow{}, got)
			assert.Equal(t, tt.wantId, got.(*memRow).Id)
		})
	}
}
