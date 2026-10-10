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
	"bytes"
	"encoding/base64"
	"os"
	"regexp"
	"strconv"
	"strings"

	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
)

const _levelTagPrefix = "paranoia-level/"

var (
	// anomaly score counters of a rule, ex: tx.inbound_anomaly_score_pl2
	_scoreVarRe   = regexp.MustCompile(`tx\.(in|out)bound_anomaly_score_pl`)
	_scoreLevelRe = regexp.MustCompile(`(bound_anomaly_score_pl)[2-4]`)
)

type RulesetsLoader struct {
	buf bytes.Buffer

	// IDs of the rules loaded so far.
	ids map[uint32]struct{}
	// Rule state overrides applied while loading, key: rule ID.
	states map[uint32]corerulesetpb.RuleState
	// Blocking paranoia level, rules above it are skipped by the CRS.
	level uint32
}

// NewRulesetsLoader returns a loader applying the rule overrides of the
// setting to the rules it loads.
func NewRulesetsLoader(setting *corerulesetpb.CoreRuleset) *RulesetsLoader {
	return &RulesetsLoader{
		ids:    make(map[uint32]struct{}),
		states: ruleStates(setting),
		level:  blockingLevel(setting),
	}
}

// Load appends the rules in order. Rules forced on above the paranoia
// level are hoisted in front of the others, out of reach of the
// paranoia level skip rules of their file.
func (rl *RulesetsLoader) Load(rulesetsFn []func() *corerulesetpb.CoreRule) {
	rest := make([][]byte, 0, len(rulesetsFn))

	for _, fn := range rulesetsFn {
		rule := fn()

		conf, err := base64.StdEncoding.DecodeString(rule.GetConfiguration())
		if err != nil {
			continue
		}

		id := ruleID(rule)
		if id != 0 {
			rl.ids[id] = struct{}{}
		}

		switch rl.states[id] {
		case corerulesetpb.RuleState_RULE_STATE_DETECTION_ONLY:
			conf = detectionOnly(conf)
		case corerulesetpb.RuleState_RULE_STATE_ENABLED:
			if ruleLevel(rule) > rl.level {
				rl.write(forceEnabled(conf))
				continue
			}
		}

		rest = append(rest, conf)
	}

	for _, conf := range rest {
		rl.write(conf)
	}
}

// Write appends raw directives.
func (rl *RulesetsLoader) Write(directives string) {
	rl.write([]byte(directives))
}

// Loaded reports whether the rule ID has been loaded.
func (rl *RulesetsLoader) Loaded(id uint32) bool {
	_, ok := rl.ids[id]
	return ok
}

func (rl *RulesetsLoader) Export() string {
	if rl.buf.Len() == 0 {
		return ""
	}

	if err := os.WriteFile("WAF.conf.lock", rl.buf.Bytes(), 0644); err != nil {
		return ""
	}

	return rl.buf.String()
}

func (rl *RulesetsLoader) write(conf []byte) {
	if len(conf) == 0 {
		return
	}

	rl.buf.Write(conf)

	if !bytes.HasSuffix(conf, []byte("\n")) {
		rl.buf.WriteByte('\n')
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

// ruleID 0 = not a rule (marker, directive).
func ruleID(rule *corerulesetpb.CoreRule) uint32 {
	ids := rule.GetActions().GetFields().GetId()
	if len(ids) == 0 {
		return 0
	}

	id, err := strconv.ParseUint(ids[0], 10, 32)
	if err != nil {
		return 0
	}

	return uint32(id)
}

// ruleLevel paranoia level of the rule, 0 = none.
func ruleLevel(rule *corerulesetpb.CoreRule) uint32 {
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
