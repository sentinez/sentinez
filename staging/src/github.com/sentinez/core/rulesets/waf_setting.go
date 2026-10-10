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

package corers

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	crslang "github.com/coreruleset/crslang/types"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

// IDs of the rules generated from a setting, kept below the CRS range
// (900000-999999) so they never collide with a CRS rule.
const (
	_settingRuleID = 899000
	_scopedRuleID  = 899001
)

// ErrInvalidSetting a setting value cannot be mapped to a directive.
var ErrInvalidSetting = errors.New("invalid core ruleset setting")

// Setting values end up inside quoted directives, anything that could
// escape the quoting or expand a macro is rejected.
var (
	_methodRe      = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)
	_contentTypeRe = regexp.MustCompile(
		`^[A-Za-z0-9][A-Za-z0-9!#$&^_.+-]*/[A-Za-z0-9][A-Za-z0-9!#$&^_.+-]*$`,
	)
	_tagRe      = regexp.MustCompile(`^[A-Za-z0-9_./-]+$`)
	_variableRe = regexp.MustCompile(`^[A-Z_]+(:[^\s"'\\,;|]+)?$`)
)

// nolint:lll
var _categoryFlags = map[corerulesetpb.Category]Flag{
	corerulesetpb.Category_CATEGORY_COMMON_EXCEPTIONS:    ReqCommonExceptions,
	corerulesetpb.Category_CATEGORY_METHOD_ENFORCEMENT:   ReqMethodEnforcement,
	corerulesetpb.Category_CATEGORY_SCANNER_DETECTION:    ReqScannerDetection,
	corerulesetpb.Category_CATEGORY_PROTOCOL_ENFORCEMENT: ReqProtocolEnforcement,
	corerulesetpb.Category_CATEGORY_PROTOCOL_ATTACK:      ReqProtocolAttack,
	corerulesetpb.Category_CATEGORY_MULTIPART_ATTACK:     ReqMultipartAttack,
	corerulesetpb.Category_CATEGORY_LFI:                  ReqAppAttackLFI,
	corerulesetpb.Category_CATEGORY_RFI:                  ReqAppAttackRFI,
	corerulesetpb.Category_CATEGORY_RCE:                  ReqAppAttackRCE,
	corerulesetpb.Category_CATEGORY_PHP:                  ReqAppAttackPHP,
	corerulesetpb.Category_CATEGORY_GENERIC:              ReqAppAttackGeneric,
	corerulesetpb.Category_CATEGORY_XSS:                  ReqAppAttackXSS,
	corerulesetpb.Category_CATEGORY_SQLI:                 ReqAppAttackSQLI,
	corerulesetpb.Category_CATEGORY_SESSION_FIXATION:     ReqAppAttackSessionFixation,
	corerulesetpb.Category_CATEGORY_JAVA:                 ReqAppAttackJava,
	corerulesetpb.Category_CATEGORY_DATA_LEAKAGES:        RespDataLeakages,
	corerulesetpb.Category_CATEGORY_DATA_LEAKAGES_SQL:    RespDataLeakagesSQL,
	corerulesetpb.Category_CATEGORY_DATA_LEAKAGES_JAVA:   RespDataLeakagesJava,
	corerulesetpb.Category_CATEGORY_DATA_LEAKAGES_PHP:    RespDataLeakagesPHP,
	corerulesetpb.Category_CATEGORY_DATA_LEAKAGES_IIS:    RespDataLeakagesIIS,
	corerulesetpb.Category_CATEGORY_WEB_SHELLS:           RespWebShells,
	corerulesetpb.Category_CATEGORY_DATA_LEAKAGES_RUBY:   RespDataLeakagesRuby,
}

// VersionOf maps the setting version, unspecified = latest supported.
func VersionOf(v corerulesetpb.Version) Version {
	if v == corerulesetpb.Version_VERSION_V4_16_0 {
		return WAF4160
	}

	return WAF4170
}

