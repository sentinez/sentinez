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

package edgeapi

import (
	"context"

	corehttpreq "github.com/sentinez/core/http/request"
	corerules "github.com/sentinez/core/rules"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
)

var _ edgepb.EdgeServiceServer = (*EdgeService)(nil)

func New() *EdgeService {
	return &EdgeService{}
}

type EdgeService struct{}

func (e *EdgeService) EvaluateIngress(ctx context.Context,
	request *edgepb.EvaluateIngressRequest,
) (*edgepb.EvaluateIngressResponse, error) {

	enginectx := corehttpreq.NewRequestContext(ctx, request.GetRequestContext())
	// defer corehttpreq.(enginectx)

	rule := corerules.NewIngress(&rulepb.RuleBased{})
	matched := &rulepb.MatchedRules{}
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
