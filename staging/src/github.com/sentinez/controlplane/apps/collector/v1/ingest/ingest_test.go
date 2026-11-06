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

package ingest

import (
	"context"
	"errors"
	"testing"

	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/sentinez/shared/szprefix"
	"github.com/sentinez/shared/sztopic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

// fakeCreate records copies of the batches of create, the events going
// back to the pool after it, failing with err.
type fakeCreate struct {
	batches [][]*secrulepb.SecRuleEvent
	err     error
}

func (f *fakeCreate) create(_ context.Context,
	events ...*secrulepb.SecRuleEvent) error {
	if f.err != nil {
		return f.err
	}
	f.batches = append(f.batches, clones(events))

	return nil
}

func clones[M proto.Message](msgs []M) []M {
	out := make([]M, len(msgs))
	for i, m := range msgs {
		out[i], _ = proto.Clone(m).(M)
	}

	return out
}

func ruleRecord(t *testing.T, topic string, offset int64,
	event *secrulepb.SecRuleEvent) *kgo.Record {
	t.Helper()
	value, err := event.MarshalVT()
	require.NoError(t, err)

	return &kgo.Record{Topic: topic, Partition: 2, Offset: offset,
		Value: value}
}

func TestSecRuleEventSinkInsert(t *testing.T) {
	repo := &fakeCreate{}
	s := NewSecRuleEventSink(repo.create)

	records := []*kgo.Record{
		ruleRecord(t, "t", 7, &secrulepb.SecRuleEvent{Id: "srevt_a"}),
		{Topic: "t", Value: []byte{0xff}}, // malformed, dropped
		ruleRecord(t, "t", 9, &secrulepb.SecRuleEvent{RequestPath: "/x"}),
	}
	require.NoError(t, s.Insert(t.Context(), records))

	require.Len(t, repo.batches, 1)
	got := repo.batches[0]
	require.Len(t, got, 2)
	assert.Equal(t, "srevt_a", got[0].GetId())
	assert.Equal(t, szprefix.SecRuleEvent.Key()+"t-2-9", got[1].GetId())
	assert.Equal(t, "/x", got[1].GetRequestPath())
}

func TestHttpIngressEventSinkInsert(t *testing.T) {
	var got []*httppb.IngressEvent
	s := NewHttpIngressEventSink(func(_ context.Context,
		events ...*httppb.IngressEvent) error {
		got = append(got, clones(events)...)
		return nil
	})

	withId, err := (&httppb.IngressEvent{Id: "req_a", Status: 200}).MarshalVT()
	require.NoError(t, err)
	noId, err := (&httppb.IngressEvent{Host: "a.test"}).MarshalVT()
	require.NoError(t, err)

	records := []*kgo.Record{
		{Topic: "t", Partition: 1, Offset: 3, Value: withId},
		{Topic: "t", Value: []byte{0xff}}, // malformed, dropped
		{Topic: "t", Partition: 1, Offset: 5, Value: noId},
	}
	require.NoError(t, s.Insert(t.Context(), records))

	require.Len(t, got, 2)
	assert.Equal(t, "req_a", got[0].GetId())
	assert.Equal(t, int32(200), got[0].GetStatus())
	assert.Equal(t, szprefix.Request.Key()+"t-1-5", got[1].GetId())
	assert.Equal(t, "a.test", got[1].GetHost())
}

func TestHttpEgressEventSinkInsert(t *testing.T) {
	var got []*httppb.EgressEvent
	s := NewHttpEgressEventSink(func(_ context.Context,
		events ...*httppb.EgressEvent) error {
		got = append(got, clones(events)...)
		return nil
	})

	withId, err := (&httppb.EgressEvent{Id: "req_a", Status: 502}).
		MarshalVT()
	require.NoError(t, err)
	noId, err := (&httppb.EgressEvent{Upstream: "http://o"}).MarshalVT()
	require.NoError(t, err)

	records := []*kgo.Record{
		{Topic: "t", Partition: 1, Offset: 3, Value: withId},
		{Topic: "t", Value: []byte{0xff}}, // malformed, dropped
		{Topic: "t", Partition: 1, Offset: 5, Value: noId},
	}
	require.NoError(t, s.Insert(t.Context(), records))

	require.Len(t, got, 2)
	assert.Equal(t, "req_a", got[0].GetId())
	assert.Equal(t, int32(502), got[0].GetStatus())
	assert.Equal(t, szprefix.Request.Key()+"t-1-5", got[1].GetId())
	assert.Equal(t, "http://o", got[1].GetUpstream())
}

func TestSinkReleasesEvents(t *testing.T) {
	var kept []*secrulepb.SecRuleEvent
	s := NewSecRuleEventSink(func(_ context.Context,
		events ...*secrulepb.SecRuleEvent) error {
		kept = append(kept, events...)
		return nil
	})

	records := []*kgo.Record{
		ruleRecord(t, "t", 1, &secrulepb.SecRuleEvent{RequestPath: "/x"}),
	}
	require.NoError(t, s.Insert(t.Context(), records))

	require.Len(t, kept, 1)
	assert.Empty(t, kept[0].GetRequestPath(), "released once inserted")
}

func TestSecRuleEventSinkChunks(t *testing.T) {
	repo := &fakeCreate{}
	s := NewSecRuleEventSink(repo.create)

	records := make([]*kgo.Record, _insertRows+1)
	for i := range records {
		records[i] = ruleRecord(t, "t", int64(i), &secrulepb.SecRuleEvent{})
	}
	require.NoError(t, s.Insert(t.Context(), records))

	require.Len(t, repo.batches, 2)
	assert.Len(t, repo.batches[0], _insertRows)
	assert.Len(t, repo.batches[1], 1)
}

func TestHandleRoutesTopics(t *testing.T) {
	secRules, rateLimits := &fakeCreate{}, &fakeCreate{}
	in := New(nil, map[sztopic.Topic]Sink{
		"secrule":   NewSecRuleEventSink(secRules.create),
		"ratelimit": NewSecRuleEventSink(rateLimits.create),
	})

	records := []*kgo.Record{
		ruleRecord(t, "secrule", 1, &secrulepb.SecRuleEvent{}),
		ruleRecord(t, "ratelimit", 2, &secrulepb.SecRuleEvent{}),
		ruleRecord(t, "secrule", 3, &secrulepb.SecRuleEvent{}),
		ruleRecord(t, "unknown", 4, &secrulepb.SecRuleEvent{}),
	}
	require.NoError(t, in.Handle(t.Context(), records))

	require.Len(t, secRules.batches, 1)
	assert.Len(t, secRules.batches[0], 2)
	require.Len(t, rateLimits.batches, 1)
	assert.Len(t, rateLimits.batches[0], 1)
}

func TestHandleInsertError(t *testing.T) {
	errInsert := errors.New("db down")
	in := New(nil, map[sztopic.Topic]Sink{
		"secrule": NewSecRuleEventSink(
			(&fakeCreate{err: errInsert}).create),
	})

	records := []*kgo.Record{
		ruleRecord(t, "secrule", 1, &secrulepb.SecRuleEvent{}),
	}
	require.ErrorIs(t, in.Handle(t.Context(), records), errInsert)
}

func TestRunDisabled(t *testing.T) {
	in := New(nil, nil)

	require.NoError(t, in.Run(t.Context()))
	in.Close()
}