// FlagOf maps the enabled categories to rule group flags, empty = all.
func FlagOf(categories []corerulesetpb.Category) Flag {
	var flag Flag

	for _, c := range categories {
		flag |= _categoryFlags[c]
	}

	if flag == 0 {
		return AllCRS
	}

	return flag
}

// settingEnabled reports whether the setting turns the rule engine on.
func settingEnabled(s *corerulesetpb.CoreRuleset) bool {
	if s.GetStatus() != typepb.Status_STATUS_ACTIVE {
		return false
	}

	switch s.GetMode() {
	case corerulesetpb.EngineMode_ENGINE_MODE_ON,
		corerulesetpb.EngineMode_ENGINE_MODE_DETECTION_ONLY:
		return true
	}

	return false
}

// blockingLevel tx.blocking_paranoia_level, 0 = default (1).
func blockingLevel(s *corerulesetpb.CoreRuleset) uint32 {
	return max(s.GetParanoiaLevel(), 1)
}

// ruleStates global (not expression scoped) rule state overrides by rule
// ID.
func ruleStates(
	s *corerulesetpb.CoreRuleset,
) map[uint32]corerulesetpb.RuleState {
	states := make(map[uint32]corerulesetpb.RuleState, len(s.GetOverrides()))

	for _, o := range s.GetOverrides() {
		if !hasExpr(o.GetExpr()) {
			states[o.GetId()] = o.GetState()
		}
	}

	return states
}

func validateSetting(s *corerulesetpb.CoreRuleset) error {
	if err := validateRequest(s.GetRequest()); err != nil {
		return err
	}

	for _, e := range s.GetExclusions() {
		if err := validateExclusion(e); err != nil {
			return err
		}
	}

	for _, o := range s.GetOverrides() {
		if err := validateOverride(o); err != nil {
			return err
		}
	}

	return nil
}

func validateRequest(r *corerulesetpb.RequestPolicy) error {
	for _, m := range r.GetAllowedMethods() {
		if !_methodRe.MatchString(m) {
			return fmt.Errorf("%w: method %q", ErrInvalidSetting, m)
		}
	}

	for _, ct := range r.GetAllowedContentTypes() {
		if !_contentTypeRe.MatchString(ct) {
			return fmt.Errorf("%w: content type %q", ErrInvalidSetting, ct)
		}
	}

	return nil
}

func validateExclusion(e *corerulesetpb.Exclusion) error {
	for _, tag := range e.GetTags() {
		if !_tagRe.MatchString(tag) {
			return fmt.Errorf("%w: tag %q", ErrInvalidSetting, tag)
		}
	}

	for _, t := range e.GetTargets() {
		if !_variableRe.MatchString(t.GetVariable()) {
			return fmt.Errorf(
				"%w: variable %q", ErrInvalidSetting, t.GetVariable(),
			)
		}
	}

	_, err := exprChains(e.GetExpr())

	return err
}

// validateOverride only removing a rule can be scoped to an expression,
// the other states rewrite the rule itself when the rulesets are loaded.
func validateOverride(o *corerulesetpb.RuleOverride) error {
	if !hasExpr(o.GetExpr()) {
		return nil
	}

	switch o.GetState() {
	case corerulesetpb.RuleState_RULE_STATE_DETECTION_ONLY,
		corerulesetpb.RuleState_RULE_STATE_ENABLED:
		return fmt.Errorf(
			"%w: rule %d state %s cannot be scoped to an expression",
			ErrInvalidSetting, o.GetId(), o.GetState(),
		)
	}

	_, err := exprChains(o.GetExpr())

	return err
}

