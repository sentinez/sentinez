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
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/szstorage/dbx/query"
	"github.com/sentinez/pkg/sztable"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// BucketColumn the column of a View holding the start of its buckets.
const BucketColumn = "bucket"

const (
	_viewExistsStmt = `SELECT to_regclass($1) IS NOT NULL`

	// WITH NO DATA: WITH DATA would materialize the bucket in progress,
	// hiding its next rows from RealTime until refreshed; backfillView
	// materializes the complete ones.
	_createViewStmt = `
		CREATE MATERIALIZED VIEW %s
		WITH (timescaledb.continuous) AS
		SELECT %s FROM %s GROUP BY %s
		WITH NO DATA
	`

	_setRealTimeStmt = `
		ALTER MATERIALIZED VIEW %s
		SET (timescaledb.materialized_only = %t)
	`

	// a policy is removed then added again, so a change applies.
	_removeRefreshStmt = `
		SELECT remove_continuous_aggregate_policy('%s',
			if_exists => TRUE)
	`
	_addRefreshStmt = `
		SELECT add_continuous_aggregate_policy('%s',
			start_offset => %s, end_offset => %s,
			schedule_interval => %s)
	`

	_refreshStmt = `CALL refresh_continuous_aggregate('%s', %s, %s)`
)

var errBucket = errors.New("view bucket: must be positive")

// ViewSchema describes a continuous aggregate: a materialized view of
// Source, the rows grouped by time buckets of Bucket then GroupBy, that
// TimescaleDB refreshes by its Refresh policy.
//
// The view is created once, its buckets of the rows of Source stored
// already materialized then: a change of its definition (Source, Bucket,
// GroupBy, Aggregates) needs a new Name, or the view dropped; RealTime,
// Refresh and Retention apply on every sync.
type ViewSchema struct {
	Name   sztable.Table
	Source Hypertable
	// Bucket the time range of a bucket, BucketColumn its start.
	Bucket time.Duration
	// GroupBy the columns of Source grouped by after the bucket.
	GroupBy []string
	// Aggregates the aggregated columns, after GroupBy.
	Aggregates []Aggregate
	// RealTime also aggregates, on query, the rows of Source not
	// materialized yet; otherwise the view only shows refreshed buckets.
	RealTime bool
	// Refresh the policy refreshing the view, none when its Schedule is
	// zero: refresh it with View.Refresh then.
	Refresh RefreshPolicy
	// Retention drops the buckets older than it, kept forever when zero;
	// keep it below the Retention of Source, or the refresh of a range
	// whose rows were dropped empties its buckets.
	Retention time.Duration
}

// Aggregate a column of a view: the SQL aggregate Expr over the rows of
// a bucket, ex: {Expr: "count(*)", As: "total"}.
type Aggregate struct {
	Expr string
	As   string
}

// RefreshPolicy refreshes, every Schedule, the buckets from now -
// StartOffset (all the older ones too when zero) to now - EndOffset.
type RefreshPolicy struct {
	StartOffset time.Duration
	EndOffset   time.Duration
	Schedule    time.Duration
}

// View a continuous aggregate, scanned into T.
type View[T any] struct {
	db   dbx.Database[T]
	kind timeKind
}

// NewView returns the View described by schema, synced on the way: the
// view is created over its source hypertable, which must exist (see
// New), then its real time mode, refresh and retention policies are set.
func NewView[T any](ctx context.Context, conf *settingpb.Config,
	schema ViewSchema) (*View[T], error) {

	if schema.Bucket <= 0 {
		return nil, errBucket
	}

	source, err := schema.Source.normalize()
	if err != nil {
		return nil, err
	}
	schema.Source = source

	sourceName, err := postgres.TableName(conf, source.Schema.Table)
	if err != nil {
		return nil, err
	}

	db, err := postgres.Open[T](ctx, conf, schema.Name,
		func(ctx context.Context, client postgres.Client,
			viewName string) error {
			return syncView(ctx, client, viewName, sourceName, schema)
		})
	if err != nil {
		return nil, err
	}

	return &View[T]{db: db, kind: source.kind}, nil
}

// Table the name of the view.
func (v *View[T]) Table() string {
	return v.db.Table()
}

