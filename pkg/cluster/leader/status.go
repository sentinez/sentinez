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
	_statusPrefix   = "sentinez/dmz/edge/zone/"
	_statusInterval = 10 * time.Second
)

// Status is the record the leader keeps of its zone, in Consul under
// sentinez/dmz/edge/zone/<zone>, fenced by Term.
type Status struct {
	Zone      string    `json:"zone"`
	Node      string    `json:"node"`
	Term      uint64    `json:"term"`
	Version   string    `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

// fencedStore stores a value with the fencing token of its writer and
// rejects a lower token with consul.ErrFenced, see consul.PutFenced.
type fencedStore interface {
	PutFenced(ctx context.Context, key string, token uint64, value any) error
}

var _ fencedStore = (*consul.Client)(nil)

// statusTarget is where the Status of a zone is written.
type statusTarget struct {
	store    fencedStore
	zone     string
	interval time.Duration
}

func newStatusTarget(store fencedStore, zone string) *statusTarget {
	return &statusTarget{
		store:    store,
		zone:     zone,
		interval: _statusInterval,
	}
}

// PublishStatus returns the Task that writes the Status of the zone at
// once, then every interval while the node leads; version names the
// setting the node runs. A fenced write ends the term: a newer leader
// wrote the record. Other failures are logged and retried at the next
// tick. A standalone Elector has nowhere to write: nil, it only leads.
func (e *Elector) PublishStatus(version string) Task {
	target := e.status
	if target == nil {
		return nil
	}

	return func(ctx context.Context, term Term) error {
		ticker := time.NewTicker(target.interval)
		defer ticker.Stop()

		for {
			if err := target.put(ctx, term, version); err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
			}
		}
	}
}

// put writes the Status of term; only a fenced write is an error.
func (t *statusTarget) put(
	ctx context.Context, term Term, version string,
) error {
	err := t.store.PutFenced(ctx, _statusPrefix+t.zone, term.Index, &Status{
		Zone:      t.zone,
		Node:      term.Node,
		Term:      term.Index,
		Version:   version,
		UpdatedAt: time.Now().UTC(),
	})
	if errors.Is(err, consul.ErrFenced) {
		return err
	}

	if err != nil && ctx.Err() == nil {
		szlog.Warnf("leader: zone status term %d: %v", term.Index, err)
	}

	return nil
}
