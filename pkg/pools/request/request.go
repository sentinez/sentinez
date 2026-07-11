// Copyright 2026 Sentinéz Labs.
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

package request

import (
	"io"

	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/sync"
)

var (
	_ io.Closer = (*RequestEvent)(nil)

	pool = sync.NewPoolCtr(func() *RequestEvent {
		return &RequestEvent{RequestEvent: &typepb.RequestEvent{}}
	})
)

type RequestEvent struct {
	*typepb.RequestEvent
}

func (re *RequestEvent) Close() error {
	Release(re)

	return nil
}

func Acquire() *RequestEvent {
	return pool.Get()
}

func Release(obj *RequestEvent) {
	obj.Reset()
	pool.Put(obj)
}
