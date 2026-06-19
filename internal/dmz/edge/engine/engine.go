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

	corerules "github.com/sentinez/core/rules"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleengine/v1"
	"github.com/sentinez/sentinez/internal/dmz/edge/engine/requests"
)

var _ edgepb.EdgeEngineServiceServer = (*Engine)(nil)

func New() *Engine {
	return &Engine{}
}

type Engine struct{}

func (e *Engine) EvaluateIngress(ctx context.Context,
	request *edgepb.EvaluateIngressRequest,
) (*edgepb.EvaluateIngressResponse, error) {

	enginectx := requests.New(ctx, request.GetRequestContext())
	defer requests.Free(enginectx)

	rule := corerules.NewIngress(&ruleenginepb.RuleBased{})
	matched := &ruleenginepb.MatchedRules{}
	ok := rule.Eval(enginectx, matched)
	if !ok {
		return &edgepb.EvaluateIngressResponse{}, nil
	}

	return &edgepb.EvaluateIngressResponse{
		Results: []*edgepb.EvaluationResult{{
			Matched: ok,
		}},
	}, nil
}
