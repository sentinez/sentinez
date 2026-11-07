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

package szkafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/sztopic"
	"github.com/twmb/franz-go/pkg/kgo"
)

// The consumer defaults bound the memory of the records fetched but not
// handled yet to about MaxConcurrentFetches * FetchMaxBytes (48 MiB),
// however many records wait on the topics.
const (
	DefaultMaxPollRecords         = 1000
	DefaultFetchMaxBytes          = 16 << 20
	DefaultFetchMaxPartitionBytes = 1 << 20
	DefaultMaxConcurrentFetches   = 3
	DefaultFetchMaxWait           = 500 * time.Millisecond
)

const (
	_retryMin = 500 * time.Millisecond
	_retryMax = 30 * time.Second
)

var (
	errNoGroup  = errors.New("szkafka: no consumer group")
	errNoTopics = errors.New("szkafka: no topics")
)

// ConsumerConfig configures a Consumer; a zero limit takes its default.
type ConsumerConfig struct {
	// Brokers are the seed brokers, e.g. ["localhost:9094"].
	Brokers []string
	// Group is the consumer group, sharing the partitions of Topics
	// between the consumers of the group.
	Group  string
	Topics []sztopic.Topic
	// ServiceName is sent as the client id.
	ServiceName string

	// MaxPollRecords caps the records of a batch handed to the handler.
	MaxPollRecords int
	// FetchMaxBytes caps the bytes of one fetch response, compressed.
	FetchMaxBytes int32
	// FetchMaxPartitionBytes caps the bytes of a partition in a fetch.
	FetchMaxPartitionBytes int32
	// MaxConcurrentFetches caps the fetches in flight or buffered.
	MaxConcurrentFetches int
	// FetchMaxWait is how long a broker waits for records to reply.
	FetchMaxWait time.Duration
}

// BatchHandler handles a batch of records, an error retries the batch.
type BatchHandler func(ctx context.Context, records []*kgo.Record) error

// Consumer polls batches of records of a consumer group and commits the
// ones handled: delivery is at least once, a batch handled before a crash
// but not committed yet is handled again.
type Consumer struct {
	client    *kgo.Client
	maxPoll   int
	closeOnce sync.Once
}

// NewConsumer returns the consumer of cfg, it consumes once Run.
func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	switch {
	case len(cfg.Brokers) == 0:
		return nil, errNoBrokers
	case cfg.Group == "":
		return nil, errNoGroup
	case len(cfg.Topics) == 0:
		return nil, errNoTopics
	}
	cfg = cfg.withDefaults()

	client, err := kgo.NewClient(consumerOpts(cfg)...)
	if err != nil {
		return nil, fmt.Errorf("szkafka: new consumer: %w", err)
	}

	return &Consumer{client: client, maxPoll: cfg.MaxPollRecords}, nil
}

// Run hands the polled batches to handle until ctx is done or the
// consumer closed. A batch is retried, with backoff, until handled: no
// other batch is polled meanwhile, so the fetched records stay bounded.
func (c *Consumer) Run(ctx context.Context, handle BatchHandler) error {
	for {
		fetches := c.client.PollRecords(ctx, c.maxPoll)
		if fetches.IsClientClosed() || ctx.Err() != nil {
			return nil
		}
		fetches.EachError(func(topic string, part int32, err error) {
			szlog.Warnf("szkafka: fetch %s[%d]: %v", topic, part, err)
		})

		records := fetches.Records()
		if len(records) == 0 {
			continue
		}
		if !handleRetry(ctx, handle, records) {
			return nil // ctx done, the batch is not committed
		}

		c.client.MarkCommitRecords(records...)
	}
}

// Close commits the handled records, leaves the group and closes the
// client; it is safe to call more than once.
func (c *Consumer) Close() {
	c.closeOnce.Do(c.client.Close)
}

func (cfg ConsumerConfig) withDefaults() ConsumerConfig {
	if cfg.MaxPollRecords <= 0 {
		cfg.MaxPollRecords = DefaultMaxPollRecords
	}
	if cfg.FetchMaxBytes <= 0 {
		cfg.FetchMaxBytes = DefaultFetchMaxBytes
	}
	if cfg.FetchMaxPartitionBytes <= 0 {
		cfg.FetchMaxPartitionBytes = DefaultFetchMaxPartitionBytes
	}
	if cfg.MaxConcurrentFetches <= 0 {
		cfg.MaxConcurrentFetches = DefaultMaxConcurrentFetches
	}
	if cfg.FetchMaxWait <= 0 {
		cfg.FetchMaxWait = DefaultFetchMaxWait
	}

	return cfg
}

func consumerOpts(cfg ConsumerConfig) []kgo.Opt {
	return []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ClientID(cfg.ServiceName),
		kgo.ConsumerGroup(cfg.Group),
		kgo.ConsumeTopics(topicNames(cfg.Topics)...),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		// commit only the records marked handled; the default revoke
		// hook commits them before the partitions move.
		kgo.AutoCommitMarks(),
		kgo.FetchMaxBytes(cfg.FetchMaxBytes),
		kgo.FetchMaxPartitionBytes(cfg.FetchMaxPartitionBytes),
		kgo.MaxConcurrentFetches(cfg.MaxConcurrentFetches),
		kgo.FetchMaxWait(cfg.FetchMaxWait),
		// a broker response may not exceed a fetch, plus its framing.
		kgo.BrokerMaxReadBytes(2 * cfg.FetchMaxBytes),
	}
}

// handleRetry calls handle until it handles records, then reports true,
// or ctx is done, then reports false.
func handleRetry(ctx context.Context, handle BatchHandler,
	records []*kgo.Record) bool {
	backoff := _retryMin
	for {
		err := handle(ctx, records)
		if err == nil {
			return true
		}
		szlog.Errorf("szkafka: handle %d records, retry in %s: %v",
			len(records), backoff, err)

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return false
		case <-timer.C:
		}
		backoff = min(2*backoff, _retryMax)
	}
}

// topicNames the names of topics, as kgo takes them.
func topicNames(topics []sztopic.Topic) []string {
	names := make([]string, len(topics))
	for i, topic := range topics {
		names[i] = topic.String()
	}

	return names
}
