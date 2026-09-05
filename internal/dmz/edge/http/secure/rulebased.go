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
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
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
			typepb.LogKind_LOG_KIND_RULE, ll,
		),
	}
}

type RuleBased struct {
	*corechains.Node
	logger zlog.Log
	store  *memory.MemStore
}

func (rb *RuleBased) Handle(ctx corehttp.Context) error {
	// zlog.Debug("[edge] >>> visit rule")

	eval, rule := rb.store.RuleBased().LoadContext(ctx)
	if eval == nil || rule == nil {
		return rb.HandleNext(ctx)
	}

	matched := matchedPool.Get()
	defer matchedPool.Put(matched)

	if ok := eval(ctx, matched); !ok {
		return rb.HandleNext(ctx)
	}

	zlog.Debugf("edge: action = %v", rule.GetAction().GetType())

	switch rule.GetAction().GetType() {
	case rulepb.ActionType_ACTION_TYPE_BLOCK:
		zlog.Debugf("edge: matched rule %v", matched)
		return corehttp.Forbidden(ctx)
	}

	return rb.HandleNext(ctx)
}
