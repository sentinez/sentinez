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

package szcrs

import (
	"strconv"
	"sync"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/collection"
	"github.com/corazawaf/coraza/v3/experimental/plugins/plugintypes"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/sentinez/core/szhttp"
)

// TX variables CRS sums the anomaly scores of the matched rules in.
const (
	_txInboundScore  = "blocking_inbound_anomaly_score"
	_txOutboundScore = "blocking_outbound_anomaly_score"
)

var (
	_poolWAF = sync.Pool{
		New: func() any {
			return &Rulesets{}
		},
	}
)

func NewRulesets(ctx szhttp.Context, waf coraza.WAF) *Rulesets {
	if waf == nil {
		return nil
	}

	rule, ok := _poolWAF.Get().(*Rulesets)
	if !ok {
		rule = &Rulesets{}
	}
	rule.tx = newTransaction(waf, ctx)

	return rule
}

type Rulesets struct {
	tx types.Transaction
}

func (rs *Rulesets) ExecIngress(ctx szhttp.Context) error {
	if rs == nil {
		return nil
	}

	if err := processRequestHandler(ctx, rs.tx); err != nil {
		debugLogger(rs.tx, err, "faild to process request")
		return err
	}

	return nil
}

func (rs *Rulesets) ExecEgress(ctx szhttp.Context) error {
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

// AnomalyScore returns the inbound and outbound anomaly scores CRS sums up
// in its blocking evaluation (REQUEST-949, RESPONSE-959) of the transaction.
func (rs *Rulesets) AnomalyScore() (inbound, outbound int) {
	if rs == nil {
		return 0, 0
	}

	state, ok := rs.tx.(plugintypes.TransactionState)
	if !ok {
		return 0, 0
	}

	vars := state.Variables().TX()

	return txInt(vars, _txInboundScore), txInt(vars, _txOutboundScore)
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
	_poolWAF.Put(rs)
}

func (rs *Rulesets) IsRuleEngineOff() bool {
	if rs == nil {
		return false
	}

	return rs.tx.IsRuleEngineOff()
}

// IsFrameworkRule reports whether id is a CRS rule that detects nothing:
// the setup and initialization rules, and the blocking evaluation and
// correlation rules concluding on the anomaly score.
func IsFrameworkRule(id int) bool {
	switch id / 1000 {
	case 900, 901, 949, 959, 980:
		return true
	default:
		return false
	}
}

func txInt(vars collection.Map, key string) int {
	values := vars.Get(key)
	if len(values) == 0 {
		return 0
	}

	n, _ := strconv.Atoi(values[0])

	return n
}
