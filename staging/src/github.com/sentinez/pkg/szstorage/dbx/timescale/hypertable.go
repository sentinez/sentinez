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
	"fmt"
	"strings"
	"time"

	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/szstorage/dbx/stmt"
)

const (
	_createExtensionStmt = `CREATE EXTENSION IF NOT EXISTS timescaledb`

	// a primary key of a plain table, id the one of a postgres table,
	// prevents creating the hypertable: (id, time) replaces it.
	_dropPrimaryKeyStmt = `ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s_pkey`

	// migrate_data moves the rows of a plain table into chunks; no
	// default index on the time column, the Schema indexes it.
	_createHypertableStmt = `
		SELECT create_hypertable('%s', by_range('%s', %s),
			if_not_exists => TRUE, migrate_data => TRUE,
			create_default_indexes => FALSE)
	`

	_setChunkIntervalStmt = `SELECT set_chunk_time_interval('%s', %s)`

	// _nowFuncStmt the current time of a millisTime column, function
	// <table>_now; the policies of an integer hypertable need it.
	_nowFuncStmt = `
		CREATE OR REPLACE FUNCTION %s_now() RETURNS BIGINT
		LANGUAGE SQL STABLE
		AS $$ SELECT (EXTRACT(EPOCH FROM now()) * 1000)::BIGINT $$
	`

	_setNowFuncStmt = `
		SELECT set_integer_now_func('%s', '%s_now',
			replace_if_exists => TRUE)
	`

	// key name: <table>_id_<time column>_key
	_createKeyStmt = `
		CREATE UNIQUE INDEX IF NOT EXISTS %s_id_%s_key ON %s (id, %s)
	`

	// a policy is removed then added again, so a change applies.
	_removeRetentionStmt = `
		SELECT remove_retention_policy('%s', if_exists => TRUE)
	`
	_addRetentionStmt = `SELECT add_retention_policy('%s', %s)`
)

// syncHypertable runs the statements of hypertableStmts in one round
// trip, as a single implicit transaction.
func syncHypertable(ctx context.Context, client postgres.Client,
	tableName string, ht Hypertable) error {

	stmts := hypertableStmts(tableName, ht)

	// no arguments: pgx sends it with the simple protocol, which allows
	// several statements.
	if _, err := client.Exec(ctx, strings.Join(stmts, ";")); err != nil {
		return fmt.Errorf("sync hypertable %s: %w", tableName, err)
	}

	return nil
}

// hypertableStmts the statements creating, or migrating, the hypertable
// tableName described by ht, normalized: its table and columns, then the
// hypertable, then its key and indexes, then its retention policy.
func hypertableStmts(tableName string, ht Hypertable) []string {
	stmts := make([]string, 0, 10+len(ht.Schema.Columns)+
		len(ht.Schema.Indexes))
	stmts = append(stmts,
		_createExtensionStmt,
		stmt.CreateKeylessTable(tableName),
	)

	for _, column := range ht.Schema.Columns {
		stmts = append(stmts, stmt.AddColumn(tableName, column))
	}

	stmts = append(stmts,
		fmt.Sprintf(_dropPrimaryKeyStmt, tableName, tableName))
	stmts = append(stmts, chunkStmts(tableName, ht)...)
	stmts = append(stmts, fmt.Sprintf(_createKeyStmt,
		tableName, ht.TimeColumn, tableName, ht.TimeColumn))

	for _, index := range ht.Schema.Indexes {
		stmts = append(stmts, stmt.CreateIndex(tableName, index))
	}

	return append(stmts,
		retentionStmts(tableName, ht.Retention, ht.kind)...)
}

// chunkStmts turns tableName into a hypertable chunked by the time
// column of ht, with its now function when it is a millisTime one.
func chunkStmts(tableName string, ht Hypertable) []string {
	chunk := ht.kind.interval(ht.ChunkInterval)
	stmts := []string{
		fmt.Sprintf(_createHypertableStmt,
			tableName, ht.TimeColumn, chunk),
		fmt.Sprintf(_setChunkIntervalStmt, tableName, chunk),
	}

	if ht.kind != millisTime {
		return stmts
	}

	return append(stmts,
		fmt.Sprintf(_nowFuncStmt, tableName),
		fmt.Sprintf(_setNowFuncStmt, tableName, tableName),
	)
}

// retentionStmts sets the retention policy of the hypertable, or the
// continuous aggregate, name: none when retention is zero.
func retentionStmts(name string, retention time.Duration,
	kind timeKind) []string {

	stmts := []string{fmt.Sprintf(_removeRetentionStmt, name)}
	if retention <= 0 {
		return stmts
	}

	return append(stmts, fmt.Sprintf(_addRetentionStmt,
		name, kind.interval(retention)))
}
