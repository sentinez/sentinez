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

package cluster

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/olric-data/olric"
	"github.com/redis/go-redis/v9"
	"github.com/sentinez/pkg/szlog"
)

// Start waits for the cluster, subscribes to the map, loads what the
// other nodes hold, then keeps the map in sync until Close. It
// subscribes before it loads, so no write falls in between; writes made
// before Start are shared by that first load.
func (m *SyncMap[T, PT]) Start(ctx context.Context) error {
	if m.remote.Load() != nil {
		return errSyncStarted
	}

	select {
	case <-m.cluster.Ready():
	case <-ctx.Done():
		return ctx.Err()
	}

	r, err := m.connect(ctx)
	if err != nil {
		return err
	}

	runCtx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	m.remote.Store(r)

	// a failed load is retried by the next reconcile
	if err := m.reconcile(ctx, r); err != nil {
		szlog.Warnf("sync %s: load: %v", m.name, err)
	}

	m.wg.Add(2)
	go m.receive(runCtx, r.sub)
	go m.reconcileEvery(runCtx, r)

	return nil
}

// Close stops syncing; the local values stay.
func (m *SyncMap[T, PT]) Close() error {
	r := m.remote.Load()
	if r == nil {
		return nil
	}

	r.cancel()
	err := r.sub.Close()
	m.wg.Wait()

	return err
}

// connect opens the DMap of the map and subscribes to its channel.
func (m *SyncMap[T, PT]) connect(ctx context.Context) (*remote, error) {
	dm, err := m.cluster.client.NewDMap(m.name)
	if err != nil {
		return nil, fmt.Errorf("sync %s: dmap: %w", m.name, err)
	}

	ps, sub, err := m.cluster.subscribe(ctx, m.channel())
	if err != nil {
		return nil, fmt.Errorf("sync %s: %w", m.name, err)
	}

	return &remote{dmap: dm, pubsub: ps, sub: sub}, nil
}

// receive applies the writes the other nodes publish.
func (m *SyncMap[T, PT]) receive(ctx context.Context, sub *redis.PubSub) {
	defer m.wg.Done()

	for msg := range sub.Channel() {
		key, rec, err := decodeMessage([]byte(msg.Payload))
		if err != nil || rec.origin == m.cluster.node {
			continue
		}

		if err := m.take(key, rec); err != nil && ctx.Err() == nil {
			szlog.Warnf("sync %s: receive %s: %v", m.name, key, err)
		}
	}
}

func (m *SyncMap[T, PT]) reconcileEvery(ctx context.Context, r *remote) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		if err := m.reconcile(ctx, r); err != nil && ctx.Err() == nil {
			szlog.Warnf("sync %s: reconcile: %v", m.name, err)
		}
	}
}

// reconcile takes the records of the DMap newer than the local ones,
// then writes back the local ones the DMap lacks or holds older.
func (m *SyncMap[T, PT]) reconcile(ctx context.Context, r *remote) error {
	stored, err := m.pull(ctx, r)
	if err != nil {
		return err
	}

	m.local.Range(func(key string, e entry[T]) bool {
		if m.expired(key, e) {
			return true
		}

		if s, ok := stored[key]; !ok || e.newer(s) {
			if err := m.put(ctx, r, key, e.record); err != nil {
				szlog.Warnf("sync %s: repair %s: %v", m.name, key, err)
			}
		}

		return ctx.Err() == nil
	})

	return nil
}

// pull applies every record of the DMap and returns them by key.
func (m *SyncMap[T, PT]) pull(
	ctx context.Context, r *remote,
) (map[string]record, error) {
	it, err := r.dmap.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("sync %s: scan: %w", m.name, err)
	}
	defer it.Close()

	stored := make(map[string]record)
	for it.Next() {
		key := it.Key()

		rec, err := m.get(ctx, r, key)
		if err != nil {
			continue
		}

		// the version is all reconcile compares
		stored[key] = record{version: rec.version, origin: rec.origin}
		if err := m.take(key, rec); err != nil {
			szlog.Warnf("sync %s: pull %s: %v", m.name, key, err)
		}
	}

	return stored, nil
}

// take applies rec to key if it is newer, decoding it only then; apply
// checks again under the lock.
func (m *SyncMap[T, PT]) take(key string, rec record) error {
	if cur, ok := m.local.Load(key); ok && !rec.newer(cur.record) {
		return nil
	}

	e, err := m.decode(rec)
	if err != nil {
		return err
	}

	m.apply(key, e)

	return nil
}

// get reads the record of key from the DMap.
func (m *SyncMap[T, PT]) get(
	ctx context.Context, r *remote, key string,
) (record, error) {
	resp, err := r.dmap.Get(ctx, key)
	if err != nil {
		if !errors.Is(err, olric.ErrKeyNotFound) {
			szlog.Warnf("sync %s: get %s: %v", m.name, key, err)
		}

		return record{}, err
	}

	raw, err := resp.Byte()
	if err != nil {
		return record{}, err
	}

	return decodeRecord(raw)
}

// expired drops a local tombstone older than the DMap keeps it.
func (m *SyncMap[T, PT]) expired(key string, e entry[T]) bool {
	age := time.Duration(m.clock.now() - min(e.version, m.clock.now()))
	if !e.deleted || age < _tombstoneTTL {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if cur, ok := m.local.Load(key); ok && cur.version == e.version {
		m.local.Delete(key)
	}

	return true
}
