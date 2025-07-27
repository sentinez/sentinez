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

// Package database provides the database interface.
package database

import (
	"github.com/Masterminds/squirrel"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func GetOffset(pageIndex, pageSize int) int {
	if pageIndex < 1 {
		pageIndex = 1
	}
	return (pageIndex - 1) * pageSize
}

func selectFrom(tableName string) squirrel.SelectBuilder {
	return squirrel.Select("data").From(tableName)
}

func SelectFrom(tableName string, page *common.Pages) squirrel.SelectBuilder {

	zlog.Debugf("selecting from table: %s", tableName)

	builder := selectFrom(tableName)

	if page == nil {
		return builder
	}

	return paging(builder, page)
}

func paging(builder squirrel.SelectBuilder,
	page *common.Pages) squirrel.SelectBuilder {

	if page == nil {
		return builder
	}

	if page.GetSize() == 0 || page.GetIndex() == 0 {
		return builder
	}

	offset := GetOffset(int(page.GetIndex()), int(page.GetSize()))
	return builder.
		Limit(uint64(page.GetSize())).
		Offset(uint64(offset)).
		OrderBy("created_at DESC")
}
