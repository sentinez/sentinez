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
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/query"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

type M map[string]any

type Client interface {
	Exec(ctx context.Context, sql string,
		args ...any) (pgconn.CommandTag, error)

	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Paging selects the rows of page, the latest created first by the
// createdAt column.
func Paging(builder squirrel.SelectBuilder,
	page *typepb.Pages, createdAt string) squirrel.SelectBuilder {

	if page.GetSize() == 0 || page.GetIndex() == 0 {
		return builder
	}

	return LimitOffset(builder, page).OrderBy(createdAt + " DESC")
}

// LimitOffset selects the rows of page, all of them when page is nil or
// has no size or index; unlike Paging it leaves the order to the caller.
func LimitOffset(builder squirrel.SelectBuilder,
	page *typepb.Pages) squirrel.SelectBuilder {

	size, index := int(page.GetSize()), int(page.GetIndex())
	if size == 0 || index == 0 {
		return builder
	}

	return builder.
		Limit(uint64(size)).
		Offset(uint64(query.GetOffset(index, size)))
}

func SelectBuilder[T any](db dbx.Database[T],
	page *typepb.Pages, columns ...string) squirrel.SelectBuilder {

	builder := squirrel.Select(columns...).From(db.Table())
	if page == nil {
		return builder
	}

	return Paging(builder, page, db.CreatedAt())
}

// _idColumn is the primary key every table has (see dbx/stmt).
const _idColumn = "id"

func InsertBuilder[T any](db dbx.Database[T], mapp M) squirrel.InsertBuilder {
	return insertBuilder(db, mapp, slices.Sorted(maps.Keys(mapp))).
		Suffix("RETURNING " + _idColumn)
}

// InsertRowsBuilder inserts rows into the table of db in one statement,
// each the values valuesOf returns for columns, in their order. Their
// strings are made text postgres accepts (see textValue).
func InsertRowsBuilder[T, R any](db dbx.Database[T], columns []string,
	rows []R, valuesOf func(R) ([]any, error)) (squirrel.InsertBuilder, error) {

	builder := squirrel.Insert(db.Table()).
		Columns(columns...).
		PlaceholderFormat(squirrel.Dollar)
	for _, row := range rows {
		values, err := valuesOf(row)
		if err != nil {
			return squirrel.InsertBuilder{}, err
		}
		for i, v := range values {
			values[i] = textValue(v)
		}
		builder = builder.Values(values...)
	}

	return builder, nil
}

// UpsertBuilder inserts the row of mapp, or updates its columns when a row
// with the same id exists. mapp must hold the id.
func UpsertBuilder[T any](db dbx.Database[T], mapp M) squirrel.InsertBuilder {
	columns := slices.Sorted(maps.Keys(mapp))
	sets := []string{db.UpdatedAt() + " = now() AT TIME ZONE 'UTC'"}
	for _, column := range columns {
		if column != _idColumn {
			sets = append(sets, column+" = EXCLUDED."+column)
		}
	}

	return insertBuilder(db, mapp, columns).Suffix(
		"ON CONFLICT (" + _idColumn + ") DO UPDATE SET " +
			strings.Join(sets, ", ") + " RETURNING " + _idColumn)
}

// insertBuilder inserts the row of mapp; columns, its keys sorted, keep
// the statement the same for the same keys. Its strings are made text
// postgres accepts (see textValue).
func insertBuilder[T any](db dbx.Database[T], mapp M,
	columns []string) squirrel.InsertBuilder {

	values := make([]any, 0, len(columns))
	for _, column := range columns {
		values = append(values, textValue(mapp[column]))
	}

	return squirrel.Insert(db.Table()).
		Columns(columns...).
		Values(values...).
		PlaceholderFormat(squirrel.Dollar)
}

func UpdateBuilder[T any](
	db dbx.Database[T], id string) squirrel.UpdateBuilder {

	return squirrel.Update(db.Table()).
		Set(db.UpdatedAt(), time.Now().UTC()).
		Where(squirrel.Eq{"id": id})
}
