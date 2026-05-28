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

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sentinez/core/storage/dbx"
	"github.com/sentinez/core/storage/dbx/stmt"
)

func syncOption(ctx context.Context,
	pool *pgxpool.Pool, option *dbx.Table) error {

	createTableStmt := stmt.CreateTable(option.Table)
	if _, err := pool.Exec(ctx, createTableStmt); err != nil {
		return fmt.Errorf("create table err: %w", err)
	}

	for column, columnType := range option.Column {
		addColumnStmt := stmt.AddColumn(option.Table, column, columnType)
		if _, err := pool.Exec(ctx, addColumnStmt); err != nil {
			return fmt.Errorf("add column err: %w", err)
		}
	}

	return nil
}
