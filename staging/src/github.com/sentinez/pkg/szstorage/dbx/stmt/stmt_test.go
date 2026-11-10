// Copyright 2025 Sentinéz Labs.
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

package stmt

import (
	"strings"
	"testing"

	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/stretchr/testify/assert"
)

// createIndexTest a case of TestCreateIndex.
type createIndexTest struct {
	name      string
	giveTable string
	give      dbx.Index
	want      string
}

func TestCreateIndex(t *testing.T) {
	for _, tt := range createIndexTests() {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateIndex(tt.giveTable, tt.give)

			assert.Equal(t, tt.want, strings.Join(strings.Fields(got), " "))
		})
	}
}

func createIndexTests() []createIndexTest {
	return []createIndexTest{
		{
			name:      "gin",
			giveTable: "dev_sentinez_security_coreruleset",
			give:      dbx.Index{Column: "content", Method: "GIN"},
			want: "CREATE INDEX IF NOT EXISTS " +
				"dev_sentinez_security_coreruleset_content_gin_idx " +
				"ON dev_sentinez_security_coreruleset USING GIN (content)",
		},
		{
			name:      "btree",
			giveTable: "t",
			give:      dbx.Index{Column: "name", Method: "BTREE"},
			want: "CREATE INDEX IF NOT EXISTS t_name_btree_idx " +
				"ON t USING BTREE (name)",
		},
		{
			name:      "composite",
			giveTable: "t",
			give: dbx.Index{
				Columns: []string{"resource_id", "ts"},
				Method:  "BTREE",
			},
			want: "CREATE INDEX IF NOT EXISTS t_resource_id_ts_btree_idx " +
				"ON t USING BTREE (resource_id, ts)",
		},
	}
}

func TestAddColumn(t *testing.T) {
	tests := []struct {
		name      string
		giveTable string
		give      dbx.Column
		want      string
	}{
		{
			name:      "jsonb",
			giveTable: "t",
			give:      dbx.Column{Name: "content", Type: "JSONB"},
			want:      "ALTER TABLE t ADD COLUMN IF NOT EXISTS content JSONB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddColumn(tt.giveTable, tt.give)

			assert.Equal(t, tt.want, strings.Join(strings.Fields(got), " "))
		})
	}
}

func TestCreateTable(t *testing.T) {
	tests := []struct {
		name       string
		giveSchema dbx.Schema
		want       string
	}{
		{
			name: "default timestamps",
			want: "CREATE TABLE IF NOT EXISTS t ( id TEXT PRIMARY KEY, " +
				"created_at " + _timestampDDL + ", " +
				"updated_at " + _timestampDDL + " )",
		},
		{
			name: "schema timestamps",
			giveSchema: dbx.Schema{
				CreatedAt: "owner_created_at",
				UpdatedAt: "owner_updated_at",
			},
			want: "CREATE TABLE IF NOT EXISTS t ( id TEXT PRIMARY KEY, " +
				"owner_created_at " + _timestampDDL + ", " +
				"owner_updated_at " + _timestampDDL + " )",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateTable("t", tt.giveSchema)

			assert.Equal(t, tt.want, strings.Join(strings.Fields(got), " "))
		})
	}
}

const _timestampDDL = "TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'UTC')"
