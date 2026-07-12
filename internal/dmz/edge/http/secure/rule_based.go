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
	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleengine/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/memory/ruleengine"
	"github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	matchedPool = sync.NewPool[ruleenginepb.MatchedRules]()
)

func NewRuleBased(ll zlog.Level) corechains.ChainNode {
	return &RuleBased{
		Node: corechains.NewNode(),
		logger: zlog.NewJSONLogger(
			edgepb.GetMetaEdgeServiceKey(),
			typepb.LogKind_LOG_KIND_RULE, ll,
		),
	}
}

type RuleBased struct {
	*corechains.Node
	logger zlog.Logger
}

func (r *RuleBased) Handle(ctx corehttp.Context) error {
	// zlog.Debug("[edge] >>> visit rule")

	rule := ruleengine.GetEngine().LoadContext(ctx)
	if rule == nil {
		return r.HandleNext(ctx)
	}

	matched := matchedPool.Get()
	defer matchedPool.Put(matched)

	if ok := rule.Eval(ctx, matched); !ok {
		return r.HandleNext(ctx)
	}

	zlog.Debugf("edge: action = %v", rule.Action().GetType())

	switch rule.Action().GetType() {
	case ruleenginepb.ActionType_ACTION_TYPE_BLOCK:
		zlog.Debugf("[edge] matched rule %v", matched)
		return corehttp.Forbidden(ctx)
	}

	return r.HandleNext(ctx)
}
