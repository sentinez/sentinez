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
	"strconv"
	"time"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/sentinez/core/szcrs"
	"github.com/sentinez/core/szhttp/httpreq"
	"github.com/sentinez/core/szstorage/cache/mem"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"github.com/sentinez/sentinez/internal/memory/rulesets"
	"github.com/sentinez/shared/szerror"
)

var _ edgepb.EdgeServiceServer = (*EdgeService)(nil)

func New() *EdgeService {
	return &EdgeService{
		space: mem.NewDefault[coraza.WAF](),
	}
}

type EdgeService struct {
	space *mem.Cache[coraza.WAF]
}

// Status implements [edgepb.EdgeServiceServer].
func (e *EdgeService) Status(_ context.Context,
	_ *edgepb.StatusRequest) (*edgepb.StatusResponse, error) {

	return &edgepb.StatusResponse{Msg: "OK"}, nil
}

// EvalCoreRuleset replays request_context against a WAF built from
// core_ruleset and reports every CRS rule it matched.
func (e *EdgeService) EvalCoreRuleset(ctx context.Context,
	request *edgepb.EvalCoreRulesetRequest,
) (*edgepb.EvalCoreRulesetResponse, error) {

	var (
		waf coraza.WAF
		ok  bool
		err error
	)

	waf, ok = e.space.Get(request.String())
	if !ok {
		waf, err = rulesets.NewWAF(request.GetCoreRuleset())
		if err != nil {
			return nil, szerror.StatusInvalidArgumentF("core ruleset: %v", err)
		}
	}

	e.space.SetWithTTL(request.String(), waf, time.Hour)

	req := httpreq.NewRequestContext(ctx, request.GetRequestContext())

	eval, err := szcrs.Evaluate(waf, req)
	if err != nil {
		return nil, szerror.StatusInternalErrorF("evaluate: %v", err)
	}

	return &edgepb.EvalCoreRulesetResponse{Results: toResults(eval)}, nil
}

// toResults maps each matched rule to a result. Only the rule that
// interrupted the request carries an action.
func toResults(eval szcrs.Evaluation) []*edgepb.EvalResult {
	results := make([]*edgepb.EvalResult, 0, len(eval.Matched))

	for _, mr := range eval.Matched {
		result := &edgepb.EvalResult{
			RuleId:  strconv.Itoa(mr.Rule().ID()),
			Matched: true,
		}

		it := eval.Interruption
		if it != nil && it.RuleID == mr.Rule().ID() {
			result.Actions = []*rulepb.Action{{Type: actionType(it)}}
		}

		results = append(results, result)
	}

	return results
}

func actionType(it *types.Interruption) rulepb.ActionType {
	if it.Action == "redirect" {
		return rulepb.ActionType_ACTION_TYPE_REDIRECT
	}

	return rulepb.ActionType_ACTION_TYPE_BLOCK
}
