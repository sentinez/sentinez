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

// Package ruleseteventsrepo stores the rulesetpb.RulesetEvent
// records consumed from kafka (sztopic.EventHTTPRulesets).
package ruleseteventsrepo

import (
	"context"

	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventreader"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/topvalues"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var _ IRulesetEvent = (*RulesetEvents)(nil)

// _spec the reader of the events: their histogram is split by behavior.
var _spec = eventreader.Spec[rulesetpb.RulesetEvent]{
	Hypertable: schema.EventHypertable(
		schema.AnalyticRulesetsEvent(), rulesetpb.RulesetEvent_MatchedTime),
	Key: rulesetpb.RulesetEvent_Behavior,
	Columns: eventfilter.Columns{
		RequestId:   rulesetpb.RulesetEvent_RequestId,
		ResourceId:  rulesetpb.RulesetEvent_ResourceId,
		MatchedTime: rulesetpb.RulesetEvent_MatchedTime,
		Fields:      _fields,
	},
	Select:   rulesetpb.RulesetEventColumns(),
	Scan:     rulesetpb.ScanRulesetEvent[dbx.Rows],
	ScanOne:  rulesetpb.ScanOneRulesetEvent[dbx.Row],
	Values:   rulesetpb.RulesetEventValues,
	IdColumn: rulesetpb.RulesetEvent_Id,
}

// _fields the fields of the events filters match and top values count,
// in this order.
var _fields = []eventfilter.Field{
	eventfilter.TypedColumn(rulesetpb.RulesetEvent_MatchedRuleIds, dbx.IntArr),
	eventfilter.TypedColumn(rulesetpb.RulesetEvent_Severities, dbx.TextArr),
	eventfilter.TypedColumn(rulesetpb.RulesetEvent_Behavior, dbx.Int4),
	eventfilter.TypedColumn(rulesetpb.RulesetEvent_TotalScore, dbx.Int4),
	eventfilter.Column(rulesetpb.RulesetEvent_RequestIp),
	eventfilter.Column(rulesetpb.RulesetEvent_RequestDomain),
	eventfilter.Column(rulesetpb.RulesetEvent_RequestPath),
}

type IRulesetEvent interface {
	// Create inserts the events, which have ids, in one statement; an
	// event whose id and matched time are stored already is skipped, so
	// inserting a batch again is a no-op.
	Create(ctx context.Context, events ...*rulesetpb.RulesetEvent) error
	Get(ctx context.Context, id string) (*rulesetpb.RulesetEvent, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IRulesetEvent

	List(ctx context.Context,
		filter *eventfilter.Filter) ([]*rulesetpb.RulesetEvent, error)
	// Histogram counts the events per bucket, resource and behavior
	// (histogram.Point.Key).
	// The events of events only when set (see eventreader.Reader).
	Histogram(ctx context.Context, filter *histogram.Filter,
		events *eventfilter.Filter) ([]*histogram.Point, error)
	topvalues.TopValuer
}

type RulesetEvents struct {
	*eventreader.Reader[rulesetpb.RulesetEvent]
}

// New returns the repo of the sztable.AnalyticRulesetsEvent
// hypertable, and of its histograms.
func New(ctx context.Context,
	appConf *settingpb.Config) (IRulesetEvent, error) {
	reader, err := eventreader.New(ctx, appConf, _spec)
	if err != nil {
		return nil, err
	}

	return &RulesetEvents{Reader: reader}, nil
}

func (r *RulesetEvents) WithTX(tx *postgres.TxSession) IRulesetEvent {
	return &RulesetEvents{Reader: r.WithTx(tx)}
}
