package postgres

import (
	"context"
	"fmt"
	"strings"

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

// listIndexes returns the list of existing JSONB field indexes.
func listIndexes(ctx context.Context,
	pool *pgxpool.Pool, table string) ([]string, error) {

	const query = `
		SELECT indexname
		FROM pg_indexes
		WHERE tablename = $1;
	`
	rows, err := pool.Query(ctx, query, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []string
	for rows.Next() {
		var index string
		if err := rows.Scan(&index); err != nil {
			return nil, err
		}
		if strings.HasPrefix(index, "idx_"+table+"_") {
			indexes = append(indexes, index)
		}
	}

	return indexes, nil
}

// createIndexOnJSONB creates index if not exists.
func createIndexOnJSONB(ctx context.Context,
	pool *pgxpool.Pool, table, field string) error {

	indexName := fmt.Sprintf("idx_%s_%s", table, field)
	ddl := fmt.Sprintf(`CREATE INDEX %s ON %s ((data->>'%s'));`,
		indexName, table, field)
	_, err := pool.Exec(ctx, ddl)

	return err
}

// dropIndex drops a given index.
func dropIndex(ctx context.Context,
	pool *pgxpool.Pool, indexName string) error {
	_, err := pool.Exec(ctx,
		fmt.Sprintf(`DROP INDEX IF EXISTS %s;`, indexName))

	return err
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
