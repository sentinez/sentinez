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
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/sentinez/core/storage/dbx"
	"github.com/sentinez/core/storage/dbx/query"
	storageutils "github.com/sentinez/core/storage/utils"
	"github.com/sentinez/core/storage/utils/table"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/zlog"
)

var _ dbx.Database[typepb.Empty] = (*postgres[typepb.Empty])(nil)

var (
	pool *pgxpool.Pool
	lock sync.Mutex
)

func getConnPool(conf *confpb.EnvConfig) (*pgxpool.Pool, error) {
	if pool == nil {
		var err error
		lock.Lock()
		defer lock.Unlock()

		pool, err = storageutils.NewPgxPool(conf)
		if err != nil {
			return nil, err
		}
	}

	return pool, nil
}

//nolint:funlen
func New[T any](ctx context.Context, conf *confpb.Config,
	opts ...dbx.Option) (dbx.Database[T], error) {
	tb := dbx.Table{}
	for _, opt := range opts {
		opt(&tb)
	}

	tb.Table = table.NewTable(conf, tb.Table)

	conn, err := getConnPool(conf.GetEnv())
	if err != nil {
		return nil, err
	}

	if !table.IsValidTableName(tb.Table) {
		return nil, fmt.Errorf("invalid table name: %s", tb.Table)
	}

	if err := syncOption(ctx, conn, &tb); err != nil {
		return nil, err
	}

	return &postgres[T]{client: conn, tableName: tb.Table}, nil
}

type postgres[T any] struct {
	client    Client
	tableName string
	tx        bool
}

func (p *postgres[T]) Table() string {
	return p.tableName
}

func (p *postgres[T]) Total(ctx context.Context) (total int64, err error) {
	q := sq.Select("COUNT(*) AS count").From(p.tableName)

	err = p.Query(ctx, q, &total)
	if err != nil {
		return -1, err
	}

	return total, nil
}

func (p *postgres[T]) Insert(
	ctx context.Context, builder query.Query) (string, error) {

	sql, args, err := builder.ToSql()
	if err != nil {
		return "", err
	}

	zlog.Debugf("postgres: in tx=%t exec insert: %s", p.tx, sql)

	var id string
	if err := p.client.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (p *postgres[T]) Select(ctx context.Context,
	builder query.Query, scan dbx.ScanOneFn[T]) (*T, error) {
	return p.CollectOneRow(ctx, builder, scan)
}

func (p *postgres[T]) Delete(ctx context.Context, id string) error {
	builder := sq.Delete(p.tableName).Where(sq.Eq{
		dbx.FieldID: id,
	})

	_, err := p.Exec(ctx, builder)
	return err
}

// CollectRows implements database.Database.
func (p *postgres[T]) CollectRows(ctx context.Context,
	builder query.Query,
	fn dbx.ScanFn[T]) ([]*T, error) {

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	stmt = sqlx.Rebind(sqlx.DOLLAR, stmt)
	zlog.Debugf("postgres: in tx=%t query: %s", p.tx, stmt)

	rows, err := p.client.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if fn != nil {
		return fn(rows)
	}

	return nil, fmt.Errorf("[CollectRows] missing scans function")
}

// CollectOneRow implements database.Database.
func (p *postgres[T]) CollectOneRow(ctx context.Context,
	builder query.Query, scan dbx.ScanOneFn[T]) (*T, error) {

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	stmt = sqlx.Rebind(sqlx.DOLLAR, stmt)
	zlog.Debugf("postgres: in tx=%t query: %s", p.tx, stmt)

	row := p.client.QueryRow(ctx, stmt, args...)

	if scan != nil {
		return scan(row)
	}

	return nil, fmt.Errorf("[CollectOneRow] missing scans function")
}

// Exec implements database.Database.
func (p *postgres[T]) Exec(ctx context.Context,
	builder query.Query) (dbx.ExecResult, error) {

	stmt, args, err := builder.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}

	stmt = sqlx.Rebind(sqlx.DOLLAR, stmt)

	zlog.Debugf("postgres: in tx=%t exec: %s", p.tx, stmt)

	result, err := p.client.Exec(ctx, stmt, args...)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	return result, nil
}

func (p *postgres[T]) Query(ctx context.Context,
	builder query.Query, dest ...any) error {

	stmt, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	stmt = sqlx.Rebind(sqlx.DOLLAR, stmt)

	zlog.Debugf("postgres: in tx=%t exec query: %s", p.tx, stmt)

	result, err := p.client.Query(ctx, stmt, args...)
	if err != nil {
		return err
	}

	for result.Next() {
		if err := result.Scan(dest...); err != nil {
			return err
		}
	}

	return nil
}
