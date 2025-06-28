package pgopt

import (
	"context"
	"fmt"

	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/zlog"

	"github.com/jackc/pgx/v5/pgxpool"
)

const queryExec = `
CREATE TABLE IF NOT EXISTS %s (
	id TEXT PRIMARY KEY,
	data JSONB NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
`

func TableKV(pool *pgxpool.Pool, name string) error {
	sql := fmt.Sprintf(queryExec, name)
	zlog.Debug("[postgres] executing", sql)

	_, err := pool.Exec(context.Background(), sql)
	if err != nil {
		zlog.Error("[postgres] executing err: ", err)
		return errors.F("[postgres] failed to executing queries")
	}

	return nil
}
