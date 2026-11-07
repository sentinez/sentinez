// Copyright 2025 Duc-Hung Ho.
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

package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/stmt"
)

// syncSchema creates tableName, adds the missing columns of schema, then
// its indexes, an index needs its column; in one round trip, run as a
// single implicit transaction.
func syncSchema(ctx context.Context, client Client,
	tableName string, schema dbx.Schema) error {

	stmts := make([]string, 0, 1+len(schema.Columns)+len(schema.Indexes))
	stmts = append(stmts, stmt.CreateTable(tableName))

	for _, column := range schema.Columns {
		stmts = append(stmts, stmt.AddColumn(tableName, column))
	}

	for _, index := range schema.Indexes {
		stmts = append(stmts, stmt.CreateIndex(tableName, index))
	}

	// no arguments: pgx sends it with the simple protocol, which allows
	// several statements.
	if _, err := client.Exec(ctx, strings.Join(stmts, ";")); err != nil {
		return fmt.Errorf("sync schema %s: %w", tableName, err)
	}

	return nil
}
