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

package secrule

import (
	"io"

	"github.com/sentinez/pkg/szprefix"
	"github.com/sentinez/pkg/szrand"
	"github.com/sentinez/pkg/szsync"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
)

var (
	_ io.Closer = (*SecRuleEvent)(nil)

	_pool = szsync.NewPoolCtr(func() *SecRuleEvent {
		return &SecRuleEvent{SecRuleEvent: &secrulepb.SecRuleEvent{
			Action: &rulepb.Action{},
		}}
	})
)

type SecRuleEvent struct {
	*secrulepb.SecRuleEvent
}

func (re *SecRuleEvent) Close() error {
	Release(re)

	return nil
}

func Acquire() *SecRuleEvent {
	obj := _pool.Get()
	obj.Id = szrand.NewULID(szprefix.SecRuleEvent)

	return obj
}

func Release(obj *SecRuleEvent) {
	obj.ResetVT()
	_pool.Put(obj)
}
