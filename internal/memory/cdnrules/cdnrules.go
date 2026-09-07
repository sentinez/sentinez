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

package cdnrules

import (
	"sync"

	corehttp "github.com/sentinez/core/http"
	corerule "github.com/sentinez/core/rules"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/cdn/rule/v1"
	"github.com/sentinez/shared/jsonx"
	ssync "github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	inst *Rule
	once sync.Once
)

func New() *Rule {
	once.Do(func() {
		inst = &Rule{
			space:   ssync.NewMap[string, corerule.EvalFunc](),
			cdnRule: ssync.NewMap[string, *rulepb.CDN](),
		}
	})
	return inst
}

type Rule struct {
	space   *ssync.Map[string, corerule.EvalFunc]
	cdnRule *ssync.Map[string, *rulepb.CDN]
}

func (rc *Rule) Store(namespace string, rule *rulepb.CDN) {
	val, _ := jsonx.Marshal(rule)
	zlog.Debugf("cdn rule: load config: %s", val)

	ev := corerule.NewEval(rule.GetRule().GetExpr())

	rc.space.Store(namespace, ev)
	rc.cdnRule.Store(namespace, rule)
}

func (rc *Rule) Load(namespace string) (corerule.EvalFunc, *rulepb.CDN) {
	evalFunc, ok := rc.space.Load(namespace)
	if !ok {
		return nil, nil
	}

	cdnRule, ok := rc.cdnRule.Load(namespace)
	if !ok {
		return evalFunc, nil
	}

	return evalFunc, cdnRule
}

func (rc *Rule) LoadContext(
	ctx corehttp.Context) (corerule.EvalFunc, *rulepb.CDN) {

	if rc == nil {
		return nil, nil
	}

	return rc.Load(ctx.X().GetNamespace())
}
