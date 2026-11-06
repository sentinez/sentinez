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

// Package ingest consumes the edge traffic events of kafka and inserts
// them, in batches, in the analytic event tables.
package ingest

import (
	"context"

	"github.com/sentinez/shared/szkafka"
	"github.com/sentinez/shared/sztopic"
	"github.com/twmb/franz-go/pkg/kgo"
	"golang.org/x/sync/errgroup"
)

// Ingest hands the batches polled by its consumer to the sink of each
// record's topic.
type Ingest struct {
	consumer *szkafka.Consumer
	sinks    map[sztopic.Topic]Sink
}

// New returns the ingest of the topics of sinks; a nil consumer disables
// it.
func New(consumer *szkafka.Consumer,
	sinks map[sztopic.Topic]Sink) *Ingest {

	return &Ingest{consumer: consumer, sinks: sinks}
}

// Run inserts the polled batches until ctx is done.
func (i *Ingest) Run(ctx context.Context) error {
	if i.consumer == nil {
		return nil
	}

	return i.consumer.Run(ctx, i.Handle)
}

// Close commits the inserted records and closes the consumer.
func (i *Ingest) Close() {
	if i.consumer != nil {
		i.consumer.Close()
	}
}

// Handle inserts records with the sinks of their topics, each sink, on
// its own table, concurrently. An error retries the whole batch: the
// rows inserted already are sent again but not stored twice.
func (i *Ingest) Handle(ctx context.Context, records []*kgo.Record) error {
	byTopic := make(map[sztopic.Topic][]*kgo.Record, len(i.sinks))
	for _, r := range records {
		topic := sztopic.Topic(r.Topic)
		byTopic[topic] = append(byTopic[topic], r)
	}

	g, gctx := errgroup.WithContext(ctx)
	for topic, batch := range byTopic {
		if s, ok := i.sinks[topic]; ok {
			g.Go(func() error { return s.Insert(gctx, batch) })
		}
	}

	return g.Wait()
}
