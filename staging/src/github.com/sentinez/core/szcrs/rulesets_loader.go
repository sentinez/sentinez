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
	"regexp"
	"strconv"
	"strings"

	"github.com/sentinez/core/szruleset"
	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
)

const _levelTagPrefix = "paranoia-level/"

var (
	// anomaly score counters of a rule, ex: tx.inbound_anomaly_score_pl2
	_scoreVarRe   = regexp.MustCompile(`tx\.(in|out)bound_anomaly_score_pl`)
	_scoreLevelRe = regexp.MustCompile(`(bound_anomaly_score_pl)[2-4]`)
)

// newLoader returns a loader applying the rule overrides of the setting
// to the rules it loads.
func newLoader(crs *rulesetpb.Crs) *szruleset.Loader {
	return szruleset.NewLoader(ruleOverrides(crs))
}

// ruleOverrides rewrites the rules of a state override. Rules forced on
// above the blocking paranoia level are hoisted in front of the others,
// out of reach of the paranoia level skip rules of their file.
func ruleOverrides(crs *rulesetpb.Crs) szruleset.Rewrite {
	states := ruleStates(crs)
	level := blockingLevel(crs)

	return func(rule *rulesetpb.Rule, conf []byte) ([]byte, bool) {
		switch states[szruleset.RuleId(rule)] {
		case rulesetpb.RuleState_RULE_STATE_DETECTION_ONLY:
			return detectionOnly(conf), false
		case rulesetpb.RuleState_RULE_STATE_ENABLED:
			if ruleLevel(rule) > level {
				return forceEnabled(conf), true
			}
		}

		return conf, false
	}
}

// detectionOnly keeps the rule matching and logging but moves its score
// to counters the blocking evaluation never reads.
func detectionOnly(conf []byte) []byte {
	return _scoreVarRe.ReplaceAll(
		conf, []byte("tx.detection_only_${1}bound_score_pl"),
	)
}

// forceEnabled scores the rule at paranoia level 1, the only level the
// blocking evaluation always counts.
func forceEnabled(conf []byte) []byte {
	return _scoreLevelRe.ReplaceAll(conf, []byte("${1}1"))
}

// ruleLevel paranoia level of the rule, 0 = none.
func ruleLevel(rule *rulesetpb.Rule) uint32 {
	level, ok := strings.CutPrefix(rule.GetLevel(), _levelTagPrefix)
	if !ok {
		return 0
	}

	n, err := strconv.ParseUint(level, 10, 32)
	if err != nil {
		return 0
	}

	return uint32(n)
}
