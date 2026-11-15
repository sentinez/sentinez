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

// Package centralsync keeps the edge equal to the tables centraldata syncs:
// it loads them with the Sync stream, then follows the changes centraldata
// publishes on the Olric channel.
package centralsync

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szsign"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
)

const (
	// DefaultResync how often the tables are loaded again: a change lost
	// at the end of the channel shows in no gap, and this finds it.
	DefaultResync = 5 * time.Minute

	_retryMin = time.Second
	_retryMax = 30 * time.Second
	_retryHot = 100 * time.Millisecond
)

// errResync ends a session: the edge must load the tables again.
var errResync = errors.New("centralsync: resync")

// Stream the rows of the Sync call.
type Stream interface {
	Recv() (*centraldatapb.SyncResponse, error)
}

// Source is centraldata, as the edge reaches it.
type Source interface {
	// Subscribe returns the messages published on the channel, each
	// payload an envelope; the closer ends the subscription. It must
	// return before Sync is called, so no change falls between the two.
	Subscribe(ctx context.Context) (<-chan *redis.Message, io.Closer, error)
	// Sync opens the stream of the tables.
	Sync(ctx context.Context) (Stream, error)
}

// Sink applies the tables to the edge.
type Sink interface {
	// Replace makes the edge hold exactly the rows of changes, which are
	// all upserts: what it holds and they lack is removed.
	Replace(changes []*centraldatapb.Change) error
	// Apply applies one change.
	Apply(change *centraldatapb.Change) error
}

// Follower loads the tables of centraldata into a Sink and keeps them
// current.
type Follower struct {
	source Source
	sink   Sink
	key    []byte
	resync time.Duration

	ready     chan struct{}
	readyOnce sync.Once
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// New returns a Follower; envelopes must carry the MAC of key. resync is
// the period of the full reload ([DefaultResync] when 0).
func New(source Source, sink Sink, key []byte, resync time.Duration,
) *Follower {
	if resync <= 0 {
		resync = DefaultResync
	}

	return &Follower{
		source: source,
		sink:   sink,
		key:    key,
		resync: resync,
		ready:  make(chan struct{}),
	}
}

// Start follows centraldata in the background until Close, and waits for
// the first load. When ctx ends first it returns its error: the Follower
// keeps trying, and the edge goes on with what it has.
func (f *Follower) Start(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(context.Background())
	f.cancel = cancel

	f.wg.Add(1)
	go f.run(runCtx)

	select {
	case <-f.ready:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("centralsync: first load: %w", ctx.Err())
	}
}

// Close stops following and waits for it.
func (f *Follower) Close() {
	if f.cancel != nil {
		f.cancel()
	}

	f.wg.Wait()
}

func (f *Follower) run(ctx context.Context) {
	defer f.wg.Done()

	wait := _retryMin
	for ctx.Err() == nil {
		err := f.session(ctx)
		if errors.Is(err, errResync) {
			wait = _retryMin
			sleep(ctx, _retryHot)

			continue
		}

		if ctx.Err() != nil {
			return
		}

		szlog.Warnf("centralsync: %v, retry in %s", err, wait)
		sleep(ctx, wait)
		wait = min(wait*2, _retryMax)
	}
}

// session subscribes, loads the tables and follows the changes until
// something requires a new load, or fails.
func (f *Follower) session(ctx context.Context) error {
	msgs, closer, err := f.source.Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}
	defer func() { _ = closer.Close() }()

	changes, snap, err := f.load(ctx)
	if err != nil {
		return err
	}

	if err := f.sink.Replace(changes); err != nil {
		return fmt.Errorf("replace: %w", err)
	}

	szlog.Infof("centralsync: loaded %d rows at version %d",
		len(changes), snap.GetVersion())
	f.readyOnce.Do(func() { close(f.ready) })

	return f.follow(ctx, msgs, snap)
}

// load reads the stream up to its snapshot.
func (f *Follower) load(ctx context.Context,
) ([]*centraldatapb.Change, *centraldatapb.Snapshot, error) {
	stream, err := f.source.Sync(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("sync: %w", err)
	}

	var changes []*centraldatapb.Change
	for {
		resp, err := stream.Recv()
		if err != nil {
			return nil, nil, fmt.Errorf("sync stream: %w", err)
		}

		if done := resp.GetDone(); done != nil {
			return changes, done, nil
		}

		if row := resp.GetRow(); row != nil {
			changes = append(changes, row)
		}
	}
}

// follow applies the published changes after snap, in order, until a gap, a
// new epoch or the periodic reload.
func (f *Follower) follow(ctx context.Context, msgs <-chan *redis.Message,
	snap *centraldatapb.Snapshot) error {

	timer := time.NewTimer(f.resync)
	defer timer.Stop()

	last := snap.GetVersion()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return errResync
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("subscription closed")
			}

			next, err := f.step([]byte(msg.Payload), snap.GetEpoch(), last)
			if err != nil {
				return err
			}

			last = next
		}
	}
}

// step applies the envelope data when it is the change after last, and
// returns the version reached. A forged or old envelope changes nothing.
func (f *Follower) step(data []byte, epoch string, last uint64,
) (uint64, error) {
	change, err := szsign.Open(f.key, data)
	if err != nil {
		szlog.Warnf("centralsync: dropped a change: %v", err)

		return last, nil
	}

	if change.GetEpoch() != epoch || change.GetTable() == "" {
		return last, errResync
	}

	switch v := change.GetVersion(); {
	case v <= last:
		return last, nil
	case v != last+1:
		szlog.Warnf("centralsync: gap, version %d after %d", v, last)

		return last, errResync
	}

	if err := f.sink.Apply(change); err != nil {
		return last, fmt.Errorf("apply %s %s: %w",
			change.GetTable(), change.GetId(), err)
	}

	return change.GetVersion(), nil
}

func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
