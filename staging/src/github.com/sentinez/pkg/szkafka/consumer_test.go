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
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/sentinez/pkg/sztopic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kfake"
	"github.com/twmb/franz-go/pkg/kgo"
)

const _testTopic = "sentinez.test.events"

func newConsumerCluster(t *testing.T) []string {
	t.Helper()
	cluster, err := kfake.NewCluster(kfake.SeedTopics(1, _testTopic))
	require.NoError(t, err)
	t.Cleanup(cluster.Close)

	return cluster.ListenAddrs()
}

func produce(t *testing.T, brokers []string, n int) {
	t.Helper()
	client, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	require.NoError(t, err)
	defer client.Close()

	for i := range n {
		r := &kgo.Record{Topic: _testTopic, Value: []byte(strconv.Itoa(i))}
		require.NoError(t, client.ProduceSync(t.Context(), r).FirstErr())
	}
}

func newTestConsumer(t *testing.T, brokers []string) *Consumer {
	t.Helper()
	c, err := NewConsumer(ConsumerConfig{
		Brokers:        brokers,
		Group:          "test",
		Topics:         []sztopic.Topic{_testTopic},
		MaxPollRecords: 2,
		FetchMaxWait:   50 * time.Millisecond,
	})
	require.NoError(t, err)

	return c
}

// collector records the values handed to its handle, failing the first
// failures calls.
type collector struct {
	mu       sync.Mutex
	values   []string
	batches  []int
	failures int
	done     chan struct{}
	want     int
}

func (c *collector) handle(_ context.Context, records []*kgo.Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.failures > 0 {
		c.failures--
		return errors.New("insert failed")
	}
	c.batches = append(c.batches, len(records))
	for _, r := range records {
		c.values = append(c.values, string(r.Value))
	}
	if len(c.values) == c.want {
		close(c.done)
	}

	return nil
}

// runUntil runs c until col got every value, then closes c.
func runUntil(t *testing.T, c *Consumer, col *collector) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	errc := make(chan error, 1)
	go func() { errc <- c.Run(ctx, col.handle) }()

	select {
	case <-col.done:
	case <-ctx.Done():
		t.Fatal("timed out")
	}
	cancel()
	require.NoError(t, <-errc)
	c.Close()
}

func TestConsumerBatches(t *testing.T) {
	brokers := newConsumerCluster(t)
	produce(t, brokers, 5)

	col := &collector{done: make(chan struct{}), want: 5}
	runUntil(t, newTestConsumer(t, brokers), col)

	assert.Equal(t, []string{"0", "1", "2", "3", "4"}, col.values)
	for _, n := range col.batches {
		assert.LessOrEqual(t, n, 2)
	}
}

func TestConsumerRetriesBatch(t *testing.T) {
	brokers := newConsumerCluster(t)
	produce(t, brokers, 2)

	col := &collector{done: make(chan struct{}), want: 2, failures: 1}
	runUntil(t, newTestConsumer(t, brokers), col)

	assert.Equal(t, []string{"0", "1"}, col.values)
}

func TestConsumerCommitsOnClose(t *testing.T) {
	brokers := newConsumerCluster(t)
	produce(t, brokers, 2)

	first := &collector{done: make(chan struct{}), want: 2}
	runUntil(t, newTestConsumer(t, brokers), first)

	produce(t, brokers, 1)
	second := &collector{done: make(chan struct{}), want: 1}
	runUntil(t, newTestConsumer(t, brokers), second)

	assert.Equal(t, []string{"0"}, second.values)
}

func TestNewConsumerConfigErrors(t *testing.T) {
	tests := []struct {
		name    string
		give    ConsumerConfig
		wantErr error
	}{
		{
			name:    "no brokers",
			give:    ConsumerConfig{Group: "g", Topics: []sztopic.Topic{"t"}},
			wantErr: errNoBrokers,
		},
		{
			name: "no group",
			give: ConsumerConfig{
				Brokers: []string{"b"}, Topics: []sztopic.Topic{"t"}},
			wantErr: errNoGroup,
		},
		{
			name:    "no topics",
			give:    ConsumerConfig{Brokers: []string{"b"}, Group: "g"},
			wantErr: errNoTopics,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewConsumer(tt.give)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
