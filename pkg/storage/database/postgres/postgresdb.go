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
	"reflect"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	configspb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/configs/v1"
	"github.com/sentinez/sentinez/pkg/storage/database"
	"github.com/sentinez/sentinez/pkg/storage/database/query"
	storageutils "github.com/sentinez/sentinez/pkg/storage/utils"
	"github.com/sentinez/sentinez/pkg/storage/utils/table"
	"github.com/sentinez/sentinez/pkg/x/errorx"
	"github.com/sentinez/sentinez/pkg/x/jsonx"
	"github.com/sentinez/sentinez/pkg/zlog"
	"google.golang.org/protobuf/proto"
)

var _ database.Database[*common.Empty] = (*postgres[*common.Empty])(nil)

var (
	pool *pgxpool.Pool
	lock sync.Mutex
)

func getConnPool(conf *configspb.EnvConfig) (*pgxpool.Pool, error) {
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
func New[T proto.Message](conf *configspb.AppConfig, tableName string,
	opts ...database.Option) (database.Database[T], error) {

	tableName = table.NewTable(conf, tableName)

	conn, err := getConnPool(conf.GetEnvConf())
	if err != nil {
		return nil, err
	}

	if !table.IsValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	tb := database.Table{}
	for _, opt := range opts {
		opt(&tb)
	}

	err = proto2JSONB(context.Background(), conn, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to sync proto to pg err: %w", err)
	}

	return &postgres[T]{client: conn, tableName: tableName}, nil
}

type postgres[T proto.Message] struct {
	client    Client
	tableName string
}

func (p *postgres[T]) Table() string {
	return p.tableName
}

func (p *postgres[T]) Set(ctx context.Context, id string, entity T) error {
	data, err := jsonx.Marshal(entity)
	if err != nil {
		return nil
	}

	builder := sq.Insert(p.tableName).
		Columns(database.SchemalessFieldID, database.SchemalessFieldData).
		Values(id, string(data)).
		Suffix(fmt.Sprintf(`
        ON CONFLICT (%s)
        DO UPDATE SET
            %s = EXCLUDED.%s,
            updated_at = timezone('UTC', now())
    	`,
			database.SchemalessFieldID,
			database.SchemalessFieldData,
			database.SchemalessFieldData,
		))

	_, err = p.Exec(ctx, builder)
	return err
}

func (p *postgres[T]) Get(ctx context.Context, id string) (T, error) {
	var (
		empty T
		data  string
	)

	result := reflect.New(
		reflect.TypeOf((*T)(nil)).Elem().Elem()).Interface().(proto.Message)

	builder := sq.Select(database.SchemalessFieldData).
		From(p.tableName).
		Where(sq.Eq{
			database.SchemalessFieldID: id,
		})

	if err := p.Query(ctx, builder, &data); err != nil {
		return empty, err
	}

	if data == "" {
		return empty, errorx.ErrNotFound
	}

	if err := jsonx.Unmarshal([]byte(data), result); err != nil {
		return empty, err
	}

	return result.(T), nil
}

func (p *postgres[T]) Delete(ctx context.Context, id string) error {
	builder := sq.Delete(p.tableName).Where(sq.Eq{
		database.SchemalessFieldID: id,
	})

	_, err := p.Exec(ctx, builder)
	return err
}

// CollectRows implements database.Database.
func (p *postgres[T]) CollectRows(ctx context.Context,
	builder query.Query,
	fn func(database.Rows) ([]T, error)) ([]T, error) {

	stmt, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	stmt = sqlx.Rebind(sqlx.DOLLAR, stmt)
	zlog.Debugf("[QUERY] %s", stmt)

	rows, err := p.client.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if fn != nil {
		return fn(rows)
	}

	return nil, errorx.F("[CollectRows] missing scans function")
}

// CollectOneRow implements database.Database.
func (p *postgres[T]) CollectOneRow(ctx context.Context,
	builder query.Query, scan func(database.Row) (T, error)) (T, error) {

	var empty T

	stmt, args, err := builder.ToSql()
	if err != nil {
		return empty, err
	}

	stmt = sqlx.Rebind(sqlx.DOLLAR, stmt)

	row := p.client.QueryRow(ctx, stmt, args...)

	if scan != nil {
		return scan(row)
	}

	return empty, errorx.F("[CollectOneRow] missing scans function")
}

// Exec implements database.Database.
func (p *postgres[T]) Exec(ctx context.Context,
	builder query.Query) (database.ExecResult, error) {

	stmt, args, err := builder.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}

	stmt = sqlx.Rebind(sqlx.DOLLAR, stmt)

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
