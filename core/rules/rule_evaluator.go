// Copyright 2025 Duc-Hung Ho.
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

package corerule

import (
	"sync"

	ruleengpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	corehttp "github.com/sentinez/sentinez/core/http"
)

var (
	_ Evaluator = (*evaluator)(nil)

	evPool = sync.Pool{
		New: func() any {
			return &evaluator{}
		},
	}
)

type Evaluator interface {
	visitBinary(cond *ruleengpb.Condition) bool
	visitLogical(cond *ruleengpb.Condition) bool

	Release()
}

// newEvaluator creates a new Evaluator instance.
// Remember to call Evaluator.Release when the
// context is done to avoid memory leaks.
func newEvaluator(ctx corehttp.RequestContext) Evaluator {

	ev := evPool.Get().(*evaluator)
	ev.ctx = ctx

	return ev
}

type evaluator struct {
	ctx corehttp.RequestContext
}

func (ev *evaluator) Release() {
	ev.ctx = nil
	evPool.Put(ev)
}

func (ev *evaluator) visitBinary(cond *ruleengpb.Condition) bool {
	// zlog.Debug("ev: visit binary")

	switch cond.GetSource() {

	case ruleengpb.FieldSource_FIELD_SOURCE_PATH:
		return matchSourcePath(ev.ctx, cond)

	case ruleengpb.FieldSource_FIELD_SOURCE_QUERY:
		return matchSourceQuery(ev.ctx, cond)

	case ruleengpb.FieldSource_FIELD_SOURCE_BODY:
		return bypass

	case ruleengpb.FieldSource_FIELD_SOURCE_HEADER:
		return bypass

	case ruleengpb.FieldSource_FIELD_SOURCE_METHOD:
		return matchSourceMethod(ev.ctx, cond)

	case ruleengpb.FieldSource_FIELD_SOURCE_JA4:
		return bypass

	case ruleengpb.FieldSource_FIELD_SOURCE_HOST:
		return bypass

	case ruleengpb.FieldSource_FIELD_SOURCE_IP:
		return matchSourceIP(ev.ctx, cond)

	case ruleengpb.FieldSource_FIELD_SOURCE_TLS:
		return bypass

	default:
		return bypass
	}
}

func (ev *evaluator) visitLogical(cond *ruleengpb.Condition) bool {
	switch cond.GetLogic() {
	case ruleengpb.Logic_LOGIC_AND:
		for _, child := range cond.GetChildren() {
			childX := newCondition(child)

			if !childX.Accept(ev) {
				childX.Release()

				return false
			}

			childX.Release()
		}

		return true

	case ruleengpb.Logic_LOGIC_OR:
		for _, child := range cond.Children {
			childX := newCondition(child)

			if childX.Accept(ev) {
				childX.Release()

				return true
			}

			childX.Release()
		}

		return false

	default:
		return false
	}
}
