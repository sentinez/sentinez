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

package centralsync

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sentinez/pkg/szsign"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _wait = 5 * time.Second

var _key = bytes.Repeat([]byte{7}, szsign.MinKeyBytes)

// session what one subscribe + one stream of the fake source give.
type session struct {
	rows    []*centraldatapb.Change
	version uint64
	epoch   string
	// msgs the test publishes envelopes on.
	msgs chan *redis.Message
}

func newSession(epoch string, version uint64, ids ...string) *session {
	s := &session{version: version, epoch: epoch,
		msgs: make(chan *redis.Message, 8)}
	for _, id := range ids {
		s.rows = append(s.rows, upsert(id, 0, epoch))
	}

	return s
}

func upsert(id string, version uint64, epoch string) *centraldatapb.Change {
	return &centraldatapb.Change{
		Table:   "security.secrules",
		Op:      centraldatapb.Change_OP_UPSERT,
		Id:      id,
		Version: version,
		Epoch:   epoch,
	}
}

// fakeSource serves the sessions in order, and records the calls.
type fakeSource struct {
	mu       sync.Mutex
	sessions chan *session
	cur      *session
	calls    []string
}

func newFakeSource(sessions ...*session) *fakeSource {
	src := &fakeSource{sessions: make(chan *session, len(sessions))}
	for _, s := range sessions {
		src.sessions <- s
	}

	return src
}

func (s *fakeSource) record(call string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, call)
}

func (s *fakeSource) Calls() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]string(nil), s.calls...)
}

