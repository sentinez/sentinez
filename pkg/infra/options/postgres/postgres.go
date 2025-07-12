package pgopt

import (
	"context"
	"fmt"

	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/zlog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TableKV(pool *pgxpool.Pool, name string) error {
	const queryExec = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			data JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`
	sql := fmt.Sprintf(queryExec, name)

	_, err := pool.Exec(context.Background(), sql)
	if err != nil {
		zlog.Error("[pgopt] executing err: ", err)
		return errors.F("[pgopt] failed to executing queries")
	}

	return nil
}

func Reference(pool *pgxpool.Pool,
	fromTable, fromField, toTable, toField string,
) error {
	ctx := context.Background()
	constraintName := fmt.Sprintf("fk_%s_%s", fromTable, fromField)

	const checkConstraint = `
		SELECT 1
		FROM pg_constraint
		WHERE conname = $1;
	`
	var exists int
	err := pool.QueryRow(ctx, checkConstraint, constraintName).Scan(&exists)
	if err == nil {

		zlog.Debugf("[pgopt] foreign key already exists: %s (%s.%s -> %s.%s)",
			constraintName, fromTable, fromField, toTable, toField)
		return nil
	} else if !errors.Is(err, pgx.ErrNoRows) {

		zlog.Errorf("[pgopt] check constraint error: %v", err)
		return errors.F("[pgopt] failed to check constraint existence: %w", err)
	}

	const alterQuery = `
		ALTER TABLE %s
		ADD CONSTRAINT %s FOREIGN KEY (%s)
		REFERENCES %s(%s)
		ON DELETE CASCADE
		ON UPDATE CASCADE;
	`
	sql := fmt.Sprintf(alterQuery,
		fromTable, constraintName, fromField, toTable, toField)

	_, err = pool.Exec(ctx, sql)
	if err != nil {
		zlog.Errorf("[pgopt] executing ALTER TABLE error: %v", err)
		return errors.F("[pgopt] failed to execute ALTER TABLE: %w", err)
	}

	zlog.Debugf("[pgopt] reference created: %s (%s.%s -> %s.%s)",
		constraintName, fromTable, fromField, toTable, toField)

	return nil
}

func WithStorageOption(opt database.StorageOption) database.Option {
	return func(ref *database.Table) {
		ref.StorageOption = opt
	}
}

func WithReference(fromField, toTable, toField string) database.Option {
	return func(ref *database.Table) {
		if ref.References == nil {
			ref.References = make(map[string]database.Reference)
		}

		ref.References[fromField] = database.Reference{
			FromField: fromField,
			ToTable:   toTable,
			ToField:   toField,
		}
	}
}
