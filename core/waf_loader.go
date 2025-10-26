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

package core

import (
	rules "github.com/sentinez/sentinez/core/modsec/gen"
	rulev4160 "github.com/sentinez/sentinez/core/modsec/gen/v4-16-0"
)

func GetWAFRule(version WAFVersion, flag WAFFlag) string {
	var rulesets = &RuleLoader{}

	// load setup rules
	rulesets.Load(rules.SetupOrder)

	// load init rule
	rulesets.Load(rules.Request901InitializationOrder)

	// load core rulesets
	switch version {
	case WAF4160:
		r4160(rulesets, flag)
	}

	//load extension rules
	rulesets.Load(rules.AuditOrder)
	rulesets.Load(rules.DefaultOrder)

	// load evaluation rules
	rulesets.Load(rules.Request949BlockingEvaluationOrder)

	return rulesets.Export()
}

func r4160(rulesets *RuleLoader, flag WAFFlag) {

	if flag&ReqAppAttackRCE != 0 {
		rulesets.Load(rulev4160.Request932ApplicationAttackRceOrder)
	}

	if flag&ReqAppAttackSQLI != 0 {
		rulesets.Load(rulev4160.Request942ApplicationAttackSqliOrder)
	}

}
