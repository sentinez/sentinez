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

// Package catalog indexes the generated OWASP CRS rules by version, for
// the services that present them without running the rule engine.
package catalog

import (
	"sort"

	rulev4160 "github.com/sentinez/core/modsec/gen/v4-16-0"
	rulev4170 "github.com/sentinez/core/modsec/gen/v4-17-0"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
)

// version the rules of a CRS version, sorted by ID.
type version struct {
	rules []*corerulesetpb.RuleInfo
	byID  map[uint32]*corerulesetpb.RuleInfo
}

// nolint:lll
var _versions = map[corerulesetpb.Version]*version{
	corerulesetpb.Version_VERSION_V4_16_0: newVersion(
		rulev4160.Request901InitializationInfo,
		rulev4160.Request905CommonExceptionsInfo,
		rulev4160.Request911MethodEnforcementInfo,
		rulev4160.Request913ScannerDetectionInfo,
		rulev4160.Request920ProtocolEnforcementInfo,
		rulev4160.Request921ProtocolAttackInfo,
		rulev4160.Request922MultipartAttackInfo,
		rulev4160.Request930ApplicationAttackLfiInfo,
		rulev4160.Request931ApplicationAttackRfiInfo,
		rulev4160.Request932ApplicationAttackRceInfo,
		rulev4160.Request933ApplicationAttackPhpInfo,
		rulev4160.Request934ApplicationAttackGenericInfo,
		rulev4160.Request941ApplicationAttackXssInfo,
		rulev4160.Request942ApplicationAttackSqliInfo,
		rulev4160.Request943ApplicationAttackSessionFixationInfo,
		rulev4160.Request944ApplicationAttackJavaInfo,
		rulev4160.Request949BlockingEvaluationInfo,
		rulev4160.Response950DataLeakagesInfo,
		rulev4160.Response951DataLeakagesSqlInfo,
		rulev4160.Response952DataLeakagesJavaInfo,
		rulev4160.Response953DataLeakagesPhpInfo,
		rulev4160.Response954DataLeakagesIisInfo,
		rulev4160.Response955WebShellsInfo,
		rulev4160.Response959BlockingEvaluationInfo,
		rulev4160.Response980CorrelationInfo,
	),
	corerulesetpb.Version_VERSION_V4_17_0: newVersion(
		rulev4170.Request901InitializationInfo,
		rulev4170.Request905CommonExceptionsInfo,
		rulev4170.Request911MethodEnforcementInfo,
		rulev4170.Request913ScannerDetectionInfo,
		rulev4170.Request920ProtocolEnforcementInfo,
		rulev4170.Request921ProtocolAttackInfo,
		rulev4170.Request922MultipartAttackInfo,
		rulev4170.Request930ApplicationAttackLfiInfo,
		rulev4170.Request931ApplicationAttackRfiInfo,
		rulev4170.Request932ApplicationAttackRceInfo,
		rulev4170.Request933ApplicationAttackPhpInfo,
		rulev4170.Request934ApplicationAttackGenericInfo,
		rulev4170.Request941ApplicationAttackXssInfo,
		rulev4170.Request942ApplicationAttackSqliInfo,
		rulev4170.Request943ApplicationAttackSessionFixationInfo,
		rulev4170.Request944ApplicationAttackJavaInfo,
		rulev4170.Request949BlockingEvaluationInfo,
		rulev4170.Response950DataLeakagesInfo,
		rulev4170.Response951DataLeakagesSqlInfo,
		rulev4170.Response952DataLeakagesJavaInfo,
		rulev4170.Response953DataLeakagesPhpInfo,
		rulev4170.Response954DataLeakagesIisInfo,
		rulev4170.Response955WebShellsInfo,
		rulev4170.Response956DataLeakagesRubyInfo,
		rulev4170.Response959BlockingEvaluationInfo,
		rulev4170.Response980CorrelationInfo,
	),
}

func newVersion(files ...[]*corerulesetpb.RuleInfo) *version {
	v := &version{byID: make(map[uint32]*corerulesetpb.RuleInfo)}

	for _, file := range files {
		for _, rule := range file {
			if _, ok := v.byID[rule.GetId()]; ok {
				continue
			}

			v.byID[rule.GetId()] = rule
			v.rules = append(v.rules, rule)
		}
	}

	sort.Slice(v.rules, func(i, j int) bool {
		return v.rules[i].GetId() < v.rules[j].GetId()
	})

	return v
}

// versionOf unspecified = latest supported, as the rule engine does.
func versionOf(v corerulesetpb.Version) *version {
	if found, ok := _versions[v]; ok {
		return found
	}

	return _versions[corerulesetpb.Version_VERSION_V4_17_0]
}

// Rules returns the rules of a CRS version sorted by ID. The result is
// shared and must not be modified.
func Rules(v corerulesetpb.Version) []*corerulesetpb.RuleInfo {
	return versionOf(v).rules
}

// Rule returns a rule of a CRS version by ID. The result is shared and
// must not be modified.
func Rule(
	v corerulesetpb.Version, id uint32,
) (*corerulesetpb.RuleInfo, bool) {
	rule, ok := versionOf(v).byID[id]
	return rule, ok
}