func (s *fakeSource) Subscribe(ctx context.Context,
) (<-chan *redis.Message, io.Closer, error) {
	select {
	case cur := <-s.sessions:
		s.mu.Lock()
		s.cur = cur
		s.mu.Unlock()
		s.record("subscribe")

		return cur.msgs, io.NopCloser(nil), nil
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
}

func (s *fakeSource) Sync(context.Context) (Stream, error) {
	s.record("sync")

	s.mu.Lock()
	cur := s.cur
	s.mu.Unlock()

	var resp []*centraldatapb.SyncResponse
	for _, row := range cur.rows {
		resp = append(resp, &centraldatapb.SyncResponse{
			Kind: &centraldatapb.SyncResponse_Row{Row: row}})
	}

	resp = append(resp, &centraldatapb.SyncResponse{
		Kind: &centraldatapb.SyncResponse_Done{Done: &centraldatapb.Snapshot{
			Version: cur.version, Epoch: cur.epoch}}})

	return &fakeStream{resp: resp}, nil
}

type fakeStream struct{ resp []*centraldatapb.SyncResponse }

func (s *fakeStream) Recv() (*centraldatapb.SyncResponse, error) {
	if len(s.resp) == 0 {
		return nil, io.EOF
	}

	r := s.resp[0]
	s.resp = s.resp[1:]

	return r, nil
}

// recordingSink records what is replaced and applied.
type recordingSink struct {
	events chan string
}

func (s *recordingSink) Replace(c []*centraldatapb.Change) error {
	ids := ""
	for _, ch := range c {
		ids += ch.GetId() + ","
	}

	s.events <- "replace:" + ids

	return nil
}

func (s *recordingSink) Apply(c *centraldatapb.Change) error {
	s.events <- "apply:" + c.GetId()

	return nil
}

func (s *recordingSink) next(t *testing.T) string {
	t.Helper()

	select {
	case e := <-s.events:
		return e
	case <-time.After(_wait):
		require.FailNow(t, "the sink got nothing")

		return ""
	}
}

func start(t *testing.T, src *fakeSource, resync time.Duration,
) *recordingSink {
	t.Helper()

	sink := &recordingSink{events: make(chan string, 16)}
	f := New(src, sink, _key, resync)

	ctx, cancel := context.WithTimeout(context.Background(), _wait)
	defer cancel()
	require.NoError(t, f.Start(ctx))
	t.Cleanup(f.Close)

	return sink
}

func publish(t *testing.T, s *session, key []byte, c *centraldatapb.Change) {
	t.Helper()

	data, err := szsign.Seal(key, c)
	require.NoError(t, err)
	s.msgs <- &redis.Message{Payload: string(data)}
}

func TestLoadsThenSubscribedBeforeSync(t *testing.T) {
	s := newSession("e1", 5, "a", "b")
	src := newFakeSource(s)

	sink := start(t, src, time.Hour)

	assert.Equal(t, "replace:a,b,", sink.next(t))
	assert.Equal(t, []string{"subscribe", "sync"}, src.Calls())
}

func TestFollowsInOrderAndSkipsWhatTheSnapshotHolds(t *testing.T) {
	s := newSession("e1", 5, "a")
	sink := start(t, newFakeSource(s), time.Hour)
	assert.Equal(t, "replace:a,", sink.next(t))

	publish(t, s, _key, upsert("old", 5, "e1"))
	publish(t, s, _key, upsert("x", 6, "e1"))
	publish(t, s, _key, upsert("y", 7, "e1"))

	assert.Equal(t, "apply:x", sink.next(t))
	assert.Equal(t, "apply:y", sink.next(t))
}

func TestDropsAForgedChange(t *testing.T) {
	s := newSession("e1", 5)
	sink := start(t, newFakeSource(s), time.Hour)
	sink.next(t)

	publish(t, s, bytes.Repeat([]byte{9}, szsign.MinKeyBytes),
		upsert("forged", 6, "e1"))
	s.msgs <- &redis.Message{Payload: "garbage"}
	publish(t, s, _key, upsert("real", 6, "e1"))

	assert.Equal(t, "apply:real", sink.next(t))
}

func TestGapLoadsAgain(t *testing.T) {
	first := newSession("e1", 5, "a")
	second := newSession("e1", 9, "a", "b")
	src := newFakeSource(first, second)
	sink := start(t, src, time.Hour)
	sink.next(t)

	publish(t, first, _key, upsert("lost-6", 7, "e1"))

	assert.Equal(t, "replace:a,b,", sink.next(t))
}

func TestNewEpochLoadsAgain(t *testing.T) {
	first := newSession("e1", 5, "a")
	second := newSession("e2", 0, "b")
	sink := start(t, newFakeSource(first, second), time.Hour)
	sink.next(t)

	publish(t, first, _key, upsert("x", 1, "e2"))

	assert.Equal(t, "replace:b,", sink.next(t))
}

func TestResetLoadsAgain(t *testing.T) {
	first := newSession("e1", 5, "a")
	second := newSession("e2", 0, "b")
	sink := start(t, newFakeSource(first, second), time.Hour)
	sink.next(t)

	publish(t, first, _key, &centraldatapb.Change{Version: 0, Epoch: "e2"})

	assert.Equal(t, "replace:b,", sink.next(t))
}

func TestPeriodicResync(t *testing.T) {
	first := newSession("e1", 5, "a")
	second := newSession("e1", 5, "a")
	sink := start(t, newFakeSource(first, second), 50*time.Millisecond)
	sink.next(t)

	assert.Equal(t, "replace:a,", sink.next(t))
}

// An unreachable centraldata does not hold the edge: Start gives up when
// its context ends, and the Follower goes on trying.
func TestStartGivesUpButKeepsTrying(t *testing.T) {
	src := &failingSource{}
	f := New(src, &recordingSink{events: make(chan string, 4)}, _key, 0)
	ctx, cancel := context.WithTimeout(context.Background(),
		100*time.Millisecond)
	defer cancel()

	err := f.Start(ctx)
	f.Close()

	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.GreaterOrEqual(t, src.attempts(), 1)
}

type failingSource struct {
	mu sync.Mutex
	n  int
}

func (s *failingSource) attempts() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.n
}

func (s *failingSource) Subscribe(context.Context,
) (<-chan *redis.Message, io.Closer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.n++

	return nil, nil, errors.New("centraldata is down")
}

func (s *failingSource) Sync(context.Context) (Stream, error) {
	return nil, errors.New("unreachable")
}
