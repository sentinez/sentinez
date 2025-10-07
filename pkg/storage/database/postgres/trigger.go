package postgres

import (
	"context"
	"fmt"
	"slices"
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
func syncProtoToPostgresJSONB(ctx context.Context,
	pool *pgxpool.Pool,
	table string,
	indexFields []string,
) error {
	exists, err := tableExists(ctx, pool, table)
	if err != nil {
		return fmt.Errorf("check table exists: %w", err)
	}

	if !exists {
		ddl := fmt.Sprintf(database.SchemalessF, table)
		if _, err := pool.Exec(ctx, ddl); err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}

	// Get existing indexes
	existingIndexes, err := listIndexes(ctx, pool, table)
	if err != nil {
		return fmt.Errorf("list indexes: %w", err)
	}

	// Convert indexFields to a set
	wanted := make(map[string]struct{}, len(indexFields))
	for _, f := range indexFields {
		indexName := fmt.Sprintf("idx_%s_%s", table, f)
		wanted[indexName] = struct{}{}
	}

	// Create missing indexes
	for indexName := range wanted {
		found := slices.Contains(existingIndexes, indexName)
		if !found {
			field := strings.TrimPrefix(indexName, "idx_"+table+"_")
			if err := createIndexOnJSONB(ctx, pool, table, field); err != nil {
				return fmt.Errorf("create index %s: %w", indexName, err)
			}
		}
	}

	// Drop indexes that are no longer wanted
	for _, existing := range existingIndexes {
		if _, ok := wanted[existing]; !ok {
			if err := dropIndex(ctx, pool, existing); err != nil {
				return fmt.Errorf("drop index %s: %w", existing, err)
			}
		}
	}

	return nil
}
