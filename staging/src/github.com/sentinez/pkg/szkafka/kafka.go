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
	"maps"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/sztopic"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

// Headers of the records a Producer produces. The record value is the
// event itself, proto-encoded; the record timestamp is when it was logged.
const (
	// HeaderKind holds the LogType name, e.g. "LOG_TYPE_HTTP_INGRESS".
	HeaderKind = "sz-kind"
	// HeaderLevel holds the szlog.Level name, e.g. "info".
	HeaderLevel = "sz-level"
	// HeaderMessage holds the log message.
	HeaderMessage = "sz-msg"
	// HeaderService holds Config.ServiceName.
	HeaderService = "sz-service"
)

// DefaultTopicPartitions the partitions of the topics a producer creates,
// also the most collector replicas sharing a topic.
const DefaultTopicPartitions = 6

const (
	_linger = 5 * time.Millisecond
	// _createTimeout bounds the creation of the missing topics.
	_createTimeout = 10 * time.Second
)

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
	Topics map[typepb.LogType]sztopic.Topic
	// ServiceName is sent as the HeaderService header and the client id.
	ServiceName string
	// Partitions of the topics it produces to NewProducer creates when
	// missing, zero takes DefaultTopicPartitions; an existing topic is left as
	// is. A failed creation is only logged: the producer still starts,
	// its records dropped until the topics exist.
	Partitions int32
	// ReplicationFactor of the topics created, zero takes the broker
	// default.
	ReplicationFactor int16
}

// ConfigOption sets an optional field of the Config of ConfigOf.
type ConfigOption interface {
	apply(cfg *Config)
}

type topicsOption map[typepb.LogType]sztopic.Topic

func (o topicsOption) apply(cfg *Config) {
	cfg.Topics = o
}

// WithTopics produces each log type to its topic of topics, see
// Config.Topics.
func WithTopics(topics map[typepb.LogType]sztopic.Topic) ConfigOption {
	return topicsOption(maps.Clone(topics))
}

type partitionsOption int32

func (o partitionsOption) apply(cfg *Config) {
	cfg.Partitions = int32(o)
}

// WithTopicPartitions creates the missing topics with n partitions, see
// Config.Partitions.
func WithTopicPartitions(n int32) ConfigOption {
	return partitionsOption(n)
}

type replicationOption int16

func (o replicationOption) apply(cfg *Config) {
	cfg.ReplicationFactor = int16(o)
}

// WithReplicationFactor replicates the topics created n times, see
// Config.ReplicationFactor.
func WithReplicationFactor(n int16) ConfigOption {
	return replicationOption(n)
}

// ConfigOf returns the config producing to the comma-separated brokers as
// serviceName, with the default topics, set by opts.
func ConfigOf(brokers, serviceName string, opts ...ConfigOption) Config {
	cfg := Config{Brokers: Brokers(brokers), ServiceName: serviceName}
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	return cfg
}

// Brokers splits the comma-separated brokers, e.g. "a:9092, b:9092".
func Brokers(brokers string) []string {
	var seeds []string
	for b := range strings.SplitSeq(brokers, ",") {
		if b = strings.TrimSpace(b); b != "" {
			seeds = append(seeds, b)
		}
	}

	return seeds
}

// DefaultTopics returns the topic of each edge traffic log type.
func DefaultTopics() map[typepb.LogType]sztopic.Topic {
	return map[typepb.LogType]sztopic.Topic{
		typepb.LogType_LOG_TYPE_HTTP_INGRESS:  sztopic.EventHTTPIngress,
		typepb.LogType_LOG_TYPE_HTTP_EGRESS:   sztopic.EventHTTPEgress,
		typepb.LogType_LOG_TYPE_CORE_RULESETS: sztopic.EventHTTPRulesets,
		typepb.LogType_LOG_TYPE_SEC_RULE:      sztopic.EventHTTPSecRule,
		typepb.LogType_LOG_TYPE_RATE_LIMIT:    sztopic.EventHTTPRateLimit,
		typepb.LogType_LOG_TYPE_CDN_RULE:      sztopic.EventHTTPCDNRule,
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

// NewProducer builds a Producer, after creating its missing topics.
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

	p := &Producer{
		client:  client,
		routes:  routesOf(cfg.Topics),
		service: []byte(cfg.ServiceName),
	}
	if err := p.createTopics(cfg); err != nil {
		szlog.Warnf("%v", err)
	}

	return p, nil
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

func routesOf(
	topics map[typepb.LogType]sztopic.Topic) map[typepb.LogType]route {

	if topics == nil {
		topics = DefaultTopics()
	}

	routes := make(map[typepb.LogType]route, len(topics))
	for kind, topic := range topics {
		routes[kind] = route{
			topic: topic.String(),
			kind:  []byte(kind.String()),
		}
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

// createTopics creates the topics of p missing on the brokers, with the
// partitions and replication factor of cfg.
func (p *Producer) createTopics(cfg Config) error {
	partitions := cfg.Partitions
	if partitions <= 0 {
		partitions = DefaultTopicPartitions
	}
	replication := cfg.ReplicationFactor
	if replication <= 0 {
		replication = -1 // the broker default
	}

	ctx, cancel := context.WithTimeout(context.Background(), _createTimeout)
	defer cancel()

	resps, err := kadm.NewClient(p.client).CreateTopics(ctx,
		partitions, replication, nil, p.topics()...)
	if err != nil {
		return fmt.Errorf("szkafka: create topics: %w", err)
	}
	for _, r := range resps.Sorted() {
		if r.Err != nil && !errors.Is(r.Err, kerr.TopicAlreadyExists) {
			return fmt.Errorf("szkafka: create topic %s: %w", r.Topic, r.Err)
		}
	}

	return nil
}

// topics the distinct topics p produces to.
func (p *Producer) topics() []string {
	topics := make(map[string]struct{}, len(p.routes))
	for _, r := range p.routes {
		topics[r.topic] = struct{}{}
	}

	return slices.Collect(maps.Keys(topics))
}
