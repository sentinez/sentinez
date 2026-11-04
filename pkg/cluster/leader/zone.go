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
	"time"

	"github.com/sentinez/sentinez/api/client/consul"
	"github.com/sentinez/shared/szlog"
)

const (
	_zonePrefix   = "sentinez/dmz/edge/zone/"
	_zoneInterval = 10 * time.Second
)

// fencedStore stores a value with the fencing token of its writer and
// rejects a lower token with consul.ErrFenced, see consul.PutFenced.
type fencedStore interface {
	PutFenced(ctx context.Context, key string, token uint64, value any) error
}

var _ fencedStore = (*consul.Client)(nil)

// zoneStore holds the records the leader keeps of its zone, under
// sentinez/dmz/edge/zone/<zone>/, each write fenced by the term.
type zoneStore struct {
	store    fencedStore
	zone     string
	interval time.Duration
}

func newZoneStore(store fencedStore, zone string) *zoneStore {
	return &zoneStore{store: store, zone: zone, interval: _zoneInterval}
}

// Publish returns the Task that writes value under <zone>/<name> once
// per term, retried every interval until it is written; value is
// encoded as JSON (a json.RawMessage is kept as is). A fenced write ends
// the term. A standalone Elector has nowhere to write: nil.
func (e *Elector) Publish(name string, value any) Task {
	z := e.zone
	if z == nil {
		return nil
	}

	return func(ctx context.Context, term Term) error {
		return z.loop(ctx, func() (bool, error) {
			return z.put(ctx, term, name, value)
		})
	}
}

// loop calls write at once, then every interval, until it is done, it
// fails or ctx ends.
func (z *zoneStore) loop(
	ctx context.Context, write func() (done bool, err error),
) error {
	ticker := time.NewTicker(z.interval)
	defer ticker.Stop()

	for {
		done, err := write()
		if done || err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// put writes value under <zone>/<name> with the term as fencing token:
// true once written. Only a fenced write is an error, the others are
// logged so the caller retries.
func (z *zoneStore) put(
	ctx context.Context, term Term, name string, value any,
) (bool, error) {
	key := _zonePrefix + z.zone + "/" + name

	err := z.store.PutFenced(ctx, key, term.Index, value)
	if errors.Is(err, consul.ErrFenced) {
		return false, err
	}

	if err != nil {
		if ctx.Err() == nil {
			szlog.Warnf("leader: %s term %d: %v", key, term.Index, err)
		}

		return false, nil
	}

	return true, nil
}
