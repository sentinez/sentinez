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

// Package httpingresseventrepo stores the httppb.IngressEvent records consumed from
// kafka (sztopic.EventHTTPIngress): one per request of the edge.
package httpingresseventrepo

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

var _ IHttpIngressEvent = (*HttpIngressEvent)(nil)

var _hypertable = schema.EventHypertable(
	schema.AnalyticHttpIngressEvent(), httppb.IngressEvent_Timestamp)

var _histogramStatus = histogram.Spec{
	Source: _hypertable,
	Key:    httppb.IngressEvent_Status,
}

// _visitors the requests per client ip: the visitors of a resource.
var _visitors = histogram.ValueSpec{
	Source: _hypertable,
	Column: httppb.IngressEvent_ClientIp,
}

var _filterColumns = eventfilter.Columns{
	RequestId:   httppb.IngressEvent_Id,
	ResourceId:  httppb.IngressEvent_ResourceId,
	MatchedTime: httppb.IngressEvent_Timestamp,
	Fields:      _fields,
}

// _fields the fields of the events filters match and top values count,
// in this order.
var _fields = []eventfilter.Field{
	eventfilter.Column(httppb.IngressEvent_Method),
	eventfilter.Column(httppb.IngressEvent_Status),
	eventfilter.Column(httppb.IngressEvent_Host),
	eventfilter.Column(httppb.IngressEvent_ClientIp),
	eventfilter.Column(httppb.IngressEvent_Country),
	eventfilter.Column(httppb.IngressEvent_Protocol),
	eventfilter.Column(httppb.IngressEvent_Path),
	eventfilter.Column(httppb.IngressEvent_UserAgent),
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
	// The events of events only when set (see eventfilter.Points).
	Histogram(ctx context.Context, filter *histogram.Filter,
		events *eventfilter.Filter) ([]*histogram.Point, error)
	VisitorReader
	topvalues.TopValuer
}

// VisitorReader reads the visitors of the requests: their client ips.
type VisitorReader interface {
	// UniqueVisitors counts the distinct client ips per bucket and
	// resource.
	UniqueVisitors(ctx context.Context,
		filter *histogram.Filter) ([]*histogram.Point, error)
	// Visitors counts the requests per bucket, resource and client ip
	// (histogram.Point.Value), the most frequent ip first.
	Visitors(ctx context.Context,
		filter *histogram.Filter) ([]*histogram.Point, error)
}

type HttpIngressEvent struct {
	storage         dbx.Database[httppb.IngressEvent]
	histogramStatus *histogram.Histogram
	visitors        *histogram.Values
	topValues       *topvalues.Reader
}

// New returns the repo of the sztable.AnalyticHttpIngressEvent
// hypertable, and of its histograms.
func New(ctx context.Context,
	appConf *settingpb.Config) (IHttpIngressEvent, error) {
	storage, err := timescale.New[httppb.IngressEvent](ctx, appConf, _hypertable)
	if err != nil {
		return nil, err
	}

	histStatus, err := histogram.New(ctx, appConf, _histogramStatus)
	if err != nil {
		return nil, err
	}

	visitors, err := histogram.NewValues(ctx, appConf, _visitors)
	if err != nil {
		return nil, err
	}

	topValues, err := topvalues.Open(ctx, appConf, _hypertable.Schema.Table,
		_filterColumns)
	if err != nil {
		return nil, err
	}

	return &HttpIngressEvent{
		storage:         storage,
		histogramStatus: histStatus,
		visitors:        visitors,
		topValues:       topValues,
	}, nil
}

func (r *HttpIngressEvent) WithTX(tx *postgres.TxSession) IHttpIngressEvent {
	return &HttpIngressEvent{
		storage:         postgres.WithTx(tx, r.storage),
		histogramStatus: r.histogramStatus,
		visitors:        r.visitors,
		topValues:       r.topValues,
	}
}

func (r *HttpIngressEvent) Create(ctx context.Context,
	events ...*httppb.IngressEvent) error {
	if len(events) == 0 {
		return nil
	}

	query, err := postgres.InsertRowsBuilder(r.storage,
		httppb.IngressEventColumns(), events, httppb.IngressEventValues)
	if err != nil {
		return err
	}

	_, err = r.storage.Exec(ctx, query.Suffix(_hypertable.OnConflictDoNothing()))
	return err
}

func (r *HttpIngressEvent) Get(ctx context.Context,
	id string) (*httppb.IngressEvent, error) {
	builder := r.selectQuery().Where(sq.Eq{httppb.IngressEvent_Id: id})
	return r.storage.Select(ctx, builder, httppb.ScanOneIngressEvent)
}

func (r *HttpIngressEvent) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

// List the events of filter, the latest request first.
func (r *HttpIngressEvent) List(ctx context.Context,
	filter *eventfilter.Filter) ([]*httppb.IngressEvent, error) {
	builder, err := filter.Apply(r.selectQuery(), _filterColumns)
	if err != nil {
		return nil, err
	}

	return r.storage.CollectRows(ctx, builder, httppb.ScanIngressEvent)
}

func (r *HttpIngressEvent) Histogram(ctx context.Context,
	filter *histogram.Filter,
	events *eventfilter.Filter) ([]*histogram.Point, error) {
	return eventfilter.Points(ctx, r.histogramStatus, filter, events,
		_filterColumns)
}

func (r *HttpIngressEvent) UniqueVisitors(ctx context.Context,
	filter *histogram.Filter) ([]*histogram.Point, error) {
	return r.visitors.Distinct(ctx, filter)
}

func (r *HttpIngressEvent) Visitors(ctx context.Context,
	filter *histogram.Filter) ([]*histogram.Point, error) {
	return r.visitors.Points(ctx, filter)
}

func (r *HttpIngressEvent) TopValues(ctx context.Context,
	filter *eventfilter.Filter, names []string,
	limit uint64) (*topvalues.Result, error) {
	return r.topValues.TopValues(ctx, filter, names, limit)
}

func (r *HttpIngressEvent) selectQuery() sq.SelectBuilder {
	return sq.Select(httppb.IngressEventColumns()...).From(r.storage.Table())
}
