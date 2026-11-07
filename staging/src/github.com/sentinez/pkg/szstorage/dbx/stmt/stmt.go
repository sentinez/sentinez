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
	"fmt"
	"strings"

	"github.com/sentinez/pkg/szstorage/dbx"
)

const (
	// the id column, then its constraint, then the timestamps.
	_createSchemaStmt = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT %s,
			created_at TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'UTC')
		)
	`

	_addColumnStmt = `
		ALTER TABLE %s
		ADD COLUMN IF NOT EXISTS %s %s
	`

	// index name: <table>_<column>_<method>_idx
	_createIndexStmt = `
		CREATE INDEX IF NOT EXISTS %s_%s_%s_idx
		ON %s USING %s (%s)
	`
)

// CreateTable creates tableName, id its primary key.
func CreateTable(tableName string) string {
	return fmt.Sprintf(_createSchemaStmt, tableName, "PRIMARY KEY")
}

// CreateKeylessTable creates tableName like CreateTable, without primary
// key: a hypertable, whose unique keys must hold its time column.
func CreateKeylessTable(tableName string) string {
	return fmt.Sprintf(_createSchemaStmt, tableName, "NOT NULL")
}

func AddColumn(tableName string, column dbx.Column) string {
	return fmt.Sprintf(_addColumnStmt, tableName, column.Name, column.Type)
}

func CreateIndex(tableName string, index dbx.Index) string {
	method := string(index.Method)

	return fmt.Sprintf(_createIndexStmt,
		tableName, index.Column, strings.ToLower(method),
		tableName, method, index.Column,
	)
}
