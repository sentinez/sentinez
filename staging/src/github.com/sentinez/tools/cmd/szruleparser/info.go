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

package main

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/sentinez/core/szmodsec/ruleparser"
	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
)

const _paranoiaTag = "paranoia-level/"

// _fileNumber matches the CRS file number, ex: REQUEST-932-... -> 932.
var _fileNumber = regexp.MustCompile(`^(?:REQUEST|RESPONSE)[-_](\d{3})[-_]`)

// _categories maps the CRS file number to its rule group. Files that are
// not listed (901, 949, 959, 980, setup, ...) only hold system rules.
var _categories = map[string]rulesetpb.Category{
	"905": rulesetpb.Category_CATEGORY_COMMON_EXCEPTIONS,
	"911": rulesetpb.Category_CATEGORY_METHOD_ENFORCEMENT,
	"913": rulesetpb.Category_CATEGORY_SCANNER_DETECTION,
	"920": rulesetpb.Category_CATEGORY_PROTOCOL_ENFORCEMENT,
	"921": rulesetpb.Category_CATEGORY_PROTOCOL_ATTACK,
	"922": rulesetpb.Category_CATEGORY_MULTIPART_ATTACK,
	"930": rulesetpb.Category_CATEGORY_LFI,
	"931": rulesetpb.Category_CATEGORY_RFI,
	"932": rulesetpb.Category_CATEGORY_RCE,
	"933": rulesetpb.Category_CATEGORY_PHP,
	"934": rulesetpb.Category_CATEGORY_GENERIC,
	"941": rulesetpb.Category_CATEGORY_XSS,
	"942": rulesetpb.Category_CATEGORY_SQLI,
	"943": rulesetpb.Category_CATEGORY_SESSION_FIXATION,
	"944": rulesetpb.Category_CATEGORY_JAVA,
	"950": rulesetpb.Category_CATEGORY_DATA_LEAKAGES,
	"951": rulesetpb.Category_CATEGORY_DATA_LEAKAGES_SQL,
	"952": rulesetpb.Category_CATEGORY_DATA_LEAKAGES_JAVA,
	"953": rulesetpb.Category_CATEGORY_DATA_LEAKAGES_PHP,
	"954": rulesetpb.Category_CATEGORY_DATA_LEAKAGES_IIS,
	"955": rulesetpb.Category_CATEGORY_WEB_SHELLS,
	"956": rulesetpb.Category_CATEGORY_DATA_LEAKAGES_RUBY,
}

func categoryOf(filePath string) rulesetpb.Category {
	match := _fileNumber.FindStringSubmatch(normalizeFineName(filePath))
	if match == nil {
		return rulesetpb.Category_CATEGORY_UNSPECIFIED
	}
	return _categories[match[1]]
}

// ruleInfos builds the catalog of the rules that have an ID.
func ruleInfos(filePath string, rules []ruleparser.Rule) []*rulesetpb.RuleInfo {
	category := categoryOf(filePath)

	infos := make([]*rulesetpb.RuleInfo, 0, len(rules))
	for _, rule := range rules {
		if info := ruleInfo(category, rule); info != nil {
			infos = append(infos, info)
		}
	}
	return infos
}

func ruleInfo(
	category rulesetpb.Category, rule ruleparser.Rule,
) *rulesetpb.RuleInfo {
	if rule.Actions == nil {
		return nil
	}
	fields := rule.Actions.Fields

	id, err := strconv.ParseUint(first(fields["id"]), 10, 32)
	if err != nil {
		return nil
	}

	level, _ := strconv.ParseUint(
		strings.TrimPrefix(rule.Level, _paranoiaTag), 10, 32)
	msg := first(fields["msg"])

	return &rulesetpb.RuleInfo{
		Id:            uint32(id),
		Category:      category,
		ParanoiaLevel: uint32(level),
		Severity:      first(fields["severity"]),
		Msg:           msg,
		Tags:          fields["tag"],
		// Rules without a message never report a match on their own:
		// they are flow control or helpers that other rules depend on.
		System: category == rulesetpb.Category_CATEGORY_UNSPECIFIED ||
			msg == "",
	}
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
