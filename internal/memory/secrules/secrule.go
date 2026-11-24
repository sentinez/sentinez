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
	"fmt"
	"sort"
	"sync"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szrule"
	"github.com/sentinez/pkg/szjson"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szsync"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/memory/scripts"
)

var (
	_once     sync.Once
	_ruleInst *SecRule
)

func New() *SecRule {
	_once.Do(func() {
		_ruleInst = &SecRule{
			chains: szsync.NewMap[string, []Entry](),
		}
	})

	return _ruleInst
}

// Entry is a rule ready to be evaluated: a script rule is converted to its
// expr and action when it loads (scripts.Rule), no Lua runs.
type Entry struct {
	Eval szrule.EvalFunc
	Rule *secrulepb.SecRule
}

type SecRule struct {
	chains *szsync.Map[string, []Entry]
}

// Store replaces the rule chain of namespace. Inactive rules are skipped
// and the rest are ordered by priority: a higher priority value runs
// first, ties keep their given order. A script out of the declarative
// form, or a rule with both an expr and a script, fails the whole call and
// leaves the previous chain in place.
func (sr *SecRule) Store(namespace string, rules []*secrulepb.SecRule) error {
	chain := make([]Entry, 0, len(rules))

	for _, r := range rules {
		if r.GetStatus() != typepb.Status_STATUS_ACTIVE {
			continue
		}

		val, _ := szjson.Marshal(r)
		szlog.Debugf("rule: load config: %s", val)

		e, err := newEntry(r)
		if err != nil {
			return fmt.Errorf("secrule %s: %w", r.GetName(), err)
		}

		chain = append(chain, e)
	}

	sort.SliceStable(chain, func(i, j int) bool {
		return chain[i].Rule.GetPriority() > chain[j].Rule.GetPriority()
	})

	if len(chain) == 0 {
		sr.chains.Delete(namespace)
		return nil
	}

	sr.chains.Store(namespace, chain)

	return nil
}

// newEntry the entry of r, its script converted to its expr and action.
func newEntry(r *secrulepb.SecRule) (Entry, error) {
	r, err := scripts.Rule(r)
	if err != nil {
		return Entry{}, err
	}

	return Entry{Eval: szrule.NewEval(r.GetId(), r.GetExpr()), Rule: r}, nil
}

// Load returns the priority-ordered chain of namespace. The result is
// shared and must not be modified.
func (sr *SecRule) Load(namespace string) []Entry {
	chain, _ := sr.chains.Load(namespace)
	return chain
}

func (sr *SecRule) LoadContext(ctx szhttp.Context) []Entry {
	if sr == nil {
		return nil
	}

	return sr.Load(ctx.X().GetNamespace())
}

// Delete removes what namespace holds.
func (sr *SecRule) Delete(namespace string) {
	sr.chains.Delete(namespace)
}