// setupDirectives maps the setting to the directives overriding setup.conf
// plus the exclusions and overrides scoped to an expression. They are
// loaded after the setup rules and before REQUEST-901-INITIALIZATION.
func setupDirectives(s *corerulesetpb.CoreRuleset) string {
	var sb strings.Builder

	engine := "On"
	if s.GetMode() == corerulesetpb.EngineMode_ENGINE_MODE_DETECTION_ONLY {
		engine = "DetectionOnly"
	}

	fmt.Fprintf(&sb, "SecRuleEngine %s\n", engine)
	writeRequestPolicy(&sb, s.GetRequest())
	fmt.Fprintf(&sb,
		"SecAction \"id:%d,phase:1,pass,t:none,nolog,%s\"\n",
		_settingRuleID, strings.Join(setupVars(s), ","),
	)
	writeScopedRules(&sb, s)

	return sb.String()
}

func writeRequestPolicy(sb *strings.Builder, r *corerulesetpb.RequestPolicy) {
	if r == nil {
		return
	}

	access := "Off"
	if r.GetBodyAccess() {
		access = "On"
	}

	fmt.Fprintf(sb, "SecRequestBodyAccess %s\n", access)

	if limit := r.GetBodyLimitBytes(); limit > 0 {
		fmt.Fprintf(sb, "SecRequestBodyLimit %d\n", limit)
	}
}

// setupVars setvar actions overriding the CRS setup variables.
func setupVars(s *corerulesetpb.CoreRuleset) []string {
	level := blockingLevel(s)
	vars := []string{
		fmt.Sprintf("setvar:tx.blocking_paranoia_level=%d", level),
		fmt.Sprintf(
			"setvar:tx.detection_paranoia_level=%d",
			max(s.GetDetectionParanoiaLevel(), level),
		),
	}

	if t := s.GetThreshold(); t != nil {
		vars = append(vars, thresholdVars(t)...)
	}

	if m := s.GetRequest().GetAllowedMethods(); len(m) > 0 {
		vars = append(vars, fmt.Sprintf(
			"setvar:'tx.allowed_methods=%s'", strings.Join(m, " "),
		))
	}

	if ct := s.GetRequest().GetAllowedContentTypes(); len(ct) > 0 {
		vars = append(vars, fmt.Sprintf(
			"setvar:'tx.allowed_request_content_type=|%s|'",
			strings.Join(ct, "| |"),
		))
	}

	return vars
}

// thresholdVars 0 keeps the threshold of setup.conf.
func thresholdVars(t *corerulesetpb.AnomalyThreshold) []string {
	early := 0
	if t.GetEarlyBlocking() {
		early = 1
	}

	vars := []string{fmt.Sprintf("setvar:tx.early_blocking=%d", early)}

	if in := t.GetInbound(); in > 0 {
		vars = append(vars, fmt.Sprintf(
			"setvar:tx.inbound_anomaly_score_threshold=%d", in,
		))
	}

	if out := t.GetOutbound(); out > 0 {
		vars = append(vars, fmt.Sprintf(
			"setvar:tx.outbound_anomaly_score_threshold=%d", out,
		))
	}

	return vars
}

// writeScopedRules exclusions and overrides scoped to an expression are
// applied per transaction with ctl actions.
func writeScopedRules(sb *strings.Builder, s *corerulesetpb.CoreRuleset) {
	id := _scopedRuleID

	write := func(expr *rulepb.Expression, ctls []string) {
		if !hasExpr(expr) {
			return
		}

		for _, rule := range scopedRules(id, expr, ctls) {
			sb.WriteString(rule.ToSeclang())
			id++
		}
	}

	for _, e := range s.GetExclusions() {
		write(e.GetExpr(), exclusionCtls(e))
	}

	for _, o := range s.GetOverrides() {
		if o.GetState() != corerulesetpb.RuleState_RULE_STATE_DISABLED {
			continue
		}

		write(o.GetExpr(), []string{
			fmt.Sprintf("ruleRemoveById=%d", o.GetId()),
		})
	}
}

// scopedRules the rules applying the ctls to the transactions matching
// expr, IDs starting at id. The expression is translated to SecLang: one
// chain per conjunction of its conditions.
func scopedRules(
	id int, expr *rulepb.Expression, ctls []string,
) []crslang.ChainableDirective {
	// The setting is validated before any directive is generated.
	chains, err := exprChains(expr)
	if err != nil || len(ctls) == 0 {
		return nil
	}

	rules := make([]crslang.ChainableDirective, 0, len(chains))
	for i, chain := range chains {
		rules = append(rules, chainRule(id+i, chain, ctls))
	}

	return rules
}

