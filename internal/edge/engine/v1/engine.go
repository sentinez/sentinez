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

package engine

import (
	"context"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	"github.com/sentinez/sentinez/core"
	enginectx "github.com/sentinez/sentinez/internal/edge/engine/v1/context"
)

var _ edgepb.EdgeEngineServiceServer = (*Engine)(nil)

func New() *Engine {
	return &Engine{
		in: core.NewRuleIngress(),
	}
}

type Engine struct {
	in core.RuleIngress
}

func (e *Engine) EvaluateIngress(ctx context.Context,
	request *edgepb.EvaluateIngressRequest,
) (*edgepb.EvaluateIngressResponse, error) {

	ectx := enginectx.New(ctx, request.GetRequestContext())
	defer ectx.Release()

	rule := ruleenginepb.Rule{}
	if ok := e.in.ExecRule(ectx, &rule); !ok {
		return &edgepb.EvaluateIngressResponse{}, nil
	}

	return &edgepb.EvaluateIngressResponse{
		Results: []*edgepb.EvaluationResult{{
			Actions: rule.GetActions(),
		}},
	}, nil
}
