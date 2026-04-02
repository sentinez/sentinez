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
	corerules "github.com/sentinez/core/rules"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/shared/chains"
	"github.com/sentinez/sentinez/internal/shared/mem/ruleengine"
	httpxcmn "github.com/sentinez/sentinez/pkg/network/httpx/common"
	"github.com/sentinez/shared/zlog"
)

func NewRule(ll zlog.Level) chains.Handler {
	return &Rule{
		BaseHandler: chains.New(),
		ingress:     corerules.NewIngress(),
		logger: zlog.NewJSONLogger(
			edgepb.GetMetaEdgeServiceKey(),
			typepb.LogKind_LOG_KIND_RULE, ll,
		),
	}
}

type Rule struct {
	*chains.BaseHandler
	ingress corerules.Rules
	logger  zlog.Logger
}

func (r *Rule) Handle(ctx corehttp.Context) error {
	// zlog.Debug("[edge] >>> visit rule")

	rule := ruleengine.GetEngine().LoadContext(ctx)
	matched, ok := r.ingress.EvalExpr(ctx, rule)
	if ok {
		zlog.Debugf("[edge] matched rule %v", matched)
		return httpxcmn.Forbidden(ctx)
	}

	return r.HandleNext(ctx)
}
