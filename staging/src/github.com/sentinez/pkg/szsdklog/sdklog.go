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

	"github.com/sentinez/pkg/szkafka"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/sztls"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// Exporter is where the edge traffic logs go: Kafka when SZ_KAFKA_URI is
// set, else OTLP when SZ_OTLP_EXTERNAL_URI is set, else nowhere. A nil
// *Exporter exports nothing.
type Exporter struct {
	kafka *szkafka.Producer
	otlp  *sdklog.LoggerProvider
}

// New returns the exporter configured by conf, nil when neither setting is
// set; opts configure the kafka producer, e.g. the replication factor of
// the topics it creates.
func New(conf *settingpb.Config,
	opts ...szkafka.ConfigOption) (*Exporter, error) {
	service := conf.GetMeta().GetServiceKey()

	if uri := conf.Get(settingpb.Sz_SZ_KAFKA_URI); uri != "" {
		return newKafka(uri, conf, service, opts)
	}

	uri := conf.Get(settingpb.Sz_SZ_OTLP_EXTERNAL_URI)
	if uri == "" {
		return nil, nil
	}

	otlpConf, err := szlog.OTLPConfigOf(uri, service)
	if err != nil {
		return nil, fmt.Errorf("edge log: %w", err)
	}

	p, err := szlog.NewOTLPProvider(context.Background(), otlpConf)
	if err != nil {
		return nil, fmt.Errorf("edge log: %w", err)
	}

	return &Exporter{otlp: p}, nil
}

// newKafka returns the exporter producing to the Kafka of uri, with the
// TLS files of conf for kafkas.
func newKafka(uri string, conf *settingpb.Config, service string,
	opts []szkafka.ConfigOption) (*Exporter, error) {
	ep, err := szkafka.ParseURI(uri, sztls.FilesOf(conf))
	if err != nil {
		return nil, fmt.Errorf("edge log: %w", err)
	}

	p, err := szkafka.NewProducer(szkafka.ConfigOf(ep, service, opts...))
	if err != nil {
		return nil, fmt.Errorf("edge log: %w", err)
	}

	return &Exporter{kafka: p}, nil
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
