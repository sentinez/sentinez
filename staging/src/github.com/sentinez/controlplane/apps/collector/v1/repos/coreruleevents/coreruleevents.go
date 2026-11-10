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

	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventreader"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/topvalues"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
	crspb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var _ ICoreRuleEvent = (*CoreRuleEvents)(nil)

// _spec the reader of the events: their histogram is split by behavior.
var _spec = eventreader.Spec[crspb.CoreRuleEvent]{
	Hypertable: schema.EventHypertable(
		schema.AnalyticCoreRulesetsEvent(), crspb.CoreRuleEvent_MatchedTime),
	Key: crspb.CoreRuleEvent_Behavior,
	Columns: eventfilter.Columns{
		RequestId:   crspb.CoreRuleEvent_RequestId,
		ResourceId:  crspb.CoreRuleEvent_ResourceId,
		MatchedTime: crspb.CoreRuleEvent_MatchedTime,
		Fields:      _fields,
	},
	Select:   crspb.CoreRuleEventColumns(),
	Scan:     crspb.ScanCoreRuleEvent[dbx.Rows],
	ScanOne:  crspb.ScanOneCoreRuleEvent[dbx.Row],
	Values:   crspb.CoreRuleEventValues,
	IdColumn: crspb.CoreRuleEvent_Id,
}

// _fields the fields of the events filters match and top values count,
// in this order.
var _fields = []eventfilter.Field{
	eventfilter.TypedColumn(crspb.CoreRuleEvent_MatchedRuleIds, dbx.IntArr),
	eventfilter.TypedColumn(crspb.CoreRuleEvent_Severities, dbx.TextArr),
	eventfilter.TypedColumn(crspb.CoreRuleEvent_Behavior, dbx.Int4),
	eventfilter.TypedColumn(crspb.CoreRuleEvent_TotalScore, dbx.Int4),
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
	// The events of events only when set (see eventreader.Reader).
	Histogram(ctx context.Context, filter *histogram.Filter,
		events *eventfilter.Filter) ([]*histogram.Point, error)
	topvalues.TopValuer
}

type CoreRuleEvents struct {
	*eventreader.Reader[crspb.CoreRuleEvent]
}

// New returns the repo of the sztable.AnalyticCoreRulesetsEvent
// hypertable, and of its histograms.
func New(ctx context.Context,
	appConf *settingpb.Config) (ICoreRuleEvent, error) {
	reader, err := eventreader.New(ctx, appConf, _spec)
	if err != nil {
		return nil, err
	}

	return &CoreRuleEvents{Reader: reader}, nil
}

func (r *CoreRuleEvents) WithTX(tx *postgres.TxSession) ICoreRuleEvent {
	return &CoreRuleEvents{Reader: r.WithTx(tx)}
}
