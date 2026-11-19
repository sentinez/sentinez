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

package ruleset

import (
	"io"

	"github.com/sentinez/pkg/szprefix"
	"github.com/sentinez/pkg/szrand"
	"github.com/sentinez/pkg/szsync"
	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
)

var (
	_ io.Closer = (*RulesetEvent)(nil)

	_pool = szsync.NewPoolCtr(func() *RulesetEvent {
		return &RulesetEvent{RulesetEvent: &rulesetpb.RulesetEvent{}}
	})
)

type RulesetEvent struct {
	*rulesetpb.RulesetEvent
}

func (re *RulesetEvent) Close() error {
	Release(re)

	return nil
}

func Acquire() *RulesetEvent {
	obj := _pool.Get()
	obj.Id = szrand.NewULID(szprefix.RulesetEvent)

	return obj
}

func Release(obj *RulesetEvent) {
	obj.ResetVT()
	_pool.Put(obj)
}
