// Copyright 2025 Sentinez Labs.
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

package httpxsecure

import (
	"bytes"
	"encoding/base64"
	"os"
	"sync"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/types"
	wafpb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/waf/v1"
	"github.com/sentinez/sentinez/pkg/auto/rules"
	rulev4160 "github.com/sentinez/sentinez/pkg/auto/rules/v4-16-0"
	"github.com/sentinez/sentinez/pkg/common/color"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var (
	waf  coraza.WAF
	lock sync.Mutex
)

func NewFireWall(ruleRoot string) coraza.WAF {
	lock.Lock()
	defer lock.Unlock()

	if waf == nil {
		var err error

		rootFS := os.DirFS(ruleRoot)
		conf := coraza.NewWAFConfig().WithRootFS(rootFS).
			WithErrorCallback(callback).
			WithDirectives(loadCoreRulesets())

		waf, err = coraza.NewWAF(conf)
		if err != nil {
			zlog.Errorf("failed to create WAF: %v", err)
			return nil
		}

		if waf != nil {
			zlog.Debugf("[%s] initialized successfully", color.Red.Add("WAF"))
		}
	}

	return waf
}

func loadCoreRulesets() string {
	var buf bytes.Buffer

	// load(&buf, rules.Default)

	// load setup rules
	load(&buf, rules.Setup)

	// load rules from v4.16.0
	load(&buf, rulev4160.Request901Initialization)

	// load core rulesets
	load(&buf, rulev4160.Request932ApplicationAttackRce)

	// load evaluation rules
	load(&buf, rulev4160.Request949BlockingEvaluation)

	return buf.String()
}

func load(buf *bytes.Buffer, rulesets map[string]*wafpb.Rule) {
	for _, rule := range rulesets {
		if rule == nil {
			continue
		}

		conf, err := base64.StdEncoding.DecodeString(rule.Configuration)
		if err != nil {
			continue
		}

		_, _ = buf.Write(conf)
		_, _ = buf.WriteString("\n\n")
	}
}

func callback(err types.MatchedRule) {
	msg := err.ErrorLog()
	zlog.Debugf("[%s] %s", err.Rule().Severity(), msg)
}
