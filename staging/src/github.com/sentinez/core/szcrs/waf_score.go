// Copyright 2026 Duc-Hung Ho.
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

import "github.com/sentinez/core/szruleset"

// TX variables CRS sums the anomaly scores of the matched rules in.
const (
	_txInboundScore  = "blocking_inbound_anomaly_score"
	_txOutboundScore = "blocking_outbound_anomaly_score"
)

// AnomalyScore returns the inbound and outbound anomaly scores CRS sums up
// in its blocking evaluation (REQUEST-949, RESPONSE-959) of the
// transaction.
func AnomalyScore(rs *szruleset.Rulesets) (inbound, outbound int) {
	return rs.TxInt(_txInboundScore), rs.TxInt(_txOutboundScore)
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
