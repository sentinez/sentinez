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

// Package secruleseteventsrepo stores the rulesetpb.SecRulesetEvent
// records consumed from kafka (sztopic.EventHTTPSecRuleset).
package secruleseteventsrepo

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

var _ ISecRulesetEvent = (*SecRulesetEvents)(nil)

// _spec the reader of the events: their histogram is split by behavior.
var _spec = eventreader.Spec[rulesetpb.SecRulesetEvent]{
	Hypertable: schema.EventHypertable(
		schema.AnalyticSecRulesetsEvent(),
		rulesetpb.SecRulesetEvent_MatchedTime,
	),
	Key: rulesetpb.SecRulesetEvent_Behavior,
	Columns: eventfilter.Columns{
		RequestId:   rulesetpb.SecRulesetEvent_RequestId,
		ResourceId:  rulesetpb.SecRulesetEvent_ResourceId,
		MatchedTime: rulesetpb.SecRulesetEvent_MatchedTime,
		Fields:      _fields,
	},
	Select:   rulesetpb.SecRulesetEventColumns(),
	Scan:     rulesetpb.ScanSecRulesetEvent[dbx.Rows],
	ScanOne:  rulesetpb.ScanOneSecRulesetEvent[dbx.Row],
	Values:   rulesetpb.SecRulesetEventValues,
	IdColumn: rulesetpb.SecRulesetEvent_Id,
}

// _fields the fields of the events filters match and top values count,
// in this order.
var _fields = []eventfilter.Field{
	eventfilter.TypedColumn(
		rulesetpb.SecRulesetEvent_MatchedRuleIds, dbx.IntArr),
	eventfilter.TypedColumn(rulesetpb.SecRulesetEvent_Severities, dbx.TextArr),
	eventfilter.TypedColumn(rulesetpb.SecRulesetEvent_Behavior, dbx.Int4),
	eventfilter.TypedColumn(rulesetpb.SecRulesetEvent_TotalScore, dbx.Int4),
	eventfilter.Column(rulesetpb.SecRulesetEvent_RequestIp),
	eventfilter.Column(rulesetpb.SecRulesetEvent_RequestDomain),
	eventfilter.Column(rulesetpb.SecRulesetEvent_RequestPath),
}

type ISecRulesetEvent interface {
	// Create inserts the events, which have ids, in one statement; an
	// event whose id and matched time are stored already is skipped, so
	// inserting a batch again is a no-op.
	Create(ctx context.Context, events ...*rulesetpb.SecRulesetEvent) error
	Get(ctx context.Context, id string) (*rulesetpb.SecRulesetEvent, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) ISecRulesetEvent

	List(ctx context.Context,
		filter *eventfilter.Filter) ([]*rulesetpb.SecRulesetEvent, error)
	// Histogram counts the events per bucket, resource and behavior
	// (histogram.Point.Key).
	// The events of events only when set (see eventreader.Reader).
	Histogram(ctx context.Context, filter *histogram.Filter,
		events *eventfilter.Filter) ([]*histogram.Point, error)
	topvalues.TopValuer
}

type SecRulesetEvents struct {
	*eventreader.Reader[rulesetpb.SecRulesetEvent]
}

// New returns the repo of the sztable.AnalyticSecRulesetsEvent
// hypertable, and of its histograms.
func New(ctx context.Context,
	appConf *settingpb.Config) (ISecRulesetEvent, error) {
	reader, err := eventreader.New(ctx, appConf, _spec)
	if err != nil {
		return nil, err
	}

	return &SecRulesetEvents{Reader: reader}, nil
}

func (r *SecRulesetEvents) WithTX(tx *postgres.TxSession) ISecRulesetEvent {
	return &SecRulesetEvents{Reader: r.WithTx(tx)}
}
