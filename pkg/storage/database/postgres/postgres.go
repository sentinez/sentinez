// Copyright 2025 Sentinez Labs.
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
	"reflect"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	"github.com/sentinez/sentinez/pkg/storage/database"
	"github.com/sentinez/sentinez/pkg/storage/database/query"
	"github.com/sentinez/sentinez/pkg/x/jsonx"
	"github.com/sentinez/sentinez/pkg/zlog"
	"google.golang.org/protobuf/proto"
)

type Client interface {
	Exec(ctx context.Context, sql string,
		args ...any) (pgconn.CommandTag, error)

	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func Field(field string) string {
	return fmt.Sprintf("%s->>'%s'", database.SchemalessFieldData, field)
}

func Primary(field string) string {
	return field
}

func Paging(builder squirrel.SelectBuilder,
	page *common.Pages) squirrel.SelectBuilder {

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
		OrderBy(fmt.Sprintf("%s DESC", database.SchemalessFieldCreatedAt))
}

func SelectBuilder[T proto.Message](
	db database.Database[T], page *common.Pages) squirrel.SelectBuilder {

	builder := squirrel.Select(database.SchemalessFieldData).From(db.Table())
	if page == nil {
		return builder
	}

	return Paging(builder, page)
}

func Scans[T proto.Message](r database.Rows) ([]T, error) {
	var list []T

	for r.Next() {
		var (
			data []byte
		)

		if err := r.Scan(&data); err != nil {
			return nil, err
		}

		obj := reflect.New(reflect.TypeOf((*T)(nil)).Elem().Elem()).
			Interface().(proto.Message)

		if err := jsonx.Unmarshal(data, obj); err != nil {
			return nil, err
		}

		list = append(list, obj.(T))
	}

	if err := r.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func Scan[T proto.Message](r database.Row) (T, error) {

	var (
		empty T
		data  []byte
	)

	if r == nil {
		zlog.Debug("Row is nil")
		return empty, errors.New("row is nil")
	}

	if err := r.Scan(&data); err != nil {
		//zlog.Debugf("scan: error= %v", err)
		return empty, err
	}

	obj := reflect.New(reflect.TypeOf((*T)(nil)).Elem().Elem()).
		Interface().(proto.Message)

	if err := jsonx.Unmarshal(data, obj); err != nil {
		return empty, err
	}

	return obj.(T), nil
}
