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

package centraldatasvc

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	memrepo "github.com/sentinez/controlplane/apps/centraldata/v1/repos/memory"
	pgrepo "github.com/sentinez/controlplane/apps/centraldata/v1/repos/postgres"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/sztable"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

const (
	_table    = sztable.SecuritySecRules
	_relation = "dev_sentinez_security_secrules"
	_wait     = 5 * time.Second
)

// fakeTable a synced table whose rows the test sets.
type fakeTable struct {
	mu   sync.Mutex
	rows map[string]*securitypb.SecRule
}

func newFakeTable(rows ...*securitypb.SecRule) *fakeTable {
	t := &fakeTable{rows: make(map[string]*securitypb.SecRule)}
	for _, row := range rows {
		t.rows[row.GetId()] = row
	}

	return t
}

func (t *fakeTable) set(row *securitypb.SecRule) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rows[row.GetId()] = row
}

func (t *fakeTable) drop(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.rows, id)
}

func (t *fakeTable) Name() sztable.Table { return _table }
func (t *fakeTable) Relation() string    { return _relation }

func (t *fakeTable) List(context.Context) ([]proto.Message, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var list []proto.Message
	for _, row := range t.rows {
		list = append(list, proto.Clone(row))
	}

	return list, nil
}

func (t *fakeTable) Get(_ context.Context, id string) (proto.Message, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	row, ok := t.rows[id]
	if !ok {
		return nil, szerror.ErrNotFound
	}

	return proto.Clone(row), nil
}

// fakeSub delivers the results the test sends.
type fakeSub struct {
	results chan error
	changes chan postgres.Change
}

func newFakeSub() *fakeSub {
	return &fakeSub{
		results: make(chan error),
		changes: make(chan postgres.Change),
	}
}

func (s *fakeSub) Next(ctx context.Context) (postgres.Change, error) {
	select {
	case c := <-s.changes:
		return c, nil
	case err := <-s.results:
		return postgres.Change{}, err
	case <-ctx.Done():
		return postgres.Change{}, ctx.Err()
	}
}

func (s *fakeSub) Close(context.Context) error { return nil }

// fakePublisher hands the published changes to the test.
type fakePublisher struct {
	changes chan *centraldatapb.Change
}

func (p *fakePublisher) Publish(ctx context.Context,
	c *centraldatapb.Change) error {

	select {
	case p.changes <- c:
	case <-ctx.Done():
	}

	return nil
}

// harness a started Syncer on fakes.
type harness struct {
	table  *fakeTable
	memory *memrepo.Memory
	pub    *fakePublisher
	syncer *Syncer
	subs   chan *fakeSub
	cur    *fakeSub
}

func newHarness(t *testing.T, rows ...*securitypb.SecRule) *harness {
	t.Helper()

	memory, err := memrepo.New()
	require.NoError(t, err)

	h := &harness{
		table:  newFakeTable(rows...),
		memory: memory,
		pub:    &fakePublisher{changes: make(chan *centraldatapb.Change)},
		subs:   make(chan *fakeSub, 4),
	}
	listen := func(context.Context) (Subscription, error) {
		sub := newFakeSub()
		h.subs <- sub

		return sub, nil
	}

	h.syncer = NewSyncer(memory, []pgrepo.Table{h.table}, listen, h.pub,
		time.Hour)
	require.NoError(t, h.syncer.Start(context.Background()))
	t.Cleanup(h.syncer.Close)

	h.cur = <-h.subs

	return h
}

func (h *harness) next(t *testing.T) *centraldatapb.Change {
	t.Helper()

	select {
	case c := <-h.pub.changes:
		return c
	case <-time.After(_wait):
		require.FailNow(t, "no change was published")

		return nil
	}
}

func (h *harness) notify(t *testing.T, op, id string) {
	t.Helper()

	select {
	case h.cur.changes <- postgres.Change{Table: _relation, Op: op, Id: id}:
	case <-time.After(_wait):
		require.FailNow(t, "the listener did not take the notification")
	}
}

func rule(id, name string) *securitypb.SecRule {
	return &securitypb.SecRule{Id: id, Name: name}
}

func held(t *testing.T, h *harness, id string) *securitypb.SecRule {
	t.Helper()

	row, err := memrepo.Get[securitypb.SecRule](h.memory, _table, id)
	if errors.Is(err, szerror.ErrNotFound) {
		return nil
	}

	require.NoError(t, err)

	return row
}

