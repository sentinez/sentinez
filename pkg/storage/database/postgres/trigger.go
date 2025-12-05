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
	"github.com/sentinez/sentinez/pkg/storage/database"
)

// tableExists checks if a table exists.
func tableExists(ctx context.Context,
	pool *pgxpool.Pool, table string) (bool, error) {

	const query = `
		SELECT 1 FROM information_schema.tables
		WHERE table_name = $1 LIMIT 1;
	`
	var dummy int
	err := pool.QueryRow(ctx, query, table).Scan(&dummy)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// nolint:funlen
func proto2JSONB(ctx context.Context, pool *pgxpool.Pool, table string) error {
	exists, err := tableExists(ctx, pool, table)
	if err != nil {
		return fmt.Errorf("check table exists err: %w", err)
	}

	if !exists {
		ddl := fmt.Sprintf(database.SchemalessF, table)
		if _, err := pool.Exec(ctx, ddl); err != nil {
			return fmt.Errorf("create table err: %w", err)
		}
	}

	return nil
}