// CollectRows the rows of builder, a query of the view, scanned by scan.
func (v *View[T]) CollectRows(ctx context.Context,
	builder query.Query, scan dbx.ScanFn[T]) ([]*T, error) {
	return v.db.CollectRows(ctx, builder, scan)
}

// Refresh materializes the buckets of the view held in [from, to), an
// unbounded end when zero. It runs outside a transaction.
func (v *View[T]) Refresh(ctx context.Context, from, to time.Time) error {
	_, err := v.db.Exec(ctx, sq.Expr(fmt.Sprintf(_refreshStmt,
		v.Table(), v.kind.instant(from), v.kind.instant(to))))
	if err != nil {
		return fmt.Errorf("refresh view %s: %w", v.Table(), err)
	}

	return nil
}

// syncView creates the view viewName over sourceName when missing, runs
// its policy statements in one round trip, then backfills it when it
// was created. The view is created, and refreshed, alone: TimescaleDB
// refuses to inside a transaction.
func syncView(ctx context.Context, client postgres.Client,
	viewName, sourceName string, schema ViewSchema) error {

	var exists bool
	err := client.QueryRow(ctx, _viewExistsStmt, viewName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("view %s exists: %w", viewName, err)
	}

	if !exists {
		create := createViewStmt(viewName, sourceName, schema)
		if _, err := client.Exec(ctx, create); err != nil {
			return fmt.Errorf("create view %s: %w", viewName, err)
		}
	}

	policy := strings.Join(viewPolicyStmts(viewName, schema), ";")
	if _, err := client.Exec(ctx, policy); err != nil {
		return fmt.Errorf("sync view %s: %w", viewName, err)
	}

	if exists {
		return nil
	}

	return backfillView(ctx, client, viewName, schema)
}

// backfillView materializes the complete buckets of the view up to now -
// Refresh.EndOffset: the refresh policy only covers its recent ones, the
// older rows of the source would never be.
func backfillView(ctx context.Context, client postgres.Client,
	viewName string, schema ViewSchema) error {

	to := time.Now().Add(-schema.Refresh.EndOffset)
	refresh := fmt.Sprintf(_refreshStmt, viewName, "NULL",
		schema.Source.kind.instant(to))
	if _, err := client.Exec(ctx, refresh); err != nil {
		return fmt.Errorf("backfill view %s: %w", viewName, err)
	}

	return nil
}

// createViewStmt selects the bucket, the GroupBy columns then the
// Aggregates of schema from sourceName.
func createViewStmt(viewName, sourceName string, schema ViewSchema) string {
	bucket := fmt.Sprintf("time_bucket(%s, %s) AS %s",
		schema.Source.kind.interval(schema.Bucket),
		schema.Source.TimeColumn, BucketColumn)

	columns := make([]string, 0, 1+len(schema.GroupBy)+
		len(schema.Aggregates))
	columns = append(columns, bucket)
	columns = append(columns, schema.GroupBy...)

	for _, aggregate := range schema.Aggregates {
		columns = append(columns, aggregate.Expr+" AS "+aggregate.As)
	}

	groupBy := append([]string{BucketColumn}, schema.GroupBy...)

	return fmt.Sprintf(_createViewStmt, viewName,
		strings.Join(columns, ", "), sourceName,
		strings.Join(groupBy, ", "))
}

// viewPolicyStmts sets the real time mode of the view, then its refresh
// and retention policies.
func viewPolicyStmts(viewName string, schema ViewSchema) []string {
	kind := schema.Source.kind
	stmts := []string{
		fmt.Sprintf(_setRealTimeStmt, viewName, !schema.RealTime),
	}
	stmts = append(stmts, refreshStmts(viewName, schema.Refresh, kind)...)

	return append(stmts, retentionStmts(viewName, schema.Retention, kind)...)
}

// refreshStmts sets the refresh policy of the view: none when its
// Schedule is zero.
func refreshStmts(viewName string, policy RefreshPolicy,
	kind timeKind) []string {

	stmts := []string{fmt.Sprintf(_removeRefreshStmt, viewName)}
	if policy.Schedule <= 0 {
		return stmts
	}

	start := "NULL"
	if policy.StartOffset > 0 {
		start = kind.interval(policy.StartOffset)
	}

	return append(stmts, fmt.Sprintf(_addRefreshStmt, viewName, start,
		kind.interval(policy.EndOffset), interval(policy.Schedule)))
}