// chainRule chains the links into one rule running the ctls when they
// all match. The ctls sit on the last link: the actions of the others
// run as soon as their own link matches.
func chainRule(
	id int, links []*crslang.SecRule, ctls []string,
) crslang.ChainableDirective {
	apply := make([]crslang.Action, 0, len(ctls))
	for _, ctl := range ctls {
		apply = append(apply, crslang.ActionWithParam{"ctl": ctl})
	}

	if len(links) == 0 {
		action := crslang.NewSecAction()
		action.Transformations.Transformations = []crslang.Transformation{
			crslang.None,
		}
		setHead(action.Metadata, action.Actions, id)
		action.Actions.NonDisruptiveActions = append(
			action.Actions.NonDisruptiveActions, apply...,
		)

		return action
	}

	for i, link := range links[:len(links)-1] {
		link.Actions.FlowActions = []crslang.Action{
			crslang.ActionOnly(crslang.Chain.String()),
		}
		link.ChainedRule = links[i+1]
	}

	setHead(links[0].Metadata, links[0].Actions, id)

	last := links[len(links)-1].Actions
	last.NonDisruptiveActions = append(last.NonDisruptiveActions, apply...)

	return links[0]
}

// setHead ID, phase and actions of the first rule of a chain.
func setHead(
	meta *crslang.SecRuleMetadata, actions *crslang.SeclangActions, id int,
) {
	meta.Id = id
	meta.Phase = "1"
	actions.DisruptiveAction = crslang.ActionOnly(crslang.Pass.String())
	actions.NonDisruptiveActions = []crslang.Action{
		crslang.ActionOnly(crslang.NoLog.String()),
	}
}

// hasExpr reports whether the expression scopes a setting to the requests
// matching it, empty = every request.
func hasExpr(expr *rulepb.Expression) bool {
	return len(expr.GetOrCondition()) > 0
}

// exclusionCtls parameters of the ctl actions applying the exclusion.
func exclusionCtls(e *corerulesetpb.Exclusion) []string {
	ctls := make([]string, 0, len(e.GetTags())+len(e.GetTargets()))

	for _, tag := range e.GetTags() {
		ctls = append(ctls, "ruleRemoveByTag="+tag)
	}

	for _, t := range e.GetTargets() {
		ctls = append(ctls, fmt.Sprintf(
			"ruleRemoveTargetById=%d;%s", t.GetRuleId(), t.GetVariable(),
		))
	}

	return ctls
}

// globalDirectives maps the exclusions and overrides without an
// expression to configure-time directives. They change rules already
// defined, so they are loaded after every ruleset; loaded reports whether
// a rule ID is part of the loaded rulesets.
func globalDirectives(
	s *corerulesetpb.CoreRuleset, loaded func(id uint32) bool,
) string {
	var sb strings.Builder

	for _, e := range s.GetExclusions() {
		if hasExpr(e.GetExpr()) {
			continue
		}

		for _, tag := range e.GetTags() {
			fmt.Fprintf(&sb, "SecRuleRemoveByTag %s\n", tag)
		}

		for _, t := range e.GetTargets() {
			// Updating a rule that is not loaded is a parse error.
			if !loaded(t.GetRuleId()) {
				continue
			}

			fmt.Fprintf(&sb, "SecRuleUpdateTargetById %d \"!%s\"\n",
				t.GetRuleId(), t.GetVariable(),
			)
		}
	}

	for _, o := range s.GetOverrides() {
		if !hasExpr(o.GetExpr()) &&
			o.GetState() == corerulesetpb.RuleState_RULE_STATE_DISABLED {
			fmt.Fprintf(&sb, "SecRuleRemoveById %d\n", o.GetId())
		}
	}

	return sb.String()
}
