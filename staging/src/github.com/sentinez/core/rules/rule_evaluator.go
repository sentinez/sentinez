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

	corehttp "github.com/sentinez/core/http"
	ruleengpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	"github.com/sentinez/shared/zlog"
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
	visit(cond *ruleengpb.Condition) bool

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

func (ev *evaluator) visit(cond *ruleengpb.Condition) bool {
	zlog.Debugf("ev: visit with source: %s", cond.GetSource())

	switch cond.GetSource() {

	case ruleengpb.FieldSource_FIELD_SOURCE_PATH:
		return matchSourcePath(ev.ctx, cond)

	case ruleengpb.FieldSource_FIELD_SOURCE_QUERY:
		return matchSourceQuery(ev.ctx, cond)

	case ruleengpb.FieldSource_FIELD_SOURCE_BODY:
		return matchSourceBody(ev.ctx, cond)

	case ruleengpb.FieldSource_FIELD_SOURCE_HEADER:
		return matchSourceHeader(ev.ctx, cond)

	case ruleengpb.FieldSource_FIELD_SOURCE_METHOD:
		return matchSourceMethod(ev.ctx, cond)

	case ruleengpb.FieldSource_FIELD_SOURCE_HOST:
		return matchSourceHost(ev.ctx, cond)

	case ruleengpb.FieldSource_FIELD_SOURCE_IP:
		return matchSourceIP(ev.ctx, cond)

	case ruleengpb.FieldSource_FIELD_SOURCE_TLS:
		return bypass

	case ruleengpb.FieldSource_FIELD_SOURCE_JA4:
		return bypass

	default:
		return bypass
	}
}
