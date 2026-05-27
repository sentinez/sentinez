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
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sentinez/core/storage/dbx"
	"github.com/sentinez/core/storage/dbx/query"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
)

type M map[string]any

type Client interface {
	Exec(ctx context.Context, sql string,
		args ...any) (pgconn.CommandTag, error)

	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func Paging(builder squirrel.SelectBuilder,
	page *typepb.Pages) squirrel.SelectBuilder {

	if page == nil {
		return builder
	}

	if page.GetSize() == 0 || page.GetIndex() == 0 {
		return builder
	}

	offset := query.GetOffset(int(page.GetIndex()), int(page.GetSize()))
	return builder.
		Limit(uint64(page.GetSize())).
		Offset(uint64(offset)).
		OrderBy(fmt.Sprintf("%s DESC", dbx.FieldCreatedAt))
}

func SelectBuilder[T any](db dbx.Database[T],
	page *typepb.Pages, columns ...string) squirrel.SelectBuilder {

	builder := squirrel.Select(columns...).From(db.Table())
	if page == nil {
		return builder
	}

	return Paging(builder, page)
}

func InsertBuilder[T any](db dbx.Database[T], mapp M) squirrel.InsertBuilder {
	var (
		columns []string
		values  []any
	)
	for column, value := range mapp {
		columns = append(columns, column)
		values = append(values, value)
	}

	return squirrel.Insert(db.Table()).
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar)
}

func UpdateBuilder[T any](
	db dbx.Database[T], id string) squirrel.UpdateBuilder {

	return squirrel.Update(db.Table()).
		Set(dbx.FieldUpdatedAt, time.Now().UTC()).
		Where(squirrel.Eq{"id": id})
}
