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
	"fmt"
	"strings"
	"sync"
	"time"

	memrepo "github.com/sentinez/controlplane/apps/centraldata/v1/repos/memory"
	pgrepo "github.com/sentinez/controlplane/apps/centraldata/v1/repos/postgres"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/sztable"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	"google.golang.org/protobuf/proto"
)

const (
	// DefaultReconcile how often the tables are compared with the memory,
	// to catch the notifications postgres dropped.
	DefaultReconcile = time.Minute
	_retryMin        = time.Second
	_retryMax        = 30 * time.Second
)

// Subscription the changes of the database.
type Subscription interface {
	Next(ctx context.Context) (postgres.Change, error)
	Close(ctx context.Context) error
}

// Publisher sends the changes to the edges.
type Publisher interface {
	Publish(ctx context.Context, change *centraldatapb.Change) error
}

// Syncer keeps the memory equal to the synced tables of the database, and
// publishes every change it applies.
type Syncer struct {
	memory    *memrepo.Memory
	tables    []pgrepo.Table
	listen    func(context.Context) (Subscription, error)
	publisher Publisher
	reconcile time.Duration

	// mu orders a write to the memory with its publication, so the
	// versions reach the edges in order.
	mu     sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewSyncer returns a Syncer of tables. listen opens a subscription to the
// notifications of the tables; reconcile is the period of the comparison
// with the database ([DefaultReconcile] when 0).
func NewSyncer(memory *memrepo.Memory, tables []pgrepo.Table,
	listen func(context.Context) (Subscription, error),
	publisher Publisher, reconcile time.Duration) *Syncer {

	if reconcile <= 0 {
		reconcile = DefaultReconcile
	}

	return &Syncer{
		memory:    memory,
		tables:    tables,
		listen:    listen,
		publisher: publisher,
		reconcile: reconcile,
	}
}

// Start listens first, then loads the tables, so nothing that changes
// while loading is missed; it returns once the memory holds them, and keeps
// them up to date until Close.
func (s *Syncer) Start(ctx context.Context) error {
	sub, err := s.listen(ctx)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	if err := s.sync(ctx, false /* publish */); err != nil {
		_ = sub.Close(ctx)

		return err
	}

	runCtx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.wg.Add(2)
	go s.listenLoop(runCtx, sub)
	go s.reconcileLoop(runCtx)

	return nil
}

// Close stops the loops and waits for them.
func (s *Syncer) Close() {
	if s.cancel != nil {
		s.cancel()
	}

	s.wg.Wait()
}

func (s *Syncer) listenLoop(ctx context.Context, sub Subscription) {
	defer s.wg.Done()
	defer func() { _ = sub.Close(context.Background()) }()

	for ctx.Err() == nil {
		change, err := sub.Next(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			szlog.Warnf("centraldata: listen: %v", err)
			sub = s.reconnect(ctx, sub)

			continue
		}

		s.handle(ctx, change)
	}
}

// reconnect replaces the lost subscription, retrying with backoff until
// ctx ends, and starts a new epoch: the changes of the gap are unknown, so
// the edges must load the tables again.
func (s *Syncer) reconnect(ctx context.Context,
	lost Subscription) Subscription {

	_ = lost.Close(context.Background())

	for wait := _retryMin; ctx.Err() == nil; wait = min(wait*2, _retryMax) {
		sub, err := s.listen(ctx)
		if err == nil {
			if err = s.resync(ctx); err == nil {
				return sub
			}

			_ = sub.Close(context.Background())
		}

		szlog.Warnf("centraldata: reconnect: %v", err)
		sleep(ctx, wait)
	}

	return closedSubscription{}
}

func (s *Syncer) resync(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	epoch := s.memory.Reset()
	if err := s.syncLocked(ctx, false /* publish */); err != nil {
		return err
	}

	s.publish(ctx, &centraldatapb.Change{
		Version: s.memory.Version(),
		Epoch:   epoch,
	})

	return nil
}

func (s *Syncer) reconcileLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.reconcile)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.sync(ctx, true /* publish */); err != nil {
				szlog.Warnf("centraldata: reconcile: %v", err)
			}
		}
	}
}

// sync makes the memory equal to the database, publishing what changed
// when publish.
func (s *Syncer) sync(ctx context.Context, publish bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.syncLocked(ctx, publish)
}

