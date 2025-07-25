package secrule

import (
	"bytes"
	"encoding/base64"
	"os"

	wafpb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/waf/v1"
	"github.com/sentinez/sentinez/pkg/auto/rules"
	rulev4160 "github.com/sentinez/sentinez/pkg/auto/rules/v4-16-0"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func Load() string {
	var rulesets = &RuleLoader{}

	// load setup rules
	rulesets.Load(rules.SetupOrder)

	// load rules from v4.16.0
	rulesets.Load(rulev4160.Request901InitializationOrder)

	// load core rulesets
	rulesets.Load(rulev4160.Request932ApplicationAttackRceOrder)
	rulesets.Load(rulev4160.Request942ApplicationAttackSqliOrder)

	// load extension rules
	rulesets.Load(rules.AuditOrder)
	rulesets.Load(rules.DefaultOrder)

	// load evaluation rules
	rulesets.Load(rulev4160.Request949BlockingEvaluationOrder)

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
			zlog.Errorf("failed to write rule configuration: %v", err)
			continue
		}
	}
}

func (rl *RuleLoader) Export() string {
	if rl.buf.Len() == 0 {
		return ""
	}

	if err := os.WriteFile("WAF.conf.lock", rl.buf.Bytes(), 0644); err != nil {
		zlog.Errorf("failed to write WAF configuration: %v", err)
		return ""
	}

	return rl.buf.String()
}
