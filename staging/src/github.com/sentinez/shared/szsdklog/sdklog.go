// Copyright 2025 Duc-Hung Ho.
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

// Package szsdklog picks where the edge traffic logs of the func chains are
// exported to.
package szsdklog

import (
	"context"
	"errors"
	"fmt"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/szkafka"
	"github.com/sentinez/shared/szlog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// Exporter is where the edge traffic logs go: Kafka when
// SENZ_KAFKA_BROKERS is set, else OTLP when SENZ_OTLP_EXTERNAL_ENDPOINT is
// set, else nowhere. A nil *Exporter exports nothing.
type Exporter struct {
	kafka *szkafka.Producer
	otlp  *sdklog.LoggerProvider
}

// New returns the exporter configured by conf, nil when neither setting is
// set.
func New(conf *settingpb.Config) (*Exporter, error) {
	service := conf.GetMeta().GetServiceKey()

	if brokers := conf.Get(settingpb.Senz_SENZ_KAFKA_BROKERS); brokers != "" {
		p, err := szkafka.NewProducer(szkafka.ConfigOf(brokers, service))
		if err != nil {
			return nil, fmt.Errorf("edge log: %w", err)
		}
		return &Exporter{kafka: p}, nil
	}

	endpoint := conf.Get(settingpb.Senz_SENZ_OTLP_EXTERNAL_ENDPOINT)
	if endpoint == "" {
		return nil, nil
	}
	p, err := szlog.NewOTLPProvider(context.Background(),
		szlog.OTLPConfigOf(endpoint, service))
	if err != nil {
		return nil, fmt.Errorf("edge log: %w", err)
	}

	return &Exporter{otlp: p}, nil
}

// Option returns the szlog option sending a logger's records to e.
func (e *Exporter) Option() szlog.Option {
	switch {
	case e == nil:
		return szlog.WithoutOTLP()
	case e.kafka != nil:
		return szlog.WithEventExporter(e.kafka)
	case e.otlp != nil:
		return szlog.WithOTLPProvider(e.otlp)
	default:
		return szlog.WithoutOTLP()
	}
}

// Shutdown flushes the pending records, then closes the exporter.
func (e *Exporter) Shutdown(ctx context.Context) error {
	if e == nil {
		return nil
	}

	var err error
	if e.otlp != nil {
		err = e.otlp.Shutdown(ctx)
	}

	return errors.Join(err, e.kafka.Shutdown(ctx))
}
