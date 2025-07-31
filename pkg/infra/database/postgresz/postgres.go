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

package postgresz

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func WithIndex(index ...string) database.Option {
	return func(ref *database.Table) {
		ref.Index = append(ref.Index, index...)
	}
}

func Field(field string) string {
	return fmt.Sprintf("%s->>'%s'", database.Data, field)
}

func Primary(field string) string {
	return field
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

		if err := protojson.Unmarshal(data, obj); err != nil {
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
	var empty T
	var data string

	if r == nil {
		zlog.Debug("Row is nil")
		return empty, errors.New("row is nil")
	}

	if err := r.Scan(&data); err != nil {
		zlog.Debugf("scan: error= %v", err)
		return empty, err
	}

	obj := reflect.New(reflect.TypeOf((*T)(nil)).Elem().Elem()).
		Interface().(proto.Message)

	if err := protojson.Unmarshal([]byte(data), obj); err != nil {
		return empty, err
	}

	return obj.(T), nil
}
