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

// Package secruleeventsrepo stores the secrulepb.SecRuleEvent records consumed
// from kafka: the secrule matches (sztopic.EventHTTPSecRule) and the rate
// limit rejections (sztopic.EventHTTPRateLimit), one table each.
package secruleeventsrepo

import (
	"context"

	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventreader"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/topvalues"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var _ ISecRuleEvent = (*SecRuleEvents)(nil)

var _filterColumns = eventfilter.Columns{
	RequestId:   secrulepb.SecRuleEvent_RequestId,
	ResourceId:  secrulepb.SecRuleEvent_ResourceId,
	MatchedTime: secrulepb.SecRuleEvent_MatchedTime,
	Fields:      _fields,
}

// _fields the fields of the events filters match and top values count,
// in this order; the action is its type, a rule.v1.ActionType number.
var _fields = []eventfilter.Field{
	eventfilter.TypedColumn(secrulepb.SecRuleEvent_MatchedSecruleIds,
		dbx.TextArr),
	{
		Name: secrulepb.SecRuleEvent_Action,
		Expr: secrulepb.SecRuleEvent_Action + "->>'type'",
	},
	eventfilter.Column(secrulepb.SecRuleEvent_RequestIp),
	eventfilter.Column(secrulepb.SecRuleEvent_RequestDomain),
	eventfilter.Column(secrulepb.SecRuleEvent_RequestPath),
}

// specOf the reader of the events of the table of sch, whose histogram is
// not split by key.
func specOf(sch dbx.Schema) eventreader.Spec[secrulepb.SecRuleEvent] {
	return eventreader.Spec[secrulepb.SecRuleEvent]{
		Hypertable: schema.EventHypertable(sch,
			secrulepb.SecRuleEvent_MatchedTime),
		Columns:  _filterColumns,
		Select:   secrulepb.SecRuleEventColumns(),
		Scan:     secrulepb.ScanSecRuleEvent[dbx.Rows],
		ScanOne:  secrulepb.ScanOneSecRuleEvent[dbx.Row],
		Values:   secrulepb.SecRuleEventValues,
		IdColumn: secrulepb.SecRuleEvent_Id,
	}
}

type ISecRuleEvent interface {
	// Create inserts the events, which have ids, in one statement; an
	// event whose id and matched time are stored already is skipped, so
	// inserting a batch again is a no-op.
	Create(ctx context.Context, events ...*secrulepb.SecRuleEvent) error
	Get(ctx context.Context, id string) (*secrulepb.SecRuleEvent, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) ISecRuleEvent

	List(ctx context.Context,
		filter *eventfilter.Filter) ([]*secrulepb.SecRuleEvent, error)
	// Histogram counts the events per bucket and resource.
	// The events of events only when set (see eventreader.Reader).
	Histogram(ctx context.Context, filter *histogram.Filter,
		events *eventfilter.Filter) ([]*histogram.Point, error)
	topvalues.TopValuer
}

type SecRuleEvents struct {
	*eventreader.Reader[secrulepb.SecRuleEvent]
}

// NewSecRules returns the repo of the sztable.AnalyticSecRulesEvent
// hypertable, and of its histograms.
func NewSecRules(ctx context.Context,
	appConf *settingpb.Config) (ISecRuleEvent, error) {
	return newSecRuleEvents(ctx, appConf, schema.AnalyticSecRulesEvent())
}

// NewRateLimits returns the repo of the sztable.AnalyticRateLimiterEvent
// hypertable, and of its histograms.
func NewRateLimits(ctx context.Context,
	appConf *settingpb.Config) (ISecRuleEvent, error) {
	return newSecRuleEvents(ctx, appConf, schema.AnalyticRateLimiterEvent())
}

// newSecRuleEvents returns the repo of the hypertable of sch, and of its
// histograms.
func newSecRuleEvents(ctx context.Context, appConf *settingpb.Config,
	sch dbx.Schema) (ISecRuleEvent, error) {
	reader, err := eventreader.New(ctx, appConf, specOf(sch))
	if err != nil {
		return nil, err
	}

	return &SecRuleEvents{Reader: reader}, nil
}

func (r *SecRuleEvents) WithTX(tx *postgres.TxSession) ISecRuleEvent {
	return &SecRuleEvents{Reader: r.WithTx(tx)}
}
