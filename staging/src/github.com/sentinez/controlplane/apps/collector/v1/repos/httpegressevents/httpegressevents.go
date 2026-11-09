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
	"github.com/sentinez/controlplane/apps/collector/v1/repos/topvalues"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/szstorage/dbx/timescale"
	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var _ IHttpEgressEvent = (*HttpEgressEvents)(nil)

var _hypertable = schema.EventHypertable(
	schema.AnalyticHttpEgressEvent(), httppb.EgressEvent_Timestamp)

var _histogram = histogram.Spec{
	Source: _hypertable,
	Key:    httppb.EgressEvent_Status,
}

var _filterColumns = eventfilter.Columns{
	RequestId:   httppb.EgressEvent_Id,
	ResourceId:  httppb.EgressEvent_ResourceId,
	MatchedTime: httppb.EgressEvent_Timestamp,
	Fields:      _fields,
}

// _fields the fields of the events filters match and top values count,
// in this order.
var _fields = []eventfilter.Field{
	eventfilter.Column(httppb.EgressEvent_Upstream),
	eventfilter.Column(httppb.EgressEvent_Status),
	eventfilter.Column(httppb.EgressEvent_Method),
	eventfilter.Column(httppb.EgressEvent_Path),
	eventfilter.Column(httppb.EgressEvent_ContentType),
	eventfilter.Column(httppb.EgressEvent_Protocol),
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
	// The events of events only when set (see eventfilter.Points).
	Histogram(ctx context.Context, filter *histogram.Filter,
		events *eventfilter.Filter) ([]*histogram.Point, error)
	topvalues.TopValuer
}

type HttpEgressEvents struct {
	storage   dbx.Database[httppb.EgressEvent]
	histogram *histogram.Histogram
	topValues *topvalues.Reader
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

	topValues, err := topvalues.Open(ctx, appConf, _hypertable.Schema.Table,
		_filterColumns)
	if err != nil {
		return nil, err
	}

	return &HttpEgressEvents{
		storage:   storage,
		histogram: hist,
		topValues: topValues,
	}, nil
}

func (r *HttpEgressEvents) WithTX(tx *postgres.TxSession) IHttpEgressEvent {
	return &HttpEgressEvents{
		storage:   postgres.WithTx(tx, r.storage),
		histogram: r.histogram,
		topValues: r.topValues,
	}
}

func (r *HttpEgressEvents) Create(ctx context.Context,
	events ...*httppb.EgressEvent) error {
	if len(events) == 0 {
		return nil
	}

	query, err := postgres.InsertRowsBuilder(r.storage,
		httppb.EgressEventColumns(), events, httppb.EgressEventValues)
	if err != nil {
		return err
	}

	_, err = r.storage.Exec(ctx, query.Suffix(_hypertable.OnConflictDoNothing()))
	return err
}

func (r *HttpEgressEvents) Get(ctx context.Context,
	id string) (*httppb.EgressEvent, error) {
	builder := r.selectQuery().Where(sq.Eq{httppb.EgressEvent_Id: id})
	return r.storage.Select(ctx, builder, httppb.ScanOneEgressEvent)
}

func (r *HttpEgressEvents) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

// List the events of filter, the latest request first.
func (r *HttpEgressEvents) List(ctx context.Context,
	filter *eventfilter.Filter) ([]*httppb.EgressEvent, error) {
	builder, err := filter.Apply(r.selectQuery(), _filterColumns)
	if err != nil {
		return nil, err
	}

	return r.storage.CollectRows(ctx, builder, httppb.ScanEgressEvent)
}

func (r *HttpEgressEvents) Histogram(ctx context.Context,
	filter *histogram.Filter,
	events *eventfilter.Filter) ([]*histogram.Point, error) {
	return eventfilter.Points(ctx, r.histogram, filter, events,
		_filterColumns)
}

func (r *HttpEgressEvents) TopValues(ctx context.Context,
	filter *eventfilter.Filter, names []string,
	limit uint64) (*topvalues.Result, error) {
	return r.topValues.TopValues(ctx, filter, names, limit)
}

func (r *HttpEgressEvents) selectQuery() sq.SelectBuilder {
	return sq.Select(httppb.EgressEventColumns()...).From(r.storage.Table())
}
