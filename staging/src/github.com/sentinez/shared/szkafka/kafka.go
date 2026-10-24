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

// Package szkafka produces the events szlog loggers log straight to Kafka,
// one record per event, without going through OTLP and a collector.
package szkafka

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/szlog"
	"github.com/sentinez/shared/sztopic"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

// Headers of the records a Producer produces. The record value is the
// event itself, proto-encoded; the record timestamp is when it was logged.
const (
	// HeaderKind holds the LogType name, e.g. "LOG_TYPE_HTTP".
	HeaderKind = "sz-kind"
	// HeaderLevel holds the szlog.Level name, e.g. "info".
	HeaderLevel = "sz-level"
	// HeaderMessage holds the log message.
	HeaderMessage = "sz-msg"
	// HeaderService holds Config.ServiceName.
	HeaderService = "sz-service"
)

const _linger = 5 * time.Millisecond

var (
	_ szlog.EventExporter = (*Producer)(nil)

	errNoBrokers = errors.New("szkafka: no brokers")

	_levelHeaders = [...][]byte{
		szlog.LevelDebug:   []byte(szlog.LevelDebug.String()),
		szlog.LevelInfo:    []byte(szlog.LevelInfo.String()),
		szlog.LevelWarning: []byte(szlog.LevelWarning.String()),
		szlog.LevelError:   []byte(szlog.LevelError.String()),
		szlog.LevelFatal:   []byte(szlog.LevelFatal.String()),
	}
)

// Config configures a Producer.
type Config struct {
	// Brokers are the seed brokers, e.g. ["localhost:9094"].
	Brokers []string
	// Topics maps each log type to its topic; nil uses DefaultTopics.
	// Events of an unmapped type are dropped.
	Topics map[typepb.LogType]string
	// ServiceName is sent as the HeaderService header and the client id.
	ServiceName string
}

// ConfigOf returns the config producing to the comma-separated brokers as
// serviceName, with the default topics.
func ConfigOf(brokers, serviceName string) Config {
	var seeds []string
	for b := range strings.SplitSeq(brokers, ",") {
		if b = strings.TrimSpace(b); b != "" {
			seeds = append(seeds, b)
		}
	}

	return Config{Brokers: seeds, ServiceName: serviceName}
}

// DefaultTopics returns the topic of each edge traffic log type.
func DefaultTopics() map[typepb.LogType]string {
	return map[typepb.LogType]string{
		typepb.LogType_LOG_TYPE_HTTP:       sztopic.EventHTTPIngress,
		typepb.LogType_LOG_TYPE_WAF:        sztopic.EventHTTPRulesets,
		typepb.LogType_LOG_TYPE_SEC_RULE:   sztopic.EventHTTPSecRule,
		typepb.LogType_LOG_TYPE_RATE_LIMIT: sztopic.EventHTTPRateLimit,
		typepb.LogType_LOG_TYPE_CDN_RULE:   sztopic.EventHTTPCDNRule,
	}
}

// Producer is a szlog.EventExporter producing each event to the topic of
// its log type. Exporting never blocks: when the buffer is full the event
// is dropped and counted (see Dropped).
type Producer struct {
	client  *kgo.Client
	routes  map[typepb.LogType]route
	service []byte
	dropped atomic.Uint64
}

// route is where the events of one log type go, with their kind header.
type route struct {
	topic string
	kind  []byte
}

// NewProducer builds a Producer; it connects lazily, on the first event.
// The caller owns it and must Shutdown it to flush pending events.
func NewProducer(cfg Config) (*Producer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errNoBrokers
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ClientID(cfg.ServiceName),
		kgo.RequiredAcks(kgo.LeaderAck()),
		kgo.DisableIdempotentWrite(),
		kgo.ProducerBatchCompression(kgo.ZstdCompression()),
		kgo.ProducerLinger(_linger),
	)
	if err != nil {
		return nil, fmt.Errorf("szkafka: new client: %w", err)
	}

	return &Producer{
		client:  client,
		routes:  routesOf(cfg.Topics),
		service: []byte(cfg.ServiceName),
	}, nil
}

// Export encodes event, then produces it asynchronously.
func (p *Producer) Export(kind typepb.LogType, level szlog.Level,
	msg string, event proto.Message) {
	r, ok := p.routes[kind]
	if !ok {
		p.dropped.Add(1)
		return
	}
	value, err := marshal(event)
	if err != nil {
		p.dropped.Add(1)
		return
	}

	p.client.TryProduce(context.Background(), &kgo.Record{
		Topic: r.topic,
		Value: value,
		Headers: []kgo.RecordHeader{
			{Key: HeaderKind, Value: r.kind},
			{Key: HeaderLevel, Value: levelHeader(level)},
			{Key: HeaderMessage, Value: []byte(msg)},
			{Key: HeaderService, Value: p.service},
		},
	}, p.onProduced)
}

// Dropped returns how many events could not be produced so far.
func (p *Producer) Dropped() uint64 {
	return p.dropped.Load()
}

// ForceFlush waits until every buffered event is produced or ctx is done.
func (p *Producer) ForceFlush(ctx context.Context) error {
	if p == nil {
		return nil
	}

	return p.client.Flush(ctx)
}

// Shutdown flushes the buffered events, then closes the client.
func (p *Producer) Shutdown(ctx context.Context) error {
	if p == nil {
		return nil
	}
	err := p.client.Flush(ctx)
	p.client.Close()

	return err
}

func (p *Producer) onProduced(_ *kgo.Record, err error) {
	if err != nil {
		p.dropped.Add(1)
	}
}

func routesOf(topics map[typepb.LogType]string) map[typepb.LogType]route {
	if topics == nil {
		topics = DefaultTopics()
	}

	routes := make(map[typepb.LogType]route, len(topics))
	for kind, topic := range topics {
		routes[kind] = route{topic: topic, kind: []byte(kind.String())}
	}

	return routes
}

func levelHeader(level szlog.Level) []byte {
	if level < 0 || int(level) >= len(_levelHeaders) {
		return []byte(level.String())
	}

	return _levelHeaders[level]
}

func marshal(event proto.Message) ([]byte, error) {
	if m, ok := event.(interface{ MarshalVT() ([]byte, error) }); ok {
		return m.MarshalVT()
	}

	return proto.Marshal(event)
}
