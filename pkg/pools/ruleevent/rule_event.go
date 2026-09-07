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

package ruleevent

import (
	"io"

	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/secure/rule/v1"
	"github.com/sentinez/shared/sync"
)

var (
	_ io.Closer = (*RuleEvent)(nil)

	pool = sync.NewPoolCtr(func() *RuleEvent {
		return &RuleEvent{Event: &rulepb.Event{}}
	})
)

type RuleEvent struct {
	*rulepb.Event
}

func (re *RuleEvent) Close() error {
	Release(re)

	return nil
}

func Acquire() *RuleEvent {
	return pool.Get()
}

func Release(obj *RuleEvent) {
	obj.Reset()
	pool.Put(obj)
}
