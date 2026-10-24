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
	"io"
	"testing"
	"time"

	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/szlog"
	"github.com/sentinez/shared/sztopic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kfake"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

func newCluster(t *testing.T) []string {
	t.Helper()
	cluster, err := kfake.NewCluster(
		kfake.SeedTopics(1, sztopic.EventHTTPSecRule))
	require.NoError(t, err)
	t.Cleanup(cluster.Close)

	return cluster.ListenAddrs()
}

// consumeOne reads the first record of topic.
func consumeOne(t *testing.T, brokers []string, topic string) *kgo.Record {
	t.Helper()
	client, err := kgo.NewClient(kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics(topic))
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fetches := client.PollRecords(ctx, 1)
	require.NoError(t, fetches.Err())
	records := fetches.Records()
	require.Len(t, records, 1)

	return records[0]
}

func headersOf(r *kgo.Record) map[string]string {
	m := make(map[string]string, len(r.Headers))
	for _, h := range r.Headers {
		m[h.Key] = string(h.Value)
	}

	return m
}

func TestProducerThroughLogger(t *testing.T) {
	brokers := newCluster(t)
	p, err := NewProducer(Config{Brokers: brokers, ServiceName: "edge"})
	require.NoError(t, err)

	want := &secrulepb.RuleEvent{
		Id:                "sr_1",
		MatchedSecruleIds: []string{"a", "b"},
		TenantId:          "tn_1",
		MatchedTime:       1761200000000,
	}
	l := szlog.NewLogCloser("test", typepb.LogType_LOG_TYPE_SEC_RULE,
		szlog.LevelInfo, szlog.WithEventExporter(p))
	l.Info("secrule: matched", want, io.NopCloser(nil))
	require.NoError(t, p.Shutdown(context.Background()))

	rec := consumeOne(t, brokers, sztopic.EventHTTPSecRule)
	got := &secrulepb.RuleEvent{}
	require.NoError(t, got.UnmarshalVT(rec.Value))
	assert.True(t, proto.Equal(want, got), "got %v", got)
	assert.Equal(t, map[string]string{
		HeaderKind:    "LOG_TYPE_SEC_RULE",
		HeaderLevel:   "info",
		HeaderMessage: "secrule: matched",
		HeaderService: "edge",
	}, headersOf(rec))
	assert.Zero(t, p.Dropped())
}

func TestProducerDropsUnmappedKind(t *testing.T) {
	p, err := NewProducer(Config{
		Brokers: newCluster(t),
		Topics:  map[typepb.LogType]string{},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })

	p.Export(typepb.LogType_LOG_TYPE_HTTP, szlog.LevelInfo, "x",
		&secrulepb.RuleEvent{})

	assert.Equal(t, uint64(1), p.Dropped())
}

func TestNewProducerNoBrokers(t *testing.T) {
	_, err := NewProducer(Config{})

	require.ErrorIs(t, err, errNoBrokers)
}

func TestConfigOf(t *testing.T) {
	tests := []struct {
		name        string
		giveBrokers string
		wantBrokers []string
	}{
		{name: "empty", giveBrokers: ""},
		{
			name:        "one",
			giveBrokers: "localhost:9094",
			wantBrokers: []string{"localhost:9094"},
		},
		{
			name:        "many with spaces",
			giveBrokers: "a:9092, b:9092 ,,",
			wantBrokers: []string{"a:9092", "b:9092"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ConfigOf(tt.giveBrokers, "svc")

			assert.Equal(t, tt.wantBrokers, cfg.Brokers)
			assert.Equal(t, "svc", cfg.ServiceName)
		})
	}
}
