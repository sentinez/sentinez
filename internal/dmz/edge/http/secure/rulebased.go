// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package secure

import (
	corehttp "github.com/sentinez/core/http"
	corechains "github.com/sentinez/core/http/chains"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/secure/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	matchedPool = sync.NewPool[rulepb.MatchedRules]()
)

func NewRuleBased(ll zlog.Level, store *memory.MemStore) corechains.ChainNode {
	return &RuleBased{
		Node:  corechains.NewNode(),
		store: store,
		logger: zlog.NewLog(
			edgepb.GetMetaEdgeServiceKey(),
			typepb.LogKind_LOG_KIND_RULE_BASED, ll,
		),
	}
}

type RuleBased struct {
	*corechains.Node
	logger zlog.Log
	store  *memory.MemStore
}

// Handle runs the active rules of the namespace in priority order. The
// first matching rule with a terminal action (block) ends the chain; other
// matches fall through to the next rule.
func (rb *RuleBased) Handle(ctx corehttp.Context) error {
	chain := rb.store.RuleBased().LoadContext(ctx)
	if len(chain) == 0 {
		return rb.HandleNext(ctx)
	}

	matched := matchedPool.Get()
	defer matchedPool.Put(matched)

	for _, e := range chain {
		matched.Reset()
		if !e.Eval(ctx, matched) {
			continue
		}

		action := e.Rule.GetAction().GetType()
		zlog.Debugf("edge: rule %s matched, action = %v",
			e.Rule.GetId(), action)

		if action == rulepb.ActionType_ACTION_TYPE_BLOCK {
			return corehttp.Forbidden(ctx)
		}
	}

	return rb.HandleNext(ctx)
}
