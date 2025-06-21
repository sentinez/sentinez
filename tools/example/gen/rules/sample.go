package rules

import (
	"github.com/sentinez/sentinez/api/gen/go/sentinez/edge/waf/v1"
)

var Crs = waf.CoreRulesets{
	Rules: []*waf.Rule{
		{
			Actions: &waf.RuleAction{
				Statement: "SecRuleREQBODY_PROCESSOR\"!@rx(?:URLENCODED|MULTIPART|XML|JSON)\"\"id:901340,phase:1,pass,nolog,noauditlog\"",
				Fields: &waf.RuleActionField{
					Id:    []string{"901340"},
					Phase: []string{"1"},
					Tag:   []string{"waf", "core-ruleset", "rule-901340"},
					Msg:   []string{"'RCE Bypass Technique'"},
				},
			},
			Configuration: "SecRuleREQBODY_PROCESSOR\"!@rx(?:URLENCODED|MULTIPART|XML|JSON)\"\"id:901340,phase:1,pass,nolog,noauditlog\"",
		},
	},
}
