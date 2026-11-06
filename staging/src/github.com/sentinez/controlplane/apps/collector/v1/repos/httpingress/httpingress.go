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

// Package httpingressrepo stores the httppb.IngressEvent records consumed from
// kafka (sztopic.EventHTTPIngress): one per request of the edge.
package httpingressrepo

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/histogram"
	"github.com/sentinez/controlplane/pkg/schema"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/szstorage/dbx"
	"github.com/sentinez/shared/szstorage/dbx/postgres"
	"github.com/sentinez/shared/szstorage/dbx/timescale"
)

var _ IHttpIngressEvent = (*HttpIngressEvent)(nil)

// _columns the columns of the table, in the order of values and scanOne.
var _columns = schema.AnalyticHttpIngressEvent().ColumnNames()

var _hypertable = schema.EventHypertable(
	schema.AnalyticHttpIngressEvent(), httppb.IngressEvent_Timestamp)

var _histogram = histogram.Spec{
	Source: _hypertable,
	Key:    httppb.IngressEvent_Status,
}

var _filterColumns = eventfilter.Columns{
	ResourceId:  httppb.IngressEvent_ResourceId,
	MatchedTime: httppb.IngressEvent_Timestamp,
}

type IHttpIngressEvent interface {
	// Create inserts the events, which have ids, in one statement; an
	// event whose id and timestamp are stored already is skipped, so
	// inserting a batch again is a no-op.
	Create(ctx context.Context, events ...*httppb.IngressEvent) error
	Get(ctx context.Context, id string) (*httppb.IngressEvent, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IHttpIngressEvent

	List(ctx context.Context,
		filter *eventfilter.Filter) ([]*httppb.IngressEvent, error)
	// Histogram counts the requests per bucket, resource and http
	// status (histogram.Point.Key).
	Histogram(ctx context.Context,
		filter *histogram.Filter) ([]*histogram.Point, error)
}

type HttpIngressEvent struct {
	storage   dbx.Database[httppb.IngressEvent]
	histogram *histogram.Histogram
}

// New returns the repo of the sztable.AnalyticHttpIngressEvent
// hypertable, and of its histograms.
func New(ctx context.Context,
	appConf *settingpb.Config) (IHttpIngressEvent, error) {
	storage, err := timescale.New[httppb.IngressEvent](ctx, appConf, _hypertable)
	if err != nil {
		return nil, err
	}

	hist, err := histogram.New(ctx, appConf, _histogram)
	if err != nil {
		return nil, err
	}

	return &HttpIngressEvent{storage: storage, histogram: hist}, nil
}

func (r *HttpIngressEvent) WithTX(tx *postgres.TxSession) IHttpIngressEvent {
	return &HttpIngressEvent{
		storage:   postgres.WithTx(tx, r.storage),
		histogram: r.histogram,
	}
}

func (r *HttpIngressEvent) Create(ctx context.Context,
	events ...*httppb.IngressEvent) error {
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

func (r *HttpIngressEvent) Get(ctx context.Context,
	id string) (*httppb.IngressEvent, error) {
	builder := r.selectQuery().Where(sq.Eq{httppb.IngressEvent_Id: id})
	return r.storage.Select(ctx, builder, scanOne)
}

func (r *HttpIngressEvent) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

// List the events of filter, the latest request first.
func (r *HttpIngressEvent) List(ctx context.Context,
	filter *eventfilter.Filter) ([]*httppb.IngressEvent, error) {
	builder := filter.Apply(r.selectQuery(), _filterColumns)
	return r.storage.CollectRows(ctx, builder, scan)
}

func (r *HttpIngressEvent) Histogram(ctx context.Context,
	filter *histogram.Filter) ([]*histogram.Point, error) {
	return r.histogram.Points(ctx, filter)
}

func (r *HttpIngressEvent) selectQuery() sq.SelectBuilder {
	return sq.Select(_columns...).From(r.storage.Table())
}

// values the values of event, in the order of _columns.
func values(event *httppb.IngressEvent) []any {
	return []any{
		event.GetId(),
		event.GetMethod(),
		event.GetScheme(),
		event.GetHost(),
		event.GetPath(),
		event.GetQueries(),
		event.GetHeaders(),
		event.GetProtocol(),
		event.GetRemoteAddress(),
		event.GetStatus(),
		event.GetUserAgent(),
		event.GetAsn(),
		event.GetCountry(),
		event.GetContentType(),
		event.GetTimestamp(),
		event.GetFingerprint(),
		event.GetClientIp(),
		event.GetResourceId(),
		event.GetTenantId(),
	}
}

func scan(rows dbx.Rows) ([]*httppb.IngressEvent, error) {
	var events []*httppb.IngressEvent
	for rows.Next() {
		event, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func scanOne(row dbx.Row) (*httppb.IngressEvent, error) {
	var event httppb.IngressEvent

	err := row.Scan(
		&event.Id,
		&event.Method,
		&event.Scheme,
		&event.Host,
		&event.Path,
		&event.Queries,
		&event.Headers,
		&event.Protocol,
		&event.RemoteAddress,
		&event.Status,
		&event.UserAgent,
		&event.Asn,
		&event.Country,
		&event.ContentType,
		&event.Timestamp,
		&event.Fingerprint,
		&event.ClientIp,
		&event.ResourceId,
		&event.TenantId,
	)
	if err != nil {
		return nil, err
	}

	return &event, nil
}
