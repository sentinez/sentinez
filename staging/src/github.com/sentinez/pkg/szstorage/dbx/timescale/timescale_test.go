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

package timescale

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ postgres.Client = (*fakeClient)(nil)

// fakeClient records the statements of each Exec, failing with err; its
// rows hold exists.
type fakeClient struct {
	execs  [][]string
	err    error
	exists bool
}

func (c *fakeClient) Exec(_ context.Context, sql string,
	_ ...any) (pgconn.CommandTag, error) {

	var stmts []string
	for _, stmt := range strings.Split(sql, ";") {
		stmts = append(stmts, strings.Join(strings.Fields(stmt), " "))
	}
	c.execs = append(c.execs, stmts)

	return pgconn.CommandTag{}, c.err
}

func (c *fakeClient) Query(context.Context, string,
	...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected query")
}

func (c *fakeClient) QueryRow(context.Context, string, ...any) pgx.Row {
	return fakeRow{exists: c.exists}
}

// fakeRow scans exists into its single destination, a *bool.
type fakeRow struct {
	exists bool
}

func (r fakeRow) Scan(dest ...any) error {
	exists, ok := dest[0].(*bool)
	if !ok {
		return errors.New("unexpected destination")
	}
	*exists = r.exists

	return nil
}

func eventsHypertable() Hypertable {
	return Hypertable{
		Schema: dbx.Schema{
			Columns: []dbx.Column{
				{Name: "resource_id", Type: dbx.Text},
				{Name: "ts", Type: dbx.Int8},
			},
			Indexes: []dbx.Index{{Column: "resource_id", Method: dbx.BTree}},
		},
		TimeColumn:    "ts",
		ChunkInterval: time.Hour,
		Retention:     24 * time.Hour,
	}
}

func normalized(t *testing.T, ht Hypertable) Hypertable {
	ht, err := ht.normalize()
	require.NoError(t, err)

	return ht
}

func TestOnConflictDoNothing(t *testing.T) {
	assert.Equal(t, "ON CONFLICT (id, ts) DO NOTHING",
		eventsHypertable().OnConflictDoNothing())
	assert.Equal(t, "ON CONFLICT (id, created_at) DO NOTHING",
		Hypertable{}.OnConflictDoNothing())
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name      string
		give      Hypertable
		wantTime  string
		wantKind  timeKind
		wantChunk time.Duration
	}{
		{
			name:      "created_at by default",
			wantTime:  "created_at",
			wantKind:  timestampTime,
			wantChunk: _defaultChunkInterval,
		},
		{
			name:      "bigint milliseconds",
			give:      eventsHypertable(),
			wantTime:  "ts",
			wantKind:  millisTime,
			wantChunk: time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ht, err := tt.give.normalize()

			require.NoError(t, err)
			assert.Equal(t, tt.wantTime, ht.TimeColumn)
			assert.Equal(t, tt.wantKind, ht.kind)
			assert.Equal(t, tt.wantChunk, ht.ChunkInterval)
		})
	}
}

