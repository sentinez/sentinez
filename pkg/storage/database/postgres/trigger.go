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
