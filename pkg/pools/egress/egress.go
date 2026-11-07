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

// Package egress pools the httppb.EgressEvent the edge logs per request
// proxied to an upstream origin.
package egress

import (
	"io"

	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	"github.com/sentinez/pkg/szsync"
)

var (
	_ io.Closer = (*EgressEvent)(nil)

	_pool = szsync.NewPoolCtr(func() *EgressEvent {
		return &EgressEvent{EgressEvent: &httppb.EgressEvent{}}
	})
)

// EgressEvent a pooled httppb.EgressEvent, released on Close.
type EgressEvent struct {
	*httppb.EgressEvent
}

// Close releases ee to the pool.
func (ee *EgressEvent) Close() error {
	Release(ee)

	return nil
}

// Acquire returns a reset EgressEvent of the pool.
func Acquire() *EgressEvent {
	return _pool.Get()
}

// Release resets obj and puts it back in the pool.
func Release(obj *EgressEvent) {
	obj.ResetVT()
	_pool.Put(obj)
}
