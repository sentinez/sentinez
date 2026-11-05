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

	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/sentinez/shared/szprefix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
)

// fakeCreate records the batches of create, failing with err.
type fakeCreate struct {
	batches [][]*secrulepb.RuleEvent
	err     error
}

func (f *fakeCreate) create(_ context.Context,
	events ...*secrulepb.RuleEvent) error {
	if f.err != nil {
		return f.err
	}
	f.batches = append(f.batches, events)

	return nil
}

func ruleRecord(t *testing.T, topic string, offset int64,
	event *secrulepb.RuleEvent) *kgo.Record {
	t.Helper()
	value, err := event.MarshalVT()
	require.NoError(t, err)

	return &kgo.Record{Topic: topic, Partition: 2, Offset: offset,
		Value: value}
}

func TestRuleEventSinkInsert(t *testing.T) {
	repo := &fakeCreate{}
	s := NewRuleEventSink(repo.create)

	records := []*kgo.Record{
		ruleRecord(t, "t", 7, &secrulepb.RuleEvent{Id: "srevt_a"}),
		{Topic: "t", Value: []byte{0xff}}, // malformed, dropped
		ruleRecord(t, "t", 9, &secrulepb.RuleEvent{RequestPath: "/x"}),
	}
	require.NoError(t, s.Insert(t.Context(), records))

	require.Len(t, repo.batches, 1)
	got := repo.batches[0]
	require.Len(t, got, 2)
	assert.Equal(t, "srevt_a", got[0].GetId())
	assert.Equal(t, szprefix.SecRuleEvent.Key()+"t-2-9", got[1].GetId())
	assert.Equal(t, "/x", got[1].GetRequestPath())
}

func TestRuleEventSinkChunks(t *testing.T) {
	repo := &fakeCreate{}
	s := NewRuleEventSink(repo.create)

	records := make([]*kgo.Record, _insertRows+1)
	for i := range records {
		records[i] = ruleRecord(t, "t", int64(i), &secrulepb.RuleEvent{})
	}
	require.NoError(t, s.Insert(t.Context(), records))

	require.Len(t, repo.batches, 2)
	assert.Len(t, repo.batches[0], _insertRows)
	assert.Len(t, repo.batches[1], 1)
}

func TestHandleRoutesTopics(t *testing.T) {
	secRules, rateLimits := &fakeCreate{}, &fakeCreate{}
	in := New(nil, map[string]Sink{
		"secrule":   NewRuleEventSink(secRules.create),
		"ratelimit": NewRuleEventSink(rateLimits.create),
	})

	records := []*kgo.Record{
		ruleRecord(t, "secrule", 1, &secrulepb.RuleEvent{}),
		ruleRecord(t, "ratelimit", 2, &secrulepb.RuleEvent{}),
		ruleRecord(t, "secrule", 3, &secrulepb.RuleEvent{}),
		ruleRecord(t, "unknown", 4, &secrulepb.RuleEvent{}),
	}
	require.NoError(t, in.Handle(t.Context(), records))

	require.Len(t, secRules.batches, 1)
	assert.Len(t, secRules.batches[0], 2)
	require.Len(t, rateLimits.batches, 1)
	assert.Len(t, rateLimits.batches[0], 1)
}

func TestHandleInsertError(t *testing.T) {
	errInsert := errors.New("db down")
	in := New(nil, map[string]Sink{
		"secrule": NewRuleEventSink(
			(&fakeCreate{err: errInsert}).create),
	})

	records := []*kgo.Record{
		ruleRecord(t, "secrule", 1, &secrulepb.RuleEvent{}),
	}
	require.ErrorIs(t, in.Handle(t.Context(), records), errInsert)
}

func TestRunDisabled(t *testing.T) {
	in := New(nil, nil)

	require.NoError(t, in.Run(t.Context()))
	in.Close()
}
