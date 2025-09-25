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

package rules

import (
	"bytes"
	"encoding/base64"
	"os"

	wafpb "github.com/sentinez/sentinez/api/gen/go/sentinez/std/net/waf/v1"
	rules "github.com/sentinez/sentinez/rules/gen"
	rulev4160 "github.com/sentinez/sentinez/rules/gen/v4-16-0"
)

func Load() string {
	var rulesets = &RuleLoader{}

	// load setup rules
	rulesets.Load(rules.SetupOrder)

	// load rules from v4.16.0
	rulesets.Load(rules.Request901InitializationOrder)

	// load core rulesets
	rulesets.Load(rulev4160.Request932ApplicationAttackRceOrder)
	rulesets.Load(rulev4160.Request942ApplicationAttackSqliOrder)

	// load extension rules
	rulesets.Load(rules.AuditOrder)
	rulesets.Load(rules.DefaultOrder)

	// load evaluation rules
	rulesets.Load(rules.Request949BlockingEvaluationOrder)

	return rulesets.Export()
}

type RuleLoader struct {
	buf bytes.Buffer
}

func (rl *RuleLoader) Load(rulesetsFn []func() *wafpb.Rule) {
	for _, rule := range rulesetsFn {

		conf, err := base64.StdEncoding.DecodeString(rule().Configuration)
		if err != nil {
			continue
		}

		if !bytes.HasSuffix(conf, []byte("\n")) {
			conf = append(conf, '\n')
		}

		if _, err := rl.buf.Write(conf); err != nil {
			continue
		}
	}
}

func (rl *RuleLoader) Export() string {
	if rl.buf.Len() == 0 {
		return ""
	}

	if err := os.WriteFile("WAF.conf.lock", rl.buf.Bytes(), 0644); err != nil {
		return ""
	}

	return rl.buf.String()
}
