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

package secrules

import (
	"sort"
	"sync"

	corehttp "github.com/sentinez/core/http"
	corerule "github.com/sentinez/core/rules"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/jsonx"
	ssync "github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	once     sync.Once
	ruleInst *SecRule
)

func New() *SecRule {
	once.Do(func() {
		ruleInst = &SecRule{
			chains: ssync.NewMap[string, []Entry](),
		}
	})

	return ruleInst
}

// Entry is a compiled rule ready to be evaluated.
type Entry struct {
	Eval corerule.EvalFunc
	Rule *secrulepb.SecRule
}

type SecRule struct {
	chains *ssync.Map[string, []Entry]
}

// Store replaces the rule chain of namespace. Inactive rules are skipped
// and the rest are ordered by priority: a higher priority value runs
// first, ties keep their given order.
func (sr *SecRule) Store(namespace string, rules []*secrulepb.SecRule) {
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
		sr.chains.Delete(namespace)
		return
	}

	sr.chains.Store(namespace, chain)
}

// Load returns the priority-ordered chain of namespace. The result is
// shared and must not be modified.
func (sr *SecRule) Load(namespace string) []Entry {
	chain, _ := sr.chains.Load(namespace)
	return chain
}

func (sr *SecRule) LoadContext(ctx corehttp.Context) []Entry {
	if sr == nil {
		return nil
	}

	return sr.Load(ctx.X().GetNamespace())
}
