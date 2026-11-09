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

// Package coreruleeventsrepo stores the corerulesetpb.CoreRuleEvent
// records consumed from kafka (sztopic.EventHTTPRulesets).
package coreruleeventsrepo

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
	crspb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var _ ICoreRuleEvent = (*CoreRuleEvents)(nil)

// _columns the columns of the table, in the order of values and scanOne.
var _columns = schema.AnalyticCoreRulesetsEvent().ColumnNames()

var _hypertable = schema.EventHypertable(
	schema.AnalyticCoreRulesetsEvent(), crspb.CoreRuleEvent_MatchedTime)

var _histogram = histogram.Spec{
	Source: _hypertable,
	Key:    crspb.CoreRuleEvent_Behavior,
}

var _filterColumns = eventfilter.Columns{
	RequestId:   crspb.CoreRuleEvent_RequestId,
	ResourceId:  crspb.CoreRuleEvent_ResourceId,
	MatchedTime: crspb.CoreRuleEvent_MatchedTime,
	Fields:      _fields,
}

// _fields the fields of the events filters match and top values count,
// in this order.
var _fields = []eventfilter.Field{
	eventfilter.ArrayColumn(crspb.CoreRuleEvent_MatchedRuleIds),
	eventfilter.ArrayColumn(crspb.CoreRuleEvent_Severities),
	eventfilter.Column(crspb.CoreRuleEvent_Behavior),
	eventfilter.Column(crspb.CoreRuleEvent_TotalScore),
	eventfilter.Column(crspb.CoreRuleEvent_RequestIp),
	eventfilter.Column(crspb.CoreRuleEvent_RequestDomain),
	eventfilter.Column(crspb.CoreRuleEvent_RequestPath),
}

type ICoreRuleEvent interface {
	// Create inserts the events, which have ids, in one statement; an
	// event whose id and matched time are stored already is skipped, so
	// inserting a batch again is a no-op.
	Create(ctx context.Context, events ...*crspb.CoreRuleEvent) error
	Get(ctx context.Context, id string) (*crspb.CoreRuleEvent, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) ICoreRuleEvent

	List(ctx context.Context,
		filter *eventfilter.Filter) ([]*crspb.CoreRuleEvent, error)
	// Histogram counts the events per bucket, resource and behavior
	// (histogram.Point.Key).
	// The events of events only when set (see eventfilter.Points).
	Histogram(ctx context.Context, filter *histogram.Filter,
		events *eventfilter.Filter) ([]*histogram.Point, error)
	topvalues.TopValuer
}

type CoreRuleEvents struct {
	storage   dbx.Database[crspb.CoreRuleEvent]
	histogram *histogram.Histogram
	topValues *topvalues.Reader
}

// New returns the repo of the sztable.AnalyticCoreRulesetsEvent
// hypertable, and of its histograms.
func New(ctx context.Context,
	appConf *settingpb.Config) (ICoreRuleEvent, error) {
	storage, err := timescale.New[crspb.CoreRuleEvent](ctx, appConf,
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

	return &CoreRuleEvents{
		storage:   storage,
		histogram: hist,
		topValues: topValues,
	}, nil
}

func (r *CoreRuleEvents) WithTX(tx *postgres.TxSession) ICoreRuleEvent {
	return &CoreRuleEvents{
		storage:   postgres.WithTx(tx, r.storage),
		histogram: r.histogram,
		topValues: r.topValues,
	}
}

func (r *CoreRuleEvents) Create(ctx context.Context,
	events ...*crspb.CoreRuleEvent) error {
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

func (r *CoreRuleEvents) Get(ctx context.Context,
	id string) (*crspb.CoreRuleEvent, error) {
	builder := r.selectQuery().Where(sq.Eq{crspb.CoreRuleEvent_Id: id})
	return r.storage.Select(ctx, builder, scanOne)
}

func (r *CoreRuleEvents) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

func (r *CoreRuleEvents) List(ctx context.Context,
	filter *eventfilter.Filter) ([]*crspb.CoreRuleEvent, error) {
	builder, err := filter.Apply(r.selectQuery(), _filterColumns)
	if err != nil {
		return nil, err
	}

	return r.storage.CollectRows(ctx, builder, scan)
}

func (r *CoreRuleEvents) Histogram(ctx context.Context,
	filter *histogram.Filter,
	events *eventfilter.Filter) ([]*histogram.Point, error) {
	return eventfilter.Points(ctx, r.histogram, filter, events,
		_filterColumns)
}

func (r *CoreRuleEvents) TopValues(ctx context.Context,
	filter *eventfilter.Filter, names []string,
	limit uint64) (*topvalues.Result, error) {
	return r.topValues.TopValues(ctx, filter, names, limit)
}

func (r *CoreRuleEvents) selectQuery() sq.SelectBuilder {
	return sq.Select(_columns...).From(r.storage.Table())
}

// values the values of event, in the order of _columns.
func values(event *crspb.CoreRuleEvent) []any {
	return []any{
		event.GetId(),
		event.GetMatchedRuleIds(),
		event.GetSeverities(),
		event.GetMessages(),
		event.GetTotalScore(),
		event.GetMatchedTime(),
		event.GetResourceId(),
		event.GetTenantId(),
		event.GetService(),
		event.GetRequestIp(),
		event.GetRequestDomain(),
		event.GetRequestScheme(),
		event.GetRequestPath(),
		event.GetRequestQuery(),
		event.GetRequestId(),
		event.GetBehavior(),
	}
}

func scan(rows dbx.Rows) ([]*crspb.CoreRuleEvent, error) {
	var events []*crspb.CoreRuleEvent
	for rows.Next() {
		event, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func scanOne(row dbx.Row) (*crspb.CoreRuleEvent, error) {
	var event crspb.CoreRuleEvent

	err := row.Scan(
		&event.Id,
		&event.MatchedRuleIds,
		&event.Severities,
		&event.Messages,
		&event.TotalScore,
		&event.MatchedTime,
		&event.ResourceId,
		&event.TenantId,
		&event.Service,
		&event.RequestIp,
		&event.RequestDomain,
		&event.RequestScheme,
		&event.RequestPath,
		&event.RequestQuery,
		&event.RequestId,
		&event.Behavior,
	)
	if err != nil {
		return nil, err
	}

	return &event, nil
}
