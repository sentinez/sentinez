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

package rulebased

import (
	"sync"

	corehttp "github.com/sentinez/core/http"
	corerule "github.com/sentinez/core/rules"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/secure/rule/v1"
	"github.com/sentinez/shared/jsonx"
	ssync "github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	once     sync.Once
	ruleInst *RuleBased
)

func New() *RuleBased {
	once.Do(func() {
		ruleInst = &RuleBased{
			space: ssync.NewMap[string, corerule.EvalFunc](),
			rules: ssync.NewMap[string, *rulepb.RuleBased](),
		}
	})

	return ruleInst
}

type RuleBased struct {
	space *ssync.Map[string, corerule.EvalFunc]
	rules *ssync.Map[string, *rulepb.RuleBased]
}

func (rc *RuleBased) Store(namespace string, gr *rulepb.RuleBased) {
	val, _ := jsonx.Marshal(gr)
	zlog.Debugf("rule: load config: %s", val)

	rule := corerule.NewEval(gr.GetExpr())

	rc.space.Store(namespace, rule)
}

func (rc *RuleBased) Load(
	namespace string) (corerule.EvalFunc, *rulepb.RuleBased) {

	ev, ok := rc.space.Load(namespace)
	if !ok {
		return nil, nil
	}

	rule, ok := rc.rules.Load(namespace)
	if !ok {
		return ev, nil
	}

	return ev, rule
}

func (rc *RuleBased) LoadContext(
	ctx corehttp.Context) (corerule.EvalFunc, *rulepb.RuleBased) {

	if rc == nil {
		return nil, nil
	}

	zlog.Debugf("edge: hit rule cached %s", ctx.X().GetNamespace())
	return rc.Load(ctx.X().GetNamespace())
}
