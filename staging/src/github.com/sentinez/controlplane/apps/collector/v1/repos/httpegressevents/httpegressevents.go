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

	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventreader"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/topvalues"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
	analyticpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var _ IHttpEgressEvent = (*HttpEgressEvents)(nil)

// _spec the reader of the events: their histogram is split by status.
var _spec = eventreader.Spec[httppb.EgressEvent]{
	Hypertable: schema.EventHypertable(
		schema.AnalyticHttpEgressEvent(), httppb.EgressEvent_Timestamp),
	Key: httppb.EgressEvent_Status,
	Columns: eventfilter.Columns{
		RequestId:   httppb.EgressEvent_Id,
		ResourceId:  httppb.EgressEvent_ResourceId,
		MatchedTime: httppb.EgressEvent_Timestamp,
		Fields:      _fields,
	},
	Select:   httppb.EgressEventColumns(),
	Scan:     httppb.ScanEgressEvent[dbx.Rows],
	ScanOne:  httppb.ScanOneEgressEvent[dbx.Row],
	Values:   httppb.EgressEventValues,
	IdColumn: httppb.EgressEvent_Id,
}

// _fields the fields of the events filters match and top values count,
// in this order.
var _fields = []eventfilter.Field{
	eventfilter.Column(httppb.EgressEvent_Upstream),
	eventfilter.TypedColumn(httppb.EgressEvent_Status, dbx.Int4),
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
	// The events of events only when set (see eventreader.Reader).
	Histogram(ctx context.Context, filter *histogram.Filter,
		events *eventfilter.Filter) ([]*histogram.Point, error)
	LatencyReader
	topvalues.TopValuer
}

type HttpEgressEvents struct {
	*eventreader.Reader[httppb.EgressEvent]

	latency *latencyReader
}

// New returns the repo of the sztable.AnalyticHttpEgressEvent
// hypertable, and of its histograms.
func New(ctx context.Context,
	appConf *settingpb.Config) (IHttpEgressEvent, error) {
	reader, err := eventreader.New(ctx, appConf, _spec)
	if err != nil {
		return nil, err
	}

	latency, err := openLatency(ctx, appConf,
		_spec.Hypertable.Schema.Table, _spec.Columns)
	if err != nil {
		return nil, err
	}

	return &HttpEgressEvents{Reader: reader, latency: latency}, nil
}

func (r *HttpEgressEvents) WithTX(tx *postgres.TxSession) IHttpEgressEvent {
	return &HttpEgressEvents{Reader: r.WithTx(tx), latency: r.latency}
}

// Latency implements LatencyReader.
func (r *HttpEgressEvents) Latency(ctx context.Context,
	filter *eventfilter.Filter) ([]*analyticpb.UpstreamLatency, error) {
	return r.latency.Latency(ctx, filter)
}