func (s *Syncer) syncLocked(ctx context.Context, publish bool) error {
	for _, table := range s.tables {
		rows, err := table.List(ctx)
		if err != nil {
			return err
		}

		if err := s.diff(ctx, table, rows, publish); err != nil {
			return err
		}
	}

	return nil
}

// diff upserts the rows the memory lacks or holds different, and deletes
// those the database no longer has.
func (s *Syncer) diff(ctx context.Context, table pgrepo.Table,
	rows []proto.Message, publish bool) error {

	held, err := s.held(table.Name())
	if err != nil {
		return err
	}

	var changed []proto.Message
	for _, row := range rows {
		id := idOf(row)
		old, ok := held[id]
		delete(held, id)

		if !ok || !proto.Equal(old, row) {
			changed = append(changed, row)
		}
	}

	if err := s.upsert(ctx, table.Name(), changed, publish); err != nil {
		return err
	}

	for id := range held {
		if err := s.remove(ctx, table.Name(), id, publish); err != nil {
			return err
		}
	}

	return nil
}

// held the rows of the memory by id.
func (s *Syncer) held(table sztable.Table) (map[string]proto.Message, error) {
	held := make(map[string]proto.Message)

	_, _, err := s.memory.Snapshot([]sztable.Table{table},
		func(_ sztable.Table, row any) error {
			msg, ok := row.(proto.Message)
			if !ok {
				return fmt.Errorf("%s holds %T", table, row)
			}

			held[idOf(msg)] = msg

			return nil
		})

	return held, err
}

// handle applies a notification: the row is read again, as it may have
// changed since.
func (s *Syncer) handle(ctx context.Context, change postgres.Change) {
	table, ok := s.table(change.Table)
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	row, err := table.Get(ctx, change.Id)
	switch {
	case errors.Is(err, szerror.ErrNotFound):
		err = s.remove(ctx, table.Name(), change.Id, true /* publish */)
	case err == nil:
		err = s.upsert(ctx, table.Name(), []proto.Message{row},
			true /* publish */)
	}

	if err != nil {
		szlog.Errorf("centraldata: apply %s %s: %v",
			table.Name(), change.Id, err)
	}
}

func (s *Syncer) table(relation string) (pgrepo.Table, bool) {
	for _, table := range s.tables {
		if strings.EqualFold(table.Relation(), relation) {
			return table, true
		}
	}

	return nil, false
}

// upsert writes rows to the memory: in one write, or one by one when each
// is published with its own version.
func (s *Syncer) upsert(ctx context.Context, table sztable.Table,
	rows []proto.Message, publish bool) error {

	if !publish {
		batch := make([]any, len(rows))
		for i, row := range rows {
			batch[i] = row
		}

		_, err := s.memory.UpsertAt(table, batch...)

		return err
	}

	for _, row := range rows {
		version, err := s.memory.UpsertAt(table, row)
		if err != nil {
			return err
		}

		change, err := upsertChange(table, row)
		if err != nil {
			return err
		}

		change.Version, change.Epoch = version, s.memory.Epoch()
		s.publish(ctx, change)
	}

	return nil
}

// remove deletes the row of id, which the memory may not hold.
func (s *Syncer) remove(ctx context.Context, table sztable.Table,
	id string, publish bool) error {

	version, err := s.memory.DeleteAt(table, id)
	if errors.Is(err, szerror.ErrNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	if publish {
		s.publish(ctx, &centraldatapb.Change{
			Table:   table.String(),
			Op:      centraldatapb.Change_OP_DELETE,
			Id:      id,
			Version: version,
			Epoch:   s.memory.Epoch(),
		})
	}

	return nil
}

// publish never fails the sync: an edge that misses a change sees the gap
// of versions and loads the tables again.
func (s *Syncer) publish(ctx context.Context, change *centraldatapb.Change) {
	if err := s.publisher.Publish(ctx, change); err != nil {
		szlog.Warnf("centraldata: publish %s %s: %v",
			change.GetTable(), change.GetId(), err)
	}
}

// idOf the id of a row of any model.
func idOf(row proto.Message) string {
	if m, ok := row.(interface{ GetId() string }); ok {
		return m.GetId()
	}

	return ""
}

func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

// closedSubscription ends the listen loop once ctx is done.
type closedSubscription struct{}

func (closedSubscription) Next(ctx context.Context) (postgres.Change, error) {
	<-ctx.Done()

	return postgres.Change{}, ctx.Err()
}

func (closedSubscription) Close(context.Context) error { return nil }
