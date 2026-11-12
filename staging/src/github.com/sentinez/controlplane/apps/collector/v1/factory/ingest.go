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

package collectorfac

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/sentinez/controlplane/apps/collector/v1/ingest"
	coreruleeventsrepo "github.com/sentinez/controlplane/apps/collector/v1/repos/coreruleevents"
	httpegresseventsrepo "github.com/sentinez/controlplane/apps/collector/v1/repos/httpegressevents"
	httpingresseventrepo "github.com/sentinez/controlplane/apps/collector/v1/repos/httpingressevent"
	secruleeventsrepo "github.com/sentinez/controlplane/apps/collector/v1/repos/secruleevents"
	"github.com/sentinez/pkg/szkafka"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/sztopic"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// _setupTimeout bounds the sync of the event tables.
const _setupTimeout = 30 * time.Second

// NewDefaultIngest returns the ingest of the edge events of the
// SENZ_KAFKA_BROKERS, a disabled one when it is unset.
func NewDefaultIngest(conf *settingpb.Config) (*ingest.Ingest, error) {
	brokers := conf.Get(settingpb.Senz_SENZ_KAFKA_BROKERS)
	if brokers == "" {
		szlog.Warnf("collectorfac: %s unset, ingest disabled",
			settingpb.Senz_SENZ_KAFKA_BROKERS)
		return ingest.New(nil, nil), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), _setupTimeout)
	defer cancel()

	sinks, err := newSinks(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("ingest sinks: %w", err)
	}

	service := conf.GetMeta().GetServiceKey()
	consumer, err := szkafka.NewConsumer(szkafka.ConsumerConfig{
		Brokers:     szkafka.Brokers(brokers),
		Group:       service,
		Topics:      slices.Collect(maps.Keys(sinks)),
		ServiceName: service,
	})
	if err != nil {
		return nil, fmt.Errorf("ingest consumer: %w", err)
	}

	return ingest.New(consumer, sinks), nil
}

// newSinks the sink of each topic of an edge event table.
func newSinks(ctx context.Context,
	conf *settingpb.Config) (map[sztopic.Topic]ingest.Sink, error) {
	secRules, err := secruleeventsrepo.NewSecRules(ctx, conf)
	if err != nil {
		return nil, err
	}
	rateLimits, err := secruleeventsrepo.NewRateLimits(ctx, conf)
	if err != nil {
		return nil, err
	}
	coreRules, err := coreruleeventsrepo.New(ctx, conf)
	if err != nil {
		return nil, err
	}
	ingress, err := httpingresseventrepo.New(ctx, conf)
	if err != nil {
		return nil, err
	}
	egress, err := httpegresseventsrepo.New(ctx, conf)
	if err != nil {
		return nil, err
	}

	return map[sztopic.Topic]ingest.Sink{
		sztopic.EventHTTPSecRule: ingest.NewSecRuleEventSink(
			secRules.Create),
		sztopic.EventHTTPRateLimit: ingest.NewSecRuleEventSink(
			rateLimits.Create),
		sztopic.EventHTTPRulesets: ingest.NewCoreRuleEventSink(
			coreRules.Create),
		sztopic.EventHTTPIngress: ingest.NewHttpIngressEventSink(
			ingress.Create),
		sztopic.EventHTTPEgress: ingest.NewHttpEgressEventSink(
			egress.Create),
	}, nil
}
