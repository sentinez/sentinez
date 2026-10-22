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

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szrule"
	cdnpb "github.com/sentinez/sentinez/api/proto/sentinez/types/cdn/v1"
	"github.com/sentinez/shared/jsonx"
	ssync "github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	_inst *Rule
	_once sync.Once
)

func New() *Rule {
	_once.Do(func() {
		_inst = &Rule{
			space:   ssync.NewMap[string, szrule.EvalFunc](),
			cdnRule: ssync.NewMap[string, *cdnpb.CDN](),
		}
	})
	return _inst
}

type Rule struct {
	space   *ssync.Map[string, szrule.EvalFunc]
	cdnRule *ssync.Map[string, *cdnpb.CDN]
}

func (rc *Rule) Store(namespace string, rule *cdnpb.CDN) {
	val, _ := jsonx.Marshal(rule)
	zlog.Debugf("cdn rule: load config: %s", val)

	ev := szrule.NewEval(rule.GetRuleRuntime().GetExpr())

	rc.space.Store(namespace, ev)
	rc.cdnRule.Store(namespace, rule)
}

func (rc *Rule) Load(namespace string) (szrule.EvalFunc, *cdnpb.CDN) {
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
	ctx szhttp.Context) (szrule.EvalFunc, *cdnpb.CDN) {

	if rc == nil {
		return nil, nil
	}

	return rc.Load(ctx.X().GetNamespace())
}
