package modsecx

import (
	"github.com/sentinez/sentinez/modsecx/ruleloader"

	rules "github.com/sentinez/sentinez/modsecx/gen"
	rulev4160 "github.com/sentinez/sentinez/modsecx/gen/v4-16-0"
)

func Load(version CRSVersion, rulesetsFlag CRSFlag) string {
	var rulesets = &ruleloader.RuleLoader{}

	// load setup rules
	rulesets.Load(rules.SetupOrder)

	// load init rule
	rulesets.Load(rules.Request901InitializationOrder)

	// load core rulesets
	switch version {
	case CRSv4160:
		r4160(rulesets, rulesetsFlag)
	}

	//load extension rules
	rulesets.Load(rules.AuditOrder)
	rulesets.Load(rules.DefaultOrder)

	// load evaluation rules
	rulesets.Load(rules.Request949BlockingEvaluationOrder)

	return rulesets.Export()
}

func r4160(rulesets *ruleloader.RuleLoader, flag CRSFlag) {

	if flag&ReqAppAttackRCE != 0 {
		rulesets.Load(rulev4160.Request932ApplicationAttackRceOrder)
	}

	if flag&ReqAppAttackSQLI != 0 {
		rulesets.Load(rulev4160.Request942ApplicationAttackSqliOrder)
	}

}
