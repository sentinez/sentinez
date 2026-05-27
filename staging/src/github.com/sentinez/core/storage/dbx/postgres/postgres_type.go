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

import "github.com/sentinez/core/storage/dbx"

const (
	Timestamp dbx.ColumnType = "TIMESTAMPTZ"
	String    dbx.ColumnType = "TEXT"
	StringArr dbx.ColumnType = "TEXT[]"
	IntArr    dbx.ColumnType = "INTEGER[]"
	ByteA     dbx.ColumnType = "BYTEA"
	Int4      dbx.ColumnType = "INTEGER"
	Int8      dbx.ColumnType = "BIGINT"
	Float4    dbx.ColumnType = "REAL"
	Float8    dbx.ColumnType = "DOUBLE PRECISION"
	Bool      dbx.ColumnType = "BOOLEAN"
	JSONB     dbx.ColumnType = "JSONB"
)
