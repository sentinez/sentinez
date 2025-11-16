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
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	corehttp "github.com/sentinez/sentinez/core/http"
	corerules "github.com/sentinez/sentinez/core/rules"
	"github.com/sentinez/sentinez/pkg/dmz/chains"
	"github.com/sentinez/sentinez/pkg/dmz/mem/ruleengine"
	httpxcmn "github.com/sentinez/sentinez/pkg/network/httpx/common"
	"github.com/sentinez/sentinez/shared/zlog"
)

func NewRule(ll zlog.Level) *Rule {
	return &Rule{
		BaseHandler: chains.New(),
		ingress:     corerules.NewIngress(),
		logger: zlog.NewJSONLogger(
			edgepb.GetMetaEdgeServiceKey(),
			common.LogKind_LOG_KIND_RULE, ll,
		),
	}
}

type Rule struct {
	*chains.BaseHandler
	ingress corerules.Rules
	logger  zlog.Logger
}

func (r *Rule) Handle(ctx corehttp.Context) error {
	zlog.Debugf("[edge][%s] >>> visit rule", ctx.RequestId())

	rule := ruleengine.GetEngine().LoadContext(ctx)
	if ok := r.ingress.EvalExpr(ctx, rule); ok {
		return httpxcmn.Forbidden(ctx)
	}

	return r.HandleNext(ctx)
}