func TestStartLoadsWithoutPublishing(t *testing.T) {
	h := newHarness(t, rule("a", "one"), rule("b", "two"))

	assert.Equal(t, "one", held(t, h, "a").GetName())
	assert.Equal(t, "two", held(t, h, "b").GetName())

	select {
	case c := <-h.pub.changes:
		require.FailNow(t, "published on start", "%v", c)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestNotifyUpserts(t *testing.T) {
	h := newHarness(t)
	h.table.set(rule("a", "one"))

	h.notify(t, "INSERT", "a")
	got := h.next(t)

	assert.Equal(t, "one", held(t, h, "a").GetName())
	assert.Equal(t, _table.String(), got.GetTable())
	assert.Equal(t, centraldatapb.Change_OP_UPSERT, got.GetOp())
	assert.Equal(t, "a", got.GetId())
	assert.Equal(t, h.memory.Version(), got.GetVersion())
	assert.Equal(t, h.memory.Epoch(), got.GetEpoch())

	var sent securitypb.SecRule
	require.NoError(t, got.GetRow().UnmarshalTo(&sent))
	assert.Equal(t, "one", sent.GetName())
}

func TestNotifyDeletes(t *testing.T) {
	h := newHarness(t, rule("a", "one"))
	h.table.drop("a")

	h.notify(t, "DELETE", "a")
	got := h.next(t)

	assert.Nil(t, held(t, h, "a"))
	assert.Equal(t, centraldatapb.Change_OP_DELETE, got.GetOp())
	assert.Equal(t, "a", got.GetId())
	assert.Nil(t, got.GetRow())
}

// A row inserted and deleted before the listener reads it again is not in
// the memory, and there is nothing to publish.
func TestNotifyOfGoneRow(t *testing.T) {
	h := newHarness(t)

	h.notify(t, "INSERT", "ghost")
	h.table.set(rule("a", "one"))
	h.notify(t, "INSERT", "a")
	got := h.next(t)

	assert.Equal(t, "a", got.GetId(), "ghost must not be published")
	assert.Equal(t, uint64(1), got.GetVersion())
}

func TestNotifyOfOtherTable(t *testing.T) {
	h := newHarness(t)
	h.table.set(rule("a", "one"))

	select {
	case h.cur.changes <- postgres.Change{Table: "other", Id: "a"}:
	case <-time.After(_wait):
		require.FailNow(t, "the listener did not take the notification")
	}
	h.notify(t, "INSERT", "a")

	assert.Equal(t, uint64(1), h.next(t).GetVersion())
}

// The reconcile catches what a dropped notification missed.
func TestReconcile(t *testing.T) {
	h := newHarness(t, rule("a", "one"), rule("b", "two"), rule("c", "three"))
	h.table.set(rule("a", "changed"))
	h.table.drop("b")
	h.table.set(rule("d", "new"))

	done := make(chan error, 1)
	go func() { done <- h.syncer.sync(context.Background(), true) }()

	got := map[string]centraldatapb.Change_Op{}
	for range 3 {
		c := h.next(t)
		got[c.GetId()] = c.GetOp()
	}
	require.NoError(t, <-done)

	assert.Equal(t, map[string]centraldatapb.Change_Op{
		"a": centraldatapb.Change_OP_UPSERT,
		"b": centraldatapb.Change_OP_DELETE,
		"d": centraldatapb.Change_OP_UPSERT,
	}, got)
	assert.Equal(t, "changed", held(t, h, "a").GetName())
	assert.Equal(t, "three", held(t, h, "c").GetName())
	assert.Nil(t, held(t, h, "b"))
}

// A lost connection may have lost changes: the epoch changes and a reset
// is published, so the edges load the tables again.
func TestReconnectStartsNewEpoch(t *testing.T) {
	h := newHarness(t, rule("a", "one"))
	before := h.memory.Epoch()
	h.table.set(rule("a", "changed while away"))

	h.cur.results <- errors.New("connection reset")
	reset := h.next(t)
	h.cur = <-h.subs

	assert.Empty(t, reset.GetTable())
	assert.Equal(t, centraldatapb.Change_OP_UNSPECIFIED, reset.GetOp())
	assert.NotEqual(t, before, reset.GetEpoch())
	assert.Equal(t, h.memory.Epoch(), reset.GetEpoch())
	assert.Equal(t, "changed while away", held(t, h, "a").GetName())

	// the new subscription is served.
	h.table.set(rule("b", "two"))
	h.notify(t, "INSERT", "b")
	assert.Equal(t, "b", h.next(t).GetId())
}
