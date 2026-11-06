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

// Package ruleeventsrepo stores the secrulepb.RuleEvent records consumed
// from kafka: the secrule matches (sztopic.EventHTTPSecRule) and the rate
// limit rejections (sztopic.EventHTTPRateLimit), one table each.
package ruleeventsrepo

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/histogram"
	"github.com/sentinez/controlplane/pkg/schema"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/szstorage/dbx"
	"github.com/sentinez/shared/szstorage/dbx/postgres"
	"github.com/sentinez/shared/szstorage/dbx/timescale"
)

var _ IRuleEvent = (*RuleEvents)(nil)

// _columns the columns of a rule event table, both tables sharing them, in
// the order of values and scanOne.
var _columns = schema.AnalyticSecRulesEvent().ColumnNames()

var _filterColumns = eventfilter.Columns{
	ResourceId:  secrulepb.RuleEvent_ResourceId,
	MatchedTime: secrulepb.RuleEvent_MatchedTime,
}

type IRuleEvent interface {
	// Create inserts the events, which have ids, in one statement; an
	// event whose id and matched time are stored already is skipped, so
	// inserting a batch again is a no-op.
	Create(ctx context.Context, events ...*secrulepb.RuleEvent) error
	Get(ctx context.Context, id string) (*secrulepb.RuleEvent, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IRuleEvent

	List(ctx context.Context,
		filter *eventfilter.Filter) ([]*secrulepb.RuleEvent, error)
	// Histogram counts the events per bucket and resource.
	Histogram(ctx context.Context,
		filter *histogram.Filter) ([]*histogram.Point, error)
}

type RuleEvents struct {
	storage   dbx.Database[secrulepb.RuleEvent]
	histogram *histogram.Histogram
	// onConflict the conflict clause of Create, of the hypertable.
	onConflict string
}

// NewSecRules returns the repo of the sztable.AnalyticSecRulesEvent
// hypertable, and of its histograms.
func NewSecRules(ctx context.Context,
	appConf *settingpb.Config) (IRuleEvent, error) {
	return newRuleEvents(ctx, appConf, schema.AnalyticSecRulesEvent())
}

// NewRateLimits returns the repo of the sztable.AnalyticRateLimiterEvent
// hypertable, and of its histograms.
func NewRateLimits(ctx context.Context,
	appConf *settingpb.Config) (IRuleEvent, error) {
	return newRuleEvents(ctx, appConf, schema.AnalyticRateLimiterEvent())
}

func (r *RuleEvents) WithTX(tx *postgres.TxSession) IRuleEvent {
	return &RuleEvents{
		storage:    postgres.WithTx(tx, r.storage),
		histogram:  r.histogram,
		onConflict: r.onConflict,
	}
}

func (r *RuleEvents) Create(ctx context.Context,
	events ...*secrulepb.RuleEvent) error {
	if len(events) == 0 {
		return nil
	}

	query := sq.Insert(r.storage.Table()).Columns(_columns...).
		Suffix(r.onConflict)
	for _, event := range events {
		query = query.Values(values(event)...)
	}

	_, err := r.storage.Exec(ctx, query)
	return err
}

func (r *RuleEvents) Get(ctx context.Context,
	id string) (*secrulepb.RuleEvent, error) {
	builder := r.selectQuery().Where(sq.Eq{secrulepb.RuleEvent_Id: id})
	return r.storage.Select(ctx, builder, scanOne)
}

func (r *RuleEvents) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

func (r *RuleEvents) List(ctx context.Context,
	filter *eventfilter.Filter) ([]*secrulepb.RuleEvent, error) {
	builder := filter.Apply(r.selectQuery(), _filterColumns)
	return r.storage.CollectRows(ctx, builder, scan)
}

func (r *RuleEvents) Histogram(ctx context.Context,
	filter *histogram.Filter) ([]*histogram.Point, error) {
	return r.histogram.Points(ctx, filter)
}

// newRuleEvents returns the repo of the hypertable of sch, and of its
// histograms.
func newRuleEvents(ctx context.Context, appConf *settingpb.Config,
	sch dbx.Schema) (IRuleEvent, error) {
	ht := schema.EventHypertable(sch, secrulepb.RuleEvent_MatchedTime)

	storage, err := timescale.New[secrulepb.RuleEvent](ctx, appConf, ht)
	if err != nil {
		return nil, err
	}

	hist, err := histogram.New(ctx, appConf, histogram.Spec{Source: ht})
	if err != nil {
		return nil, err
	}

	return &RuleEvents{
		storage:    storage,
		histogram:  hist,
		onConflict: ht.OnConflictDoNothing(),
	}, nil
}

func (r *RuleEvents) selectQuery() sq.SelectBuilder {
	return sq.Select(_columns...).From(r.storage.Table())
}

// values the values of event, in the order of _columns.
func values(event *secrulepb.RuleEvent) []any {
	return []any{
		event.GetId(),
		event.GetMatchedSecruleIds(),
		event.GetMessages(),
		event.GetResourceId(),
		event.GetTenantId(),
		event.GetMatchedTime(),
		event.GetAction(),
		event.GetService(),
		event.GetRequestDomain(),
		event.GetRequestScheme(),
		event.GetRequestPath(),
		event.GetRequestQuery(),
		event.GetRequestIp(),
		event.GetRequestId(),
		event.GetTags(),
	}
}

func scan(rows dbx.Rows) ([]*secrulepb.RuleEvent, error) {
	var events []*secrulepb.RuleEvent
	for rows.Next() {
		event, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func scanOne(row dbx.Row) (*secrulepb.RuleEvent, error) {
	var event secrulepb.RuleEvent

	err := row.Scan(
		&event.Id,
		&event.MatchedSecruleIds,
		&event.Messages,
		&event.ResourceId,
		&event.TenantId,
		&event.MatchedTime,
		&event.Action,
		&event.Service,
		&event.RequestDomain,
		&event.RequestScheme,
		&event.RequestPath,
		&event.RequestQuery,
		&event.RequestIp,
		&event.RequestId,
		&event.Tags,
	)
	if err != nil {
		return nil, err
	}

	return &event, nil
}
