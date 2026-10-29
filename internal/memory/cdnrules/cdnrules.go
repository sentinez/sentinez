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
	"sort"
	"sync"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szrule"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	cdnpb "github.com/sentinez/sentinez/api/proto/sentinez/types/cdn/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/szjson"
	"github.com/sentinez/shared/szlog"
	"github.com/sentinez/shared/szsync"
)

var (
	_inst *Rule
	_once sync.Once
)

func New() *Rule {
	_once.Do(func() {
		_inst = &Rule{
			chains: szsync.NewMap[string, []Entry](),
		}
	})
	return _inst
}

// Entry is one CDN rule of a namespace chain.
type Entry struct {
	Eval szrule.EvalFunc
	Rule *cdnpb.Rule
}

type Rule struct {
	chains *szsync.Map[string, []Entry]
}

// Store replaces the CDN rule chain of namespace with the rules of
// delivery. Inactive rules are skipped and the rest are ordered by
// priority: a higher priority value runs first, ties keep their given
// order.
func (rc *Rule) Store(namespace string, delivery *edgepb.Delivery) {
	chain := make([]Entry, 0, len(delivery.GetCdn()))
	for _, cdn := range delivery.GetCdn() {
		rule := cdn.GetRuleRuntime()
		if rule.GetStatus() != typepb.Status_STATUS_ACTIVE {
			continue
		}

		val, _ := szjson.Marshal(rule)
		szlog.Debugf("cdn rule: load config: %s", val)

		chain = append(chain, Entry{
			Eval: szrule.NewEval(rule.GetId(), rule.GetExpr()),
			Rule: rule,
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

func (rc *Rule) Load(namespace string) []Entry {
	chain, _ := rc.chains.Load(namespace)
	return chain
}

func (rc *Rule) LoadContext(ctx szhttp.Context) []Entry {
	if rc == nil {
		return nil
	}

	return rc.Load(ctx.X().GetNamespace())
}
