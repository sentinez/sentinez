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
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/query"
	"github.com/sentinez/pkg/szstorage/storageutils"
	"github.com/sentinez/pkg/szstorage/storageutils/table"
	"github.com/sentinez/pkg/sztable"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

var _ dbx.Database[typepb.Empty] = (*postgres[typepb.Empty])(nil)

var errMissingScan = errors.New("missing scan function")

var (
	_poolMu sync.Mutex
	_pool   atomic.Pointer[pgxpool.Pool]
)

// SyncFn creates or migrates the relation tableName through client.
type SyncFn func(ctx context.Context, client Client, tableName string) error

// New returns the Database of the table described by schema, synced on
// the way: the table, its missing columns and indexes are created.
func New[T any](ctx context.Context, conf *settingpb.Config,
	schema dbx.Schema) (dbx.Database[T], error) {

	return Open[T](ctx, conf, schema.Table,
		func(ctx context.Context, client Client, tableName string) error {
			return syncSchema(ctx, client, tableName, schema)
		})
}

// Open returns the Database of the relation name, on the shared pool,
// once sync created or migrated it; New with a sync of its own, for
// the relations a dbx.Schema does not describe (hypertables, views).
func Open[T any](ctx context.Context, conf *settingpb.Config,
	name sztable.Table, sync SyncFn) (dbx.Database[T], error) {

	tableName, err := TableName(conf, name)
	if err != nil {
		return nil, err
	}

	pool, err := getConnPool(conf)
	if err != nil {
		return nil, err
	}

	if err := sync(ctx, pool, tableName); err != nil {
		return nil, err
	}

	return &postgres[T]{client: pool, tableName: tableName}, nil
}

// TableName the name of the relation name in the database of conf: the
// one Open gives its Database.
func TableName(conf *settingpb.Config, name sztable.Table) (string, error) {
	tableName := table.NewTable(conf, name.String())
	if !table.IsValidTableName(tableName) {
		return "", fmt.Errorf("invalid table name: %s", tableName)
	}

	return tableName, nil
}

type postgres[T any] struct {
	client    Client
	tableName string
	tx        bool
}

// Table implements dbx.Database.
func (p *postgres[T]) Table() string {
	return p.tableName
}

// Total implements dbx.Database.
func (p *postgres[T]) Total(ctx context.Context) (int64, error) {
	var total int64

	q := sq.Select("COUNT(*) AS count").From(p.tableName)
	if err := p.Query(ctx, q, &total); err != nil {
		return -1, err
	}

	return total, nil
}

// Insert implements dbx.Database, builder returns the inserted id.
func (p *postgres[T]) Insert(
	ctx context.Context, builder query.Query) (string, error) {

	stmt, args, err := p.build(builder)
	if err != nil {
		return "", err
	}

	var id string
	if err := p.client.QueryRow(ctx, stmt, args...).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

// Select implements dbx.Database.
func (p *postgres[T]) Select(ctx context.Context,
	builder query.Query, scan dbx.ScanOneFn[T]) (*T, error) {
	return p.CollectOneRow(ctx, builder, scan)
}

// Delete implements dbx.Database.
func (p *postgres[T]) Delete(ctx context.Context, id string) error {
	_, err := p.Exec(ctx, sq.Delete(p.tableName).Where(sq.Eq{"id": id}))
	return err
}

// CollectRows implements dbx.Executor.
func (p *postgres[T]) CollectRows(ctx context.Context,
	builder query.Query, scan dbx.ScanFn[T]) ([]*T, error) {

	if scan == nil {
		return nil, errMissingScan
	}

	stmt, args, err := p.build(builder)
	if err != nil {
		return nil, err
	}

	rows, err := p.client.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scan(rows)
}

// CollectOneRow implements dbx.Executor.
func (p *postgres[T]) CollectOneRow(ctx context.Context,
	builder query.Query, scan dbx.ScanOneFn[T]) (*T, error) {

	if scan == nil {
		return nil, errMissingScan
	}

	stmt, args, err := p.build(builder)
	if err != nil {
		return nil, err
	}

	return scan(p.client.QueryRow(ctx, stmt, args...))
}

// Exec implements dbx.Executor.
func (p *postgres[T]) Exec(ctx context.Context,
	builder query.Query) (dbx.ExecResult, error) {

	stmt, args, err := p.build(builder)
	if err != nil {
		return nil, err
	}

	return p.client.Exec(ctx, stmt, args...)
}

// Query implements dbx.Executor, every row is scanned into dest.
func (p *postgres[T]) Query(ctx context.Context,
	builder query.Query, dest ...any) error {

	stmt, args, err := p.build(builder)
	if err != nil {
		return err
	}

	rows, err := p.client.Query(ctx, stmt, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			return err
		}
	}

	return rows.Err()
}

// build renders builder with $n placeholders.
func (p *postgres[T]) build(builder query.Query) (string, []any, error) {
	stmt, args, err := builder.ToSql()
	if err != nil {
		return "", nil, err
	}

	stmt = sqlx.Rebind(sqlx.DOLLAR, stmt)
	szlog.Debugf("postgres: in tx=%t exec: %s", p.tx, stmt)

	return stmt, args, nil
}

// getConnPool returns the pool shared by every table, opened once; the
// lock is only taken until it is.
func getConnPool(conf *settingpb.Config) (*pgxpool.Pool, error) {
	if pool := _pool.Load(); pool != nil {
		return pool, nil
	}

	_poolMu.Lock()
	defer _poolMu.Unlock()

	if pool := _pool.Load(); pool != nil {
		return pool, nil
	}

	pool, err := storageutils.NewPgxPool(conf)
	if err != nil {
		return nil, err
	}

	_pool.Store(pool)

	return pool, nil
}
