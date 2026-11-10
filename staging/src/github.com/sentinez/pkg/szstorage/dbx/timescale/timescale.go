// Copyright 2026 Duc-Hung Ho.
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

// Package timescale stores dbx tables as TimescaleDB hypertables, and
// derives continuous aggregates from them: materialized views that
// TimescaleDB keeps up to date. It is built on the postgres package and
// shares its pool, so a hypertable joins the transactions of
// postgres.Tx (postgres.WithTx).
package timescale

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

const _defaultChunkInterval = 7 * 24 * time.Hour

// Hypertable describes a hypertable: a table TimescaleDB partitions into
// chunks by the time range of TimeColumn.
//
// Unlike a postgres table, id is not the primary key: a unique index of
// a hypertable must hold its time column, so (id, TimeColumn) is unique
// instead, the conflict target of an upsert (OnConflictDoNothing).
type Hypertable struct {
	Schema dbx.Schema
	// TimeColumn partitions the rows, created_at when empty; a column of
	// Schema typed dbx.Timestamp, or dbx.Int8 holding unix milliseconds.
	TimeColumn string
	// ChunkInterval the time range of a chunk, 7 days when zero; a change
	// applies to the chunks created from then on.
	ChunkInterval time.Duration
	// Retention drops the chunks whose rows are all older than it; the
	// rows are kept forever when zero.
	Retention time.Duration

	// kind the kind of TimeColumn, set by Normalized.
	kind timeKind
}

// New returns the Database of the hypertable described by ht, synced on
// the way: the timescaledb extension, the table, its missing columns,
// the hypertable, its unique (id, time) key, its indexes and its
// retention policy are created.
func New[T any](ctx context.Context, conf *settingpb.Config,
	ht Hypertable) (dbx.Database[T], error) {

	ht, err := ht.Normalized()
	if err != nil {
		return nil, err
	}

	return postgres.Open[T](ctx, conf, ht.Schema,
		func(ctx context.Context, client postgres.Client,
			tableName string) error {
			return syncHypertable(ctx, client, tableName, ht)
		})
}

// OnConflictDoNothing the suffix of an INSERT into the hypertable that
// skips the rows whose id and time are stored already.
func (ht Hypertable) OnConflictDoNothing() string {
	return "ON CONFLICT (id, " + ht.timeColumn() + ") DO NOTHING"
}

// Normalized returns ht with its defaults and the kind of its time
// column set, as New and NewView take it; an error when the time column
// is not in the schema or neither a TIMESTAMPTZ nor a BIGINT.
func (ht Hypertable) Normalized() (Hypertable, error) {
	ht.TimeColumn = ht.timeColumn()
	if ht.ChunkInterval <= 0 {
		ht.ChunkInterval = _defaultChunkInterval
	}

	kind, err := ht.timeKind()
	ht.kind = kind

	return ht, err
}

// MillisTime whether the time column of ht, Normalized, holds unix
// milliseconds.
func (ht Hypertable) MillisTime() bool {
	return ht.kind == millisTime
}

// TimeBucket the SQL of the start of the bucket of d holding the time
// column of ht, Normalized: time_bucket(<interval of d>, <column>).
func (ht Hypertable) TimeBucket(d time.Duration) string {
	return "time_bucket(" + ht.kind.interval(d) + ", " + ht.TimeColumn + ")"
}

func (ht Hypertable) timeColumn() string {
	if ht.TimeColumn == "" {
		return ht.Schema.CreatedAtColumn()
	}

	return ht.TimeColumn
}

// timeKind the kind of the time column of ht, which must be set.
func (ht Hypertable) timeKind() (timeKind, error) {
	columnType := dbx.Timestamp // the creation time column
	if ht.TimeColumn != ht.Schema.CreatedAtColumn() {
		column, ok := ht.column(ht.TimeColumn)
		if !ok {
			return 0, fmt.Errorf("time column %s: not in schema %s",
				ht.TimeColumn, ht.Schema.Table)
		}
		columnType = column.Type
	}

	switch columnType {
	case dbx.Timestamp:
		return timestampTime, nil
	case dbx.Int8:
		return millisTime, nil
	default:
		return 0, fmt.Errorf("time column %s: type %s, want %s or %s",
			ht.TimeColumn, columnType, dbx.Timestamp, dbx.Int8)
	}
}

func (ht Hypertable) column(name string) (dbx.Column, bool) {
	for _, column := range ht.Schema.Columns {
		if column.Name == name {
			return column, true
		}
	}

	return dbx.Column{}, false
}

// timeKind how a time column stores an instant, so how the SQL literals
// of its instants and intervals are written.
type timeKind int

const (
	// timestampTime a TIMESTAMPTZ column.
	timestampTime timeKind = iota + 1
	// millisTime a BIGINT column of unix milliseconds.
	millisTime
)

// interval the SQL literal of d, a range of the time column.
func (k timeKind) interval(d time.Duration) string {
	if k == millisTime {
		return bigint(d.Milliseconds())
	}

	return interval(d)
}

// instant the SQL literal of t, a value of the time column; NULL, an
// unbounded end of a range, when t is zero.
func (k timeKind) instant(t time.Time) string {
	switch {
	case t.IsZero():
		return "NULL"
	case k == millisTime:
		return bigint(t.UnixMilli())
	default:
		return "'" + t.UTC().Format(time.RFC3339Nano) + "'::TIMESTAMPTZ"
	}
}

// interval the INTERVAL literal of d, to the millisecond.
func interval(d time.Duration) string {
	return "INTERVAL '" + strconv.FormatInt(d.Milliseconds(), 10) +
		" milliseconds'"
}

func bigint(n int64) string {
	return strconv.FormatInt(n, 10) + "::BIGINT"
}
