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

// Package eventsrepo reads the edge events stored by the collector, and
// their histograms: the read side of the analytic.
package eventsrepo

import (
	"context"
	"fmt"

	coreruleeventsrepo "github.com/sentinez/controlplane/apps/collector/v1/repos/coreruleevents"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	httpegresseventsrepo "github.com/sentinez/controlplane/apps/collector/v1/repos/httpegressevents"
	httpingressrepo "github.com/sentinez/controlplane/apps/collector/v1/repos/httpingress"
	secruleeventsrepo "github.com/sentinez/controlplane/apps/collector/v1/repos/secruleevents"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/topvalues"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
	analyticpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	crspb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// Reader the read-only view of the repo of an event table: the collector
// repos satisfy it, the writes of the collector stay out of the analytic.
type Reader[T any] interface {
	// List the events of filter, the latest first.
	List(ctx context.Context, filter *eventfilter.Filter) ([]*T, error)
	kindReader
}

// kindReader what the analytic reads of the table of a kind of events,
// whatever its event type.
type kindReader interface {
	// Histogram counts the events per bucket, resource and key.
	Histogram(ctx context.Context,
		filter *histogram.Filter) ([]*histogram.Point, error)
	topvalues.TopValuer
}

// IngressReader the Reader of the requests received by the edge, and of
// their visitors: the client ips.
type IngressReader interface {
	Reader[httppb.IngressEvent]
	httpingressrepo.VisitorReader
}

// Events the readers of every event table.
type Events struct {
	HttpIngress IngressReader
	HttpEgress  Reader[httppb.EgressEvent]
	SecRule     Reader[secrulepb.SecRuleEvent]
	RateLimit   Reader[secrulepb.SecRuleEvent]
	CoreRule    Reader[crspb.CoreRuleEvent]
}

// New returns the Events of the tables of the collector; the tables and
// their histograms are synced on the way, as the collector does.
func New(ctx context.Context, appConf *settingpb.Config) (*Events, error) {
	var (
		events Events
		err    error
	)

	events.HttpIngress, err = httpingressrepo.New(ctx, appConf)
	if err != nil {
		return nil, fmt.Errorf("http ingress: %w", err)
	}

	events.HttpEgress, err = httpegresseventsrepo.New(ctx, appConf)
	if err != nil {
		return nil, fmt.Errorf("http egress: %w", err)
	}

	events.SecRule, err = secruleeventsrepo.NewSecRules(ctx, appConf)
	if err != nil {
		return nil, fmt.Errorf("secrule: %w", err)
	}

	events.RateLimit, err = secruleeventsrepo.NewRateLimits(ctx, appConf)
	if err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	events.CoreRule, err = coreruleeventsrepo.New(ctx, appConf)
	if err != nil {
		return nil, fmt.Errorf("core rule: %w", err)
	}

	return &events, nil
}

// Histogram the points of the histogram of kind, split by the key of its
// table: the http status, the status of the origin or the behavior.
func (e *Events) Histogram(ctx context.Context, kind analyticpb.EventKind,
	filter *histogram.Filter) ([]*histogram.Point, error) {
	reader, err := e.reader(kind)
	if err != nil {
		return nil, err
	}

	return reader.Histogram(ctx, filter)
}

// TopValues the top values of the fields of names of the events of kind
// (see topvalues.TopValuer).
func (e *Events) TopValues(ctx context.Context, kind analyticpb.EventKind,
	filter *eventfilter.Filter, names []string,
	limit uint64) (*topvalues.Result, error) {
	reader, err := e.reader(kind)
	if err != nil {
		return nil, err
	}

	return reader.TopValues(ctx, filter, names, limit)
}

// reader the reader of the table of kind.
func (e *Events) reader(kind analyticpb.EventKind) (kindReader, error) {
	switch kind {
	case analyticpb.EventKind_EVENT_KIND_HTTP_INGRESS:
		return e.HttpIngress, nil
	case analyticpb.EventKind_EVENT_KIND_HTTP_EGRESS:
		return e.HttpEgress, nil
	case analyticpb.EventKind_EVENT_KIND_SEC_RULE:
		return e.SecRule, nil
	case analyticpb.EventKind_EVENT_KIND_RATE_LIMIT:
		return e.RateLimit, nil
	case analyticpb.EventKind_EVENT_KIND_CORE_RULE:
		return e.CoreRule, nil
	default:
		return nil, szerror.StatusInvalidArgumentF(
			"invalid event kind %s", kind)
	}
}
