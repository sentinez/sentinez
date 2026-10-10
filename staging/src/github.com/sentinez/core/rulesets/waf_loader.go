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
	"errors"
	"fmt"

	rules "github.com/sentinez/core/modsec/gen"
	rulev4160 "github.com/sentinez/core/modsec/gen/v4-16-0"
	rulev4170 "github.com/sentinez/core/modsec/gen/v4-17-0"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
)

// ErrUnsupportedVersion no rulesets are generated for the CRS version.
var ErrUnsupportedVersion = errors.New("unsupported core ruleset version")

// ruleGroup a CRS rule file, loaded when its flag is enabled.
type ruleGroup struct {
	flag  Flag // 0 = always loaded
	order []func() *corerulesetpb.CoreRule
}

// versionGroups rule files of a CRS version, in load order.
type versionGroups struct {
	request  []ruleGroup
	response []ruleGroup
}

var _versionGroups = map[Version]versionGroups{
	WAF4160: {
		request: []ruleGroup{
			{ReqAppAttackRCE, rulev4160.Request932ApplicationAttackRceOrder},
			{ReqAppAttackSQLI, rulev4160.Request942ApplicationAttackSqliOrder},
		},
	},
	WAF4170: {
		request: []ruleGroup{
			{ReqCommonExceptions, rulev4170.Request905CommonExceptionsOrder},
			{ReqMethodEnforcement, rulev4170.Request911MethodEnforcementOrder},
			{ReqScannerDetection, rulev4170.Request913ScannerDetectionOrder},
			{
				ReqProtocolEnforcement,
				rulev4170.Request920ProtocolEnforcementOrder,
			},
			{ReqProtocolAttack, rulev4170.Request921ProtocolAttackOrder},
			{ReqMultipartAttack, rulev4170.Request922MultipartAttackOrder},
			{ReqAppAttackLFI, rulev4170.Request930ApplicationAttackLfiOrder},
			{ReqAppAttackRFI, rulev4170.Request931ApplicationAttackRfiOrder},
			{ReqAppAttackRCE, rulev4170.Request932ApplicationAttackRceOrder},
			{ReqAppAttackPHP, rulev4170.Request933ApplicationAttackPhpOrder},
			{
				ReqAppAttackGeneric,
				rulev4170.Request934ApplicationAttackGenericOrder,
			},
			{ReqAppAttackXSS, rulev4170.Request941ApplicationAttackXssOrder},
			{ReqAppAttackSQLI, rulev4170.Request942ApplicationAttackSqliOrder},
			{
				ReqAppAttackSessionFixation,
				rulev4170.Request943ApplicationAttackSessionFixationOrder,
			},
			{ReqAppAttackJava, rulev4170.Request944ApplicationAttackJavaOrder},
		},
		response: []ruleGroup{
			{RespDataLeakages, rulev4170.Response950DataLeakagesOrder},
			{RespDataLeakagesSQL, rulev4170.Response951DataLeakagesSqlOrder},
			{RespDataLeakagesJava, rulev4170.Response952DataLeakagesJavaOrder},
			{RespDataLeakagesPHP, rulev4170.Response953DataLeakagesPhpOrder},
			{RespDataLeakagesIIS, rulev4170.Response954DataLeakagesIisOrder},
			{RespWebShells, rulev4170.Response955WebShellsOrder},
			{RespDataLeakagesRuby, rulev4170.Response956DataLeakagesRubyOrder},
			{0, rulev4170.Response959BlockingEvaluationOrder},
			{0, rulev4170.Response980CorrelationOrder},
		},
	},
}

// GenerateRulesets maps the setting to the rule content (SecLang
// directives) loaded into coraza. A setting that is not active, or with
// the engine off, maps to a rule engine without any ruleset.
func GenerateRulesets(
	version Version, setting *corerulesetpb.CoreRuleset,
) (string, error) {
	if !settingEnabled(setting) {
		return "SecRuleEngine Off\n", nil
	}

	groups, ok := _versionGroups[version]
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrUnsupportedVersion, version)
	}

	if err := validateSetting(setting); err != nil {
		return "", err
	}

	rulesets := NewRulesetsLoader(setting)
	loadRulesets(rulesets, groups, setting)

	return rulesets.Export(), nil
}

// loadRulesets the load order matters: the setting overrides the setup
// rules before the CRS reads them, global exclusions and overrides change
// rules that must already be defined.
func loadRulesets(
	rulesets *RulesetsLoader,
	groups versionGroups,
	setting *corerulesetpb.CoreRuleset,
) {
	flag := FlagOf(setting.GetCategories())

	// load setup rules, then the setting on top of them
	rulesets.Load(rules.SetupOrder)
	rulesets.Write(setupDirectives(setting))

	// load init rule
	rulesets.Load(rules.Request901InitializationOrder)

	// load core rulesets
	loadGroups(rulesets, groups.request, flag)

	// load extension rules
	rulesets.Load(rules.AuditOrder)
	rulesets.Load(rules.DefaultOrder)

	// load evaluation rules
	rulesets.Load(rules.Request949BlockingEvaluationOrder)

	loadGroups(rulesets, groups.response, flag)

	// load global exclusions and overrides
	rulesets.Write(globalDirectives(setting, rulesets.Loaded))
}

func loadGroups(rulesets *RulesetsLoader, groups []ruleGroup, flag Flag) {
	for _, g := range groups {
		if g.flag == 0 || flag&(AllCRS|g.flag) != 0 {
			rulesets.Load(g.order)
		}
	}
}
