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

package szpostgres

import "github.com/sentinez/core/szstorage/szdbx"

const (
	Timestamp szdbx.ColumnType = "TIMESTAMPTZ"
	String    szdbx.ColumnType = "TEXT"
	StringArr szdbx.ColumnType = "TEXT[]"
	IntArr    szdbx.ColumnType = "INTEGER[]"
	ByteA     szdbx.ColumnType = "BYTEA"
	Int4      szdbx.ColumnType = "INTEGER"
	Int8      szdbx.ColumnType = "BIGINT"
	Float4    szdbx.ColumnType = "REAL"
	Float8    szdbx.ColumnType = "DOUBLE PRECISION"
	Bool      szdbx.ColumnType = "BOOLEAN"
	JSONB     szdbx.ColumnType = "JSONB"
)

const (
	BTree szdbx.IndexMethod = "BTREE"
	// GIN indexes the keys and values of a JSONB column, it serves the
	// containment operator (@>).
	GIN szdbx.IndexMethod = "GIN"
)
