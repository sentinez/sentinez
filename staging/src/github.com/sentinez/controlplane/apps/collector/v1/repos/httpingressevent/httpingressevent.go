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

// Package httpingresseventrepo stores the httppb.IngressEvent records
// consumed from kafka (sztopic.EventHTTPIngress): one per request of the
// edge.
package httpingresseventrepo

import (
	"context"

	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventreader"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/topvalues"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var _ IHttpIngressEvent = (*HttpIngressEvent)(nil)

// _spec the reader of the events: their histogram is split by status.
var _spec = eventreader.Spec[httppb.IngressEvent]{
	Hypertable: schema.EventHypertable(
		schema.AnalyticHttpIngressEvent(), httppb.IngressEvent_Timestamp),
	Key: httppb.IngressEvent_Status,
	Columns: eventfilter.Columns{
		RequestId:   httppb.IngressEvent_Id,
		ResourceId:  httppb.IngressEvent_ResourceId,
		MatchedTime: httppb.IngressEvent_Timestamp,
		Fields:      _fields,
	},
	Select:   httppb.IngressEventColumns(),
	Scan:     httppb.ScanIngressEvent[dbx.Rows],
	ScanOne:  httppb.ScanOneIngressEvent[dbx.Row],
	Values:   httppb.IngressEventValues,
	IdColumn: httppb.IngressEvent_Id,
}

// _visitors the requests per client ip: the visitors of a resource.
var _visitors = histogram.ValueSpec{
	Source: _spec.Hypertable,
	Column: httppb.IngressEvent_ClientIp,
}

// _fields the fields of the events filters match and top values count,
// in this order.
var _fields = []eventfilter.Field{
	eventfilter.Column(httppb.IngressEvent_Method),
	eventfilter.TypedColumn(httppb.IngressEvent_Status, dbx.Int4),
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
	// The events of events only when set (see eventreader.Reader).
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
	*eventreader.Reader[httppb.IngressEvent]

	visitors *histogram.Values
}

// New returns the repo of the sztable.AnalyticHttpIngressEvent
// hypertable, and of its histograms.
func New(ctx context.Context,
	appConf *settingpb.Config) (IHttpIngressEvent, error) {
	reader, err := eventreader.New(ctx, appConf, _spec)
	if err != nil {
		return nil, err
	}

	visitors, err := histogram.NewValues(ctx, appConf, _visitors)
	if err != nil {
		return nil, err
	}

	return &HttpIngressEvent{Reader: reader, visitors: visitors}, nil
}

func (r *HttpIngressEvent) WithTX(tx *postgres.TxSession) IHttpIngressEvent {
	return &HttpIngressEvent{Reader: r.WithTx(tx), visitors: r.visitors}
}

func (r *HttpIngressEvent) UniqueVisitors(ctx context.Context,
	filter *histogram.Filter) ([]*histogram.Point, error) {
	return r.visitors.Distinct(ctx, filter)
}

func (r *HttpIngressEvent) Visitors(ctx context.Context,
	filter *histogram.Filter) ([]*histogram.Point, error) {
	return r.visitors.Points(ctx, filter)
}
