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

package szruleset

import (
	"strconv"
	"sync"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/experimental/plugins/plugintypes"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/sentinez/core/szhttp"
)

var (
	_poolWAF = sync.Pool{
		New: func() any {
			return &Rulesets{}
		},
	}
)

// NewRulesets starts a transaction of waf for the request, nil if waf is
// nil. Every method is a no-op on a nil Rulesets.
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

// Matched returns the rules the transaction matched, the interruption
// and whether the engine blocked the transaction. In DetectionOnly mode
// nothing is blocked, the interruption is the one the engine would have
// performed were it On (nil = none).
func (rs *Rulesets) Matched() (
	matched []types.MatchedRule, it *types.Interruption, blocked bool,
) {
	if rs == nil {
		return nil, nil, false
	}

	if rs.tx.IsInterrupted() {
		return rs.tx.MatchedRules(), rs.tx.Interruption(), true
	}

	return rs.tx.MatchedRules(), detectionOnlyInterruption(rs.tx), false
}

// TxInt returns the integer value of the TX variable key of the
// transaction, 0 if unset or not a number.
func (rs *Rulesets) TxInt(key string) int {
	if rs == nil {
		return 0
	}

	state, ok := rs.tx.(plugintypes.TransactionState)
	if !ok {
		return 0
	}

	values := state.Variables().TX().Get(key)
	if len(values) == 0 {
		return 0
	}

	n, _ := strconv.Atoi(values[0])

	return n
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

// detectionOnlyInterruption the interruption a DetectionOnly transaction
// would have performed, nil if none. Coraza does not expose it in
// types.Transaction yet.
func detectionOnlyInterruption(tx types.Transaction) *types.Interruption {
	d, ok := tx.(interface {
		DetectionOnlyInterruption() *types.Interruption
	})
	if !ok {
		return nil
	}

	return d.DetectionOnlyInterruption()
}
