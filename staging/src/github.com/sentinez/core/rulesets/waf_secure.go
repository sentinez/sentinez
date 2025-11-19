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

package corers

import (
	"sync"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/types"
	corehttp "github.com/sentinez/core/http"
)

var (
	poolWAF = sync.Pool{
		New: func() any {
			return &Rulesets{}
		},
	}
)

func NewRulesets(ctx corehttp.Context, waf coraza.WAF) *Rulesets {
	if waf == nil {
		return nil
	}

	rule := poolWAF.Get().(*Rulesets)
	rule.tx = newTransaction(waf, ctx)

	return rule
}

type Rulesets struct {
	tx types.Transaction
}

func (rs *Rulesets) ExecIngress(ctx corehttp.Context) error {
	if rs == nil {
		return nil
	}

	if err := processRequestHandler(ctx, rs.tx); err != nil {
		debugLogger(rs.tx, err, "faild to process request")
		return err
	}

	return nil
}

func (rs *Rulesets) ExecEgress(ctx corehttp.Context) error {
	if rs == nil {
		return nil
	}

	if err := processResponseHandler(ctx, rs.tx); err != nil {
		debugLogger(rs.tx, err, "faild to process response")
		return err
	}

	return nil
}

func (rs *Rulesets) Final(callback func()) {
	if rs == nil {
		return
	}

	// final phase
	rs.tx.ProcessLogging()

	callback()

	if err := rs.tx.Close(); err != nil {
		debugLogger(rs.tx, err, "failed to close transaction")
	}
}

func (rs *Rulesets) Matched() (*types.Interruption, []types.MatchedRule, bool) {
	if rs == nil {
		return nil, nil, false
	}

	if !rs.tx.IsInterrupted() {
		return nil, nil, false
	}

	return rs.tx.Interruption(), rs.tx.MatchedRules(), true
}

func (rs *Rulesets) GetTxId() string {
	if rs == nil {
		return ""
	}

	return rs.tx.ID()
}

func (rs *Rulesets) Release() {
	if rs == nil {
		return
	}

	rs.tx = nil
	poolWAF.Put(rs)
}

func (rs *Rulesets) IsRuleEngineOff() bool {
	if rs == nil {
		return false
	}

	return rs.tx.IsRuleEngineOff()
}
