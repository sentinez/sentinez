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

// Package leader elects the one edge that leads a zone.
package leader

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/sentinez/shared/szlog"
)

const _retryDelay = 5 * time.Second

var errLost = errors.New("leadership lost")

// Term is one leadership of a zone. Index grows with every new leader of
// the zone: send it with each write as a fencing token, so the store can
// reject the writes of a deposed leader that does not know it yet.
type Term struct {
	Index uint64
	Node  string
}

// Task runs while the node leads its zone. ctx is canceled as soon as the
// leadership is lost. Returning nil keeps the leadership, returning an
// error gives it up.
type Task func(ctx context.Context, term Term) error

// backend grants the leadership of a zone.
type backend interface {
	// acquire blocks until the node leads or ctx ends.
	acquire(ctx context.Context) (*lease, error)
}

// lease is a leadership granted by a backend.
type lease struct {
	term Term
	// lost is closed once the leadership is lost; nil if it never is.
	lost    <-chan struct{}
	release func() error
}

// Elector campaigns for the leadership of a zone and runs a Task while it
// holds it.
type Elector struct {
	backend    backend
	retryDelay time.Duration
	current    atomic.Pointer[Term]
}

func newElector(b backend) *Elector {
	return &Elector{backend: b, retryDelay: _retryDelay}
}

// Leader returns the term of the node if it leads its zone.
func (e *Elector) Leader() (Term, bool) {
	t := e.current.Load()
	if t == nil {
		return Term{}, false
	}

	return *t, true
}

// Run campaigns until ctx ends and runs task in every term won. A nil
// task only holds the leadership.
func (e *Elector) Run(ctx context.Context, task Task) error {
	if task == nil {
		task = func(context.Context, Term) error { return nil }
	}

	for ctx.Err() == nil {
		l, err := e.backend.acquire(ctx)
		if err != nil {
			if ctx.Err() == nil {
				szlog.Warnf("leader: campaign: %v", err)
			}

			e.wait(ctx)
			continue
		}

		if err := e.lead(ctx, l, task); err != nil {
			szlog.Warnf("leader: term %d: %v", l.term.Index, err)
			e.wait(ctx)
		}
	}

	return nil
}

// lead runs task until it fails, the leadership is lost or ctx ends,
// then gives the leadership up.
func (e *Elector) lead(ctx context.Context, l *lease, task Task) error {
	term := l.term
	szlog.Infof("leader: %s leads, term %d", term.Node, term.Index)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var taskErr error
	finished := make(chan struct{})

	e.current.Store(&term)
	go func() {
		defer close(finished)
		taskErr = task(ctx, term)
	}()

	err := hold(ctx, l, finished, func() error { return taskErr })

	// stop claiming the leadership before the task winds down
	e.current.Store(nil)
	cancel()
	<-finished

	if relErr := l.release(); relErr != nil {
		szlog.Warnf("leader: release term %d: %v", term.Index, relErr)
	}

	return err
}

// hold returns once the leadership is lost, ctx ends or the task fails;
// taskErr is read only once finished is closed.
func hold(ctx context.Context, l *lease,
	finished <-chan struct{}, taskErr func() error,
) error {
	select {
	case <-l.lost:
		return errLost
	case <-ctx.Done():
		return nil
	case <-finished:
		if err := taskErr(); err != nil {
			return err
		}
	}

	// the task is over but the node keeps leading
	select {
	case <-l.lost:
		return errLost
	case <-ctx.Done():
		return nil
	}
}

func (e *Elector) wait(ctx context.Context) {
	t := time.NewTimer(e.retryDelay)
	defer t.Stop()

	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
