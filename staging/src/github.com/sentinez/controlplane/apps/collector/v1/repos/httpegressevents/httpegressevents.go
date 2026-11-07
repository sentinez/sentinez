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

// Package httpegresseventsrepo stores the httppb.EgressEvent records
// consumed from kafka (sztopic.EventHTTPEgress): one per request the edge
// proxied to its upstream origin.
package httpegresseventsrepo

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/histogram"
	"github.com/sentinez/controlplane/pkg/schema"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/szstorage/dbx/timescale"
)

var _ IHttpEgressEvent = (*HttpEgressEvents)(nil)

// _columns the columns of the table, in the order of values and scanOne.
var _columns = schema.AnalyticHttpEgressEvent().ColumnNames()

var _hypertable = schema.EventHypertable(
	schema.AnalyticHttpEgressEvent(), httppb.EgressEvent_Timestamp)

var _histogram = histogram.Spec{
	Source: _hypertable,
	Key:    httppb.EgressEvent_Status,
}

var _filterColumns = eventfilter.Columns{
	ResourceId:  httppb.EgressEvent_ResourceId,
	MatchedTime: httppb.EgressEvent_Timestamp,
}

type IHttpEgressEvent interface {
	// Create inserts the events, which have ids, in one statement; an
	// event whose id and timestamp are stored already is skipped, so
	// inserting a batch again is a no-op.
	Create(ctx context.Context, events ...*httppb.EgressEvent) error
	Get(ctx context.Context, id string) (*httppb.EgressEvent, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IHttpEgressEvent

	List(ctx context.Context,
		filter *eventfilter.Filter) ([]*httppb.EgressEvent, error)
	// Histogram counts the proxied requests per bucket, resource and
	// status of the origin (histogram.Point.Key).
	Histogram(ctx context.Context,
		filter *histogram.Filter) ([]*histogram.Point, error)
}

type HttpEgressEvents struct {
	storage   dbx.Database[httppb.EgressEvent]
	histogram *histogram.Histogram
}

// New returns the repo of the sztable.AnalyticHttpEgressEvent
// hypertable, and of its histograms.
func New(ctx context.Context,
	appConf *settingpb.Config) (IHttpEgressEvent, error) {
	storage, err := timescale.New[httppb.EgressEvent](ctx, appConf,
		_hypertable)
	if err != nil {
		return nil, err
	}

	hist, err := histogram.New(ctx, appConf, _histogram)
	if err != nil {
		return nil, err
	}

	return &HttpEgressEvents{storage: storage, histogram: hist}, nil
}

func (r *HttpEgressEvents) WithTX(tx *postgres.TxSession) IHttpEgressEvent {
	return &HttpEgressEvents{
		storage:   postgres.WithTx(tx, r.storage),
		histogram: r.histogram,
	}
}

func (r *HttpEgressEvents) Create(ctx context.Context,
	events ...*httppb.EgressEvent) error {
	if len(events) == 0 {
		return nil
	}

	query := sq.Insert(r.storage.Table()).Columns(_columns...).
		Suffix(_hypertable.OnConflictDoNothing())
	for _, event := range events {
		query = query.Values(values(event)...)
	}

	_, err := r.storage.Exec(ctx, query)
	return err
}

func (r *HttpEgressEvents) Get(ctx context.Context,
	id string) (*httppb.EgressEvent, error) {
	builder := r.selectQuery().Where(sq.Eq{httppb.EgressEvent_Id: id})
	return r.storage.Select(ctx, builder, scanOne)
}

func (r *HttpEgressEvents) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

// List the events of filter, the latest request first.
func (r *HttpEgressEvents) List(ctx context.Context,
	filter *eventfilter.Filter) ([]*httppb.EgressEvent, error) {
	builder := filter.Apply(r.selectQuery(), _filterColumns)
	return r.storage.CollectRows(ctx, builder, scan)
}

func (r *HttpEgressEvents) Histogram(ctx context.Context,
	filter *histogram.Filter) ([]*histogram.Point, error) {
	return r.histogram.Points(ctx, filter)
}

func (r *HttpEgressEvents) selectQuery() sq.SelectBuilder {
	return sq.Select(_columns...).From(r.storage.Table())
}

// values the values of event, in the order of _columns.
func values(event *httppb.EgressEvent) []any {
	return []any{
		event.GetId(),
		event.GetUpstream(),
		event.GetMethod(),
		event.GetScheme(),
		event.GetPath(),
		event.GetQueries(),
		event.GetStatus(),
		event.GetContentType(),
		event.GetContentLength(),
		event.GetLatencyMicros(),
		event.GetTimestamp(),
		event.GetResourceId(),
		event.GetTenantId(),
		event.GetProtocol(),
	}
}

func scan(rows dbx.Rows) ([]*httppb.EgressEvent, error) {
	var events []*httppb.EgressEvent
	for rows.Next() {
		event, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func scanOne(row dbx.Row) (*httppb.EgressEvent, error) {
	var event httppb.EgressEvent

	err := row.Scan(
		&event.Id,
		&event.Upstream,
		&event.Method,
		&event.Scheme,
		&event.Path,
		&event.Queries,
		&event.Status,
		&event.ContentType,
		&event.ContentLength,
		&event.LatencyMicros,
		&event.Timestamp,
		&event.ResourceId,
		&event.TenantId,
		&event.Protocol,
	)
	if err != nil {
		return nil, err
	}

	return &event, nil
}
