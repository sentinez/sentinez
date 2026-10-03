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
	"sort"
	"sync"

	corehttp "github.com/sentinez/core/http"
	corerule "github.com/sentinez/core/rules"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/secure/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
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
			chains: ssync.NewMap[string, []Entry](),
		}
	})

	return ruleInst
}

// Entry is a compiled rule ready to be evaluated.
type Entry struct {
	Eval corerule.EvalFunc
	Rule *rulepb.RuleIngress
}

type RuleBased struct {
	chains *ssync.Map[string, []Entry]
}

// Store replaces the rule chain of namespace. Inactive rules are skipped
// and the rest are ordered by priority: a higher priority value runs
// first, ties keep their given order.
func (rc *RuleBased) Store(namespace string, rules []*rulepb.RuleIngress) {
	chain := make([]Entry, 0, len(rules))
	for _, r := range rules {
		if r.GetStatus() != typepb.Status_STATUS_ACTIVE {
			continue
		}

		val, _ := jsonx.Marshal(r)
		zlog.Debugf("rule: load config: %s", val)

		chain = append(chain, Entry{
			Eval: corerule.NewEval(r.GetExpr()),
			Rule: r,
		})
	}

	sort.SliceStable(chain, func(i, j int) bool {
		return chain[i].Rule.GetPriority() > chain[j].Rule.GetPriority()
	})

	if len(chain) == 0 {
		rc.chains.Delete(namespace)
		return
	}

	rc.chains.Store(namespace, chain)
}

// Load returns the priority-ordered chain of namespace. The result is
// shared and must not be modified.
func (rc *RuleBased) Load(namespace string) []Entry {
	chain, _ := rc.chains.Load(namespace)
	return chain
}

func (rc *RuleBased) LoadContext(ctx corehttp.Context) []Entry {
	if rc == nil {
		return nil
	}

	return rc.Load(ctx.X().GetNamespace())
}
