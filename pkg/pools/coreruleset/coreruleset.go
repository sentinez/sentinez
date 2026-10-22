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

package coreruleset

import (
	"io"

	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	"github.com/sentinez/shared/szrand"
	"github.com/sentinez/shared/szsync"
)

var (
	_prefix = ""

	_ io.Closer = (*CoreRuleEvent)(nil)

	_pool = szsync.NewPoolCtr(func() *CoreRuleEvent {
		return &CoreRuleEvent{CoreRuleEvent: &corerulesetpb.CoreRuleEvent{}}
	})
)

type CoreRuleEvent struct {
	*corerulesetpb.CoreRuleEvent
}

func (re *CoreRuleEvent) Close() error {
	Release(re)

	return nil
}

func Acquire() *CoreRuleEvent {
	obj := _pool.Get()
	obj.Id = szrand.NewId(_prefix)

	return obj
}

func Release(obj *CoreRuleEvent) {
	obj.Reset()
	_pool.Put(obj)
}
