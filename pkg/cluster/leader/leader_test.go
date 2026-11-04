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
	"sync/atomic"
	"testing"
	"time"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _waitFor = time.Second

// fakeLease is a lease that records its release; close lostCh to lose it.
type fakeLease struct {
	lease

	lostCh   chan struct{}
	released atomic.Bool
}

func newFakeLease(index uint64) *fakeLease {
	l := &fakeLease{lostCh: make(chan struct{})}
	l.lease = lease{
		term: Term{Index: index, Node: "edge-01"},
		lost: l.lostCh,
		release: func() error {
			l.released.Store(true)
			return nil
		},
	}

	return l
}

// fakeBackend grants the leases sent on leases, one per campaign.
type fakeBackend struct {
	leases chan *fakeLease
}

func (b *fakeBackend) acquire(ctx context.Context) (*lease, error) {
	select {
	case l := <-b.leases:
		return &l.lease, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// startElector runs an Elector on leases and returns the terms its task
// started, and a stop func that waits for Run to return.
func startElector(
	t *testing.T, task Task, leases ...*fakeLease,
) (*Elector, <-chan Term, func()) {
	t.Helper()

	backend := &fakeBackend{leases: make(chan *fakeLease, len(leases))}
	for _, l := range leases {
		backend.leases <- l
	}

	e := newElector(backend)
	e.retryDelay = time.Millisecond

	started := make(chan Term, len(leases))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- e.Run(ctx, func(ctx context.Context, term Term) error {
			started <- term
			return task(ctx, term)
		})
	}()

	return e, started, func() {
		cancel()
		require.NoError(t, <-done)
	}
}

func waitTerm(t *testing.T, started <-chan Term) Term {
	t.Helper()

	select {
	case term := <-started:
		return term
	case <-time.After(_waitFor):
		require.FailNow(t, "no term started")
		return Term{}
	}
}

func TestElectorLeadsUntilStopped(t *testing.T) {
	first := newFakeLease(7)
	task := func(ctx context.Context, _ Term) error {
		<-ctx.Done()
		return ctx.Err()
	}

	e, started, stop := startElector(t, task, first)

	assert.Equal(t, first.term, waitTerm(t, started))

	term, ok := e.Leader()
	assert.True(t, ok)
	assert.Equal(t, first.term, term)

	stop()

	_, ok = e.Leader()
	assert.False(t, ok)
	assert.True(t, first.released.Load())
}

func TestElectorStepsDownWhenLost(t *testing.T) {
	first, second := newFakeLease(1), newFakeLease(2)
	task := func(ctx context.Context, _ Term) error {
		<-ctx.Done()
		return nil
	}

	e, started, stop := startElector(t, task, first, second)
	defer stop()

	assert.Equal(t, first.term, waitTerm(t, started))
	close(first.lostCh)

	// the second term starts only once the first task has returned
	assert.Equal(t, second.term, waitTerm(t, started))
	assert.True(t, first.released.Load())

	term, ok := e.Leader()
	assert.True(t, ok)
	assert.Equal(t, second.term, term)
}

func TestElectorStepsDownWhenTaskFails(t *testing.T) {
	first, second := newFakeLease(1), newFakeLease(2)
	errTask := errors.New("task failed")
	task := func(ctx context.Context, term Term) error {
		if term.Index == first.term.Index {
			return errTask
		}

		<-ctx.Done()
		return nil
	}

	_, started, stop := startElector(t, task, first, second)
	defer stop()

	assert.Equal(t, first.term, waitTerm(t, started))
	assert.Equal(t, second.term, waitTerm(t, started))
	assert.True(t, first.released.Load())
}

func TestElectorKeepsLeadingAfterTaskDone(t *testing.T) {
	first := newFakeLease(1)
	task := func(context.Context, Term) error { return nil }

	e, started, stop := startElector(t, task, first)
	defer stop()

	assert.Equal(t, first.term, waitTerm(t, started))
	assert.Never(t, func() bool {
		_, ok := e.Leader()
		return !ok
	}, 50*time.Millisecond, time.Millisecond)
	assert.False(t, first.released.Load())
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		giveEnv map[string]string
		wantErr error
	}{
		{name: "standalone"},
		{
			name: "clustered without consul",
			giveEnv: map[string]string{
				"SENZ_MEMBERSHIP_ADDRESS": "edge-02:7777",
				"SENZ_DISCOVERY_ADDRESS":  "0.0.0.0:7777",
			},
			wantErr: errNoConsul,
		},
		{
			name:    "consul without zone",
			giveEnv: map[string]string{"SENZ_CONSUL_URI": "consul:8500"},
			wantErr: errNoZone,
		},
		{
			name: "consul",
			giveEnv: map[string]string{
				"SENZ_CONSUL_URI": "consul:8500", "SENZ_ZONE": "SGN",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(&settingpb.Config{Env: tt.giveEnv})
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
