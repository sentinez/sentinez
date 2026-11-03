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

package leader

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/sentinez/sentinez/api/client/consul"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeStore records the writes it is given and fails them with errs, in
// order, then succeeds.
type fakeStore struct {
	mu   sync.Mutex
	puts []fakePut
	errs []error
}

type fakePut struct {
	key   string
	token uint64
	value any
}

func (s *fakeStore) PutFenced(
	_ context.Context, key string, token uint64, value any,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.puts = append(s.puts, fakePut{key: key, token: token, value: value})

	if len(s.errs) == 0 {
		return nil
	}

	err := s.errs[0]
	s.errs = s.errs[1:]

	return err
}

func (s *fakeStore) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.puts)
}

func newZoneElector(store *fakeStore) *Elector {
	e := newElector(&fakeBackend{})
	e.zone = newZoneStore(store, "sgn")
	e.zone.interval = time.Millisecond

	return e
}

// runTask runs task on term until stop is called; stop returns its error.
func runTask(task Task, term Term) (stop func() error) {
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go func() { errCh <- task(ctx, term) }()

	return func() error {
		cancel()
		return <-errCh
	}
}

func TestPublishStatusWritesEveryInterval(t *testing.T) {
	store := &fakeStore{}
	task := newZoneElector(store).PublishStatus()

	stop := runTask(task, Term{Index: 7, Node: "edge-01"})
	require.Eventually(t, func() bool { return store.count() >= 2 },
		_waitFor, time.Millisecond)
	require.NoError(t, stop())

	got := store.puts[0]
	assert.Equal(t, "sentinez/dmz/edge/zone/sgn/status", got.key)
	assert.Equal(t, uint64(7), got.token)
	require.IsType(t, &Status{}, got.value)
	status := got.value.(*Status)
	assert.Equal(t, "sgn", status.Zone)
	assert.Equal(t, "edge-01", status.Node)
	assert.Equal(t, uint64(7), status.Term)
}

func TestPublishStatusFencedEndsTerm(t *testing.T) {
	store := &fakeStore{errs: []error{consul.ErrFenced}}
	task := newZoneElector(store).PublishStatus()

	// the first write is fenced: the task returns before the deadline
	ctx, cancel := context.WithTimeout(context.Background(), _waitFor)
	defer cancel()

	require.ErrorIs(t, task(ctx, Term{Index: 3}), consul.ErrFenced)
}

func TestPublishStatusRetriesOtherErrors(t *testing.T) {
	store := &fakeStore{errs: []error{errors.New("consul down")}}
	task := newZoneElector(store).PublishStatus()

	stop := runTask(task, Term{Index: 3})
	require.Eventually(t, func() bool { return store.count() >= 2 },
		_waitFor, time.Millisecond)

	require.NoError(t, stop())
}

func TestPublishStatusStandalone(t *testing.T) {
	e := newElector(&standalone{node: "edge-01"})

	assert.Nil(t, e.PublishStatus())
}
