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

package consul

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

const (
	// _lockMonitorRetries tolerates short Consul blips before a lock is
	// reported lost; retries end well before the session TTL expires.
	_lockMonitorRetries = 3
	// _lockWaitTime bounds each wait for a held lock, so Acquire notices
	// its ctx ending well within a shutdown timeout.
	_lockWaitTime = 2 * time.Second
)

// ErrLockLost is returned when a lock is lost before it is fenced.
var ErrLockLost = errors.New("consul: lock lost")

// Lock is a Consul lock held through a session that Consul invalidates
// ttl after its holder stops renewing it.
type Lock struct {
	lock  *api.Lock
	lost  <-chan struct{}
	index uint64
}

// Acquire blocks until it holds the lock on key or ctx ends. value tells
// the holder apart.
func (c *Client) Acquire(
	ctx context.Context, key string, value []byte, ttl time.Duration,
) (*Lock, error) {
	lock, err := c.consul.LockOpts(&api.LockOptions{
		Key:            key,
		Value:          bytes.Clone(value),
		SessionName:    key,
		SessionTTL:     ttl.String(),
		MonitorRetries: _lockMonitorRetries,
		LockWaitTime:   _lockWaitTime,
	})
	if err != nil {
		return nil, fmt.Errorf("new lock %q: %w", key, err)
	}

	stop := make(chan struct{})
	defer context.AfterFunc(ctx, func() { close(stop) })()

	lost, err := lock.Lock(stop)
	if err != nil {
		return nil, fmt.Errorf("lock %q: %w", key, err)
	}

	if lost == nil {
		return nil, ctx.Err()
	}

	l := &Lock{lock: lock, lost: lost}
	if l.index, err = c.lockIndex(ctx, key, value); err != nil {
		_ = l.Release()
		return nil, err
	}

	return l, nil
}

// Lost is closed once the lock is lost.
func (l *Lock) Lost() <-chan struct{} { return l.lost }

// Index is the LockIndex of the key, which grows each time the lock
// changes hands: a fencing token.
func (l *Lock) Index() uint64 { return l.index }

// Release gives the lock up; releasing a lost lock is a no-op.
func (l *Lock) Release() error {
	err := l.lock.Unlock()
	if errors.Is(err, api.ErrLockNotHeld) {
		return nil
	}

	return err
}

// lockIndex returns the LockIndex of key if it is locked with value.
func (c *Client) lockIndex(
	ctx context.Context, key string, value []byte,
) (uint64, error) {
	opts := (&api.QueryOptions{RequireConsistent: true}).WithContext(ctx)

	pair, _, err := c.consul.KV().Get(key, opts)
	if err != nil {
		return 0, fmt.Errorf("get %q: %w", key, err)
	}

	if pair == nil || pair.Session == "" || !bytes.Equal(pair.Value, value) {
		return 0, ErrLockLost
	}

	return pair.LockIndex, nil
}