func TestNormalizeError(t *testing.T) {
	schema := eventsHypertable().Schema

	tests := []struct {
		name    string
		give    Hypertable
		wantErr string
	}{
		{
			name:    "missing column",
			give:    Hypertable{TimeColumn: "ts"},
			wantErr: "time column ts: not in schema",
		},
		{
			name:    "wrong type",
			give:    Hypertable{TimeColumn: "resource_id", Schema: schema},
			wantErr: "time column resource_id: type TEXT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.give.normalize()

			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestTimeKindLiterals(t *testing.T) {
	at := time.Date(2026, 11, 6, 1, 2, 3, 0, time.UTC)

	tests := []struct {
		give         timeKind
		wantInterval string
		wantInstant  string
	}{
		{
			give:         timestampTime,
			wantInterval: "INTERVAL '3600000 milliseconds'",
			wantInstant:  "'2026-11-06T01:02:03Z'::TIMESTAMPTZ",
		},
		{
			give:         millisTime,
			wantInterval: "3600000::BIGINT",
			wantInstant:  "1793926923000::BIGINT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.wantInterval, func(t *testing.T) {
			assert.Equal(t, tt.wantInterval, tt.give.interval(time.Hour))
			assert.Equal(t, tt.wantInstant, tt.give.instant(at))
			assert.Equal(t, "NULL", tt.give.instant(time.Time{}))
		})
	}
}

func TestSyncHypertable(t *testing.T) {
	client := &fakeClient{}

	err := syncHypertable(context.Background(), client, "t",
		normalized(t, eventsHypertable()))

	require.NoError(t, err)
	require.Len(t, client.execs, 1)

	stmts := client.execs[0]
	require.Len(t, stmts, 13)
	assert.Equal(t, "CREATE EXTENSION IF NOT EXISTS timescaledb", stmts[0])
	assert.Contains(t, stmts[1],
		"CREATE TABLE IF NOT EXISTS t ( id TEXT NOT NULL,")
	assert.Equal(t, []string{
		"ALTER TABLE t ADD COLUMN IF NOT EXISTS resource_id TEXT",
		"ALTER TABLE t ADD COLUMN IF NOT EXISTS ts BIGINT",
		"ALTER TABLE t DROP CONSTRAINT IF EXISTS t_pkey",
		"SELECT create_hypertable('t', by_range('ts', 3600000::BIGINT), " +
			"if_not_exists => TRUE, migrate_data => TRUE, " +
			"create_default_indexes => FALSE)",
		"SELECT set_chunk_time_interval('t', 3600000::BIGINT)",
	}, stmts[2:7])
	assert.Contains(t, stmts[7], "CREATE OR REPLACE FUNCTION t_now()")
	assert.Equal(t, []string{
		"SELECT set_integer_now_func('t', 't_now', " +
			"replace_if_exists => TRUE)",
		"CREATE UNIQUE INDEX IF NOT EXISTS t_id_ts_key ON t (id, ts)",
		"CREATE INDEX IF NOT EXISTS t_resource_id_btree_idx " +
			"ON t USING BTREE (resource_id)",
		"SELECT remove_retention_policy('t', if_exists => TRUE)",
		"SELECT add_retention_policy('t', 86400000::BIGINT)",
	}, stmts[8:])
}

func TestSyncHypertableTimestamp(t *testing.T) {
	client := &fakeClient{}
	err := syncHypertable(context.Background(), client, "t",
		normalized(t, Hypertable{}))

	require.NoError(t, err)
	stmts := client.execs[0]
	assert.NotContains(t, strings.Join(stmts, ";"), "t_now")
	assert.Equal(t, []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS t_id_created_at_key " +
			"ON t (id, created_at)",
		"SELECT remove_retention_policy('t', if_exists => TRUE)",
	}, stmts[len(stmts)-2:])
}

func TestSyncHypertableError(t *testing.T) {
	client := &fakeClient{err: errors.New("boom")}

	err := syncHypertable(context.Background(), client, "t",
		normalized(t, eventsHypertable()))

	require.ErrorContains(t, err, "sync hypertable t: boom")
}

func eventsView(t *testing.T) ViewSchema {
	return ViewSchema{
		Source:     normalized(t, eventsHypertable()),
		Bucket:     time.Hour,
		GroupBy:    []string{"resource_id"},
		Aggregates: []Aggregate{{Expr: "count(*)", As: "total"}},
		Refresh: RefreshPolicy{
			EndOffset: time.Hour,
			Schedule:  30 * time.Minute,
		},
		Retention: 24 * time.Hour,
	}
}

func TestSyncView(t *testing.T) {
	client := &fakeClient{}

	err := syncView(context.Background(), client, "v", "t",
		eventsView(t))

	require.NoError(t, err)
	require.Len(t, client.execs, 3)
	assert.Equal(t, [][]string{
		{
			"CREATE MATERIALIZED VIEW v " +
				"WITH (timescaledb.continuous) AS " +
				"SELECT time_bucket(3600000::BIGINT, ts) AS bucket, " +
				"resource_id, count(*) AS total " +
				"FROM t GROUP BY bucket, resource_id WITH NO DATA",
		},
		{
			"ALTER MATERIALIZED VIEW v " +
				"SET (timescaledb.materialized_only = true)",
			"SELECT remove_continuous_aggregate_policy('v', " +
				"if_exists => TRUE)",
			"SELECT add_continuous_aggregate_policy('v', " +
				"start_offset => NULL, end_offset => 3600000::BIGINT, " +
				"schedule_interval => INTERVAL '1800000 milliseconds')",
			"SELECT remove_retention_policy('v', if_exists => TRUE)",
			"SELECT add_retention_policy('v', 86400000::BIGINT)",
		},
	}, client.execs[:2])
	assert.Contains(t, client.execs[2][0],
		"CALL refresh_continuous_aggregate('v', NULL, ")
}

func TestSyncViewExisting(t *testing.T) {
	client := &fakeClient{exists: true}

	err := syncView(context.Background(), client, "v", "t",
		eventsView(t))

	require.NoError(t, err)
	require.Len(t, client.execs, 1)
	assert.Contains(t, client.execs[0][0], "ALTER MATERIALIZED VIEW v")
}

func TestSyncViewRealTimeNoPolicies(t *testing.T) {
	client := &fakeClient{}
	view := eventsView(t)
	view.RealTime = true
	view.Refresh = RefreshPolicy{}
	view.Retention = 0

	err := syncView(context.Background(), client, "v", "t", view)

	require.NoError(t, err)
	require.Len(t, client.execs, 3)
	assert.Equal(t, []string{
		"ALTER MATERIALIZED VIEW v " +
			"SET (timescaledb.materialized_only = false)",
		"SELECT remove_continuous_aggregate_policy('v', " +
			"if_exists => TRUE)",
		"SELECT remove_retention_policy('v', if_exists => TRUE)",
	}, client.execs[1])
}

func TestSyncViewError(t *testing.T) {
	client := &fakeClient{err: errors.New("boom")}

	err := syncView(context.Background(), client, "v", "t",
		eventsView(t))

	require.ErrorContains(t, err, "create view v: boom")
}

func TestNewViewBucket(t *testing.T) {
	view := eventsView(t)
	view.Bucket = 0

	_, err := NewView[struct{}](context.Background(), nil, view)

	require.ErrorIs(t, err, errBucket)
}
