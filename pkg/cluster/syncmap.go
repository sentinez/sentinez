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
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olric-data/olric"
	"github.com/redis/go-redis/v9"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szsync"
	"google.golang.org/protobuf/proto"
)

var errSyncStarted = errors.New("cluster: sync map already started")

const (
	_syncChannelPrefix = "sentinez.sync."
	_syncInterval      = 30 * time.Second
	_tombstoneTTL      = 24 * time.Hour
)

// protoPtr is *T when *T is a proto message.
type protoPtr[T any] interface {
	*T
	proto.Message
}

// entry is the record of a key with its decoded value.
type entry[T any] struct {
	record
	value *T
}

// remote is the Olric side of a started SyncMap.
type remote struct {
	dmap   olric.DMap
	pubsub *olric.PubSub
	sub    *redis.PubSub
	// cancel stops the sync goroutines
	cancel context.CancelFunc
}

// SyncMap keeps a map in sync between the nodes of a cluster. Reads are
// local and lock-free; a write is applied locally, then stored in the
// Olric DMap of the map and published to the other nodes, which apply
// it if it is newer (last writer wins). PubSub is at most once, so every
// interval the map reconciles with the DMap: it takes the newer records
// and writes back the ones it holds newer.
type SyncMap[T any, PT protoPtr[T]] struct {
	name     string
	cluster  *Cluster
	clock    *hlc
	interval time.Duration

	mu       sync.Mutex // serializes the compare-and-store of apply
	local    *szsync.Map[string, entry[T]]
	onChange func(key string, prev, next PT)

	remote atomic.Pointer[remote]
	wg     sync.WaitGroup
}

// NewSyncMap returns the map name of c; it is local until Start.
func NewSyncMap[T any, PT protoPtr[T]](
	c *Cluster, name string,
) *SyncMap[T, PT] {
	return &SyncMap[T, PT]{
		name:     name,
		cluster:  c,
		clock:    newHLC(),
		interval: _syncInterval,
		local:    szsync.NewMap[string, entry[T]](),
	}
}

// OnChange sets fn to be called each time the value of a key changes on
// this node: written here, received or reconciled. prev or next is nil
// when the key had or has no value. Calls are serialized, in the order
// the changes apply; fn must not write the map. Set it before Start.
func (m *SyncMap[T, PT]) OnChange(fn func(key string, prev, next PT)) {
	m.onChange = fn
}

// Load returns the value of key on this node.
func (m *SyncMap[T, PT]) Load(key string) (PT, bool) {
	e, _ := m.local.Load(key)
	v := e.live()

	return v, v != nil
}

// Range calls fn for each value on this node until it returns false.
func (m *SyncMap[T, PT]) Range(fn func(key string, value PT) bool) {
	m.local.Range(func(key string, e entry[T]) bool {
		v := e.live()
		return v == nil || fn(key, v)
	})
}

// Store sets key to value on every node; value must not change after.
// A value equal to the current one is not written again, so the other
// nodes do not see a change.
func (m *SyncMap[T, PT]) Store(
	ctx context.Context, key string, value PT,
) error {
	payload, err := proto.MarshalOptions{Deterministic: true}.Marshal(value)
	if err != nil {
		return fmt.Errorf("sync %s: encode %s: %w", m.name, key, err)
	}

	if cur, ok := m.local.Load(key); ok && !cur.deleted &&
		bytes.Equal(cur.payload, payload) {
		return nil
	}

	return m.write(ctx, key, entry[T]{value: value, record: record{
		payload: payload,
	}})
}

// Delete removes key on every node, through a tombstone kept for a day.
func (m *SyncMap[T, PT]) Delete(ctx context.Context, key string) error {
	return m.write(ctx, key, entry[T]{record: record{deleted: true}})
}

// write versions e as the newest write of key, applies it, then shares
// it if the map is started; else the next reconcile shares it.
func (m *SyncMap[T, PT]) write(
	ctx context.Context, key string, e entry[T],
) error {
	e.version = m.clock.next()
	e.origin = m.cluster.node
	m.apply(key, e)

	r := m.remote.Load()
	if r == nil {
		return nil
	}

	if err := m.put(ctx, r, key, e.record); err != nil {
		return err
	}

	msg := encodeMessage(key, e.record)
	if _, err := r.pubsub.Publish(ctx, m.channel(), msg); err != nil {
		szlog.Warnf("sync %s: publish %s: %v", m.name, key, err)
	}

	return nil
}

// apply stores e as the value of key if it is newer than the local one.
func (m *SyncMap[T, PT]) apply(key string, e entry[T]) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	cur, ok := m.local.Load(key)
	if ok && !e.newer(cur.record) {
		return false
	}

	m.local.Store(key, e)
	m.clock.observe(e.version)

	if prev, next := cur.live(), e.live(); m.onChange != nil &&
		(prev != nil || next != nil) {
		m.onChange(key, prev, next)
	}

	return true
}

// live is the value of e, nil for a tombstone or no entry.
func (e entry[T]) live() *T {
	if e.deleted {
		return nil
	}

	return e.value
}

// put stores r in the DMap; a tombstone expires after a day.
func (m *SyncMap[T, PT]) put(
	ctx context.Context, r *remote, key string, rec record,
) error {
	var opts []olric.PutOption
	if rec.deleted {
		opts = append(opts, olric.EX(_tombstoneTTL))
	}

	if err := r.dmap.Put(ctx, key, rec.encode(), opts...); err != nil {
		return fmt.Errorf("sync %s: put %s: %w", m.name, key, err)
	}

	return nil
}

// decode turns r into the entry of a key.
func (m *SyncMap[T, PT]) decode(r record) (entry[T], error) {
	e := entry[T]{record: r}
	if r.deleted {
		return e, nil
	}

	value := PT(new(T))
	if err := proto.Unmarshal(r.payload, value); err != nil {
		return e, fmt.Errorf("sync %s: decode: %w", m.name, err)
	}

	e.value = value

	return e, nil
}

func (m *SyncMap[T, PT]) channel() string {
	return _syncChannelPrefix + m.name
}
