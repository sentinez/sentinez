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
	"time"
)

// _statusName the record of the zone Status, under the zone store.
const _statusName = "status"

// Status is the record the leader keeps of its zone, in Consul under
// sentinez/dmz/edge/zone/<zone>/status, fenced by Term.
type Status struct {
	Zone      string    `json:"zone"`
	Node      string    `json:"node"`
	Term      uint64    `json:"term"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PublishStatus returns the Task that writes the Status of the zone at
// once, then every interval while the node leads. A fenced write ends
// the term: a newer leader wrote the record. Other failures are logged
// and retried at the next tick. A standalone Elector has nowhere to
// write: nil, it only leads.
func (e *Elector) PublishStatus() Task {
	z := e.zone
	if z == nil {
		return nil
	}

	return func(ctx context.Context, term Term) error {
		return z.loop(ctx, func() (bool, error) {
			_, err := z.put(ctx, term, _statusName, &Status{
				Zone:      z.zone,
				Node:      term.Node,
				Term:      term.Index,
				UpdatedAt: time.Now().UTC(),
			})

			// never done: the record is refreshed while leading
			return false, err
		})
	}
}
