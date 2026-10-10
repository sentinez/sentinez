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
	"fmt"
	"net/netip"
	"regexp"
	"strings"

	crslang "github.com/coreruleset/crslang/types"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// Operands end up inside quoted directives, see the setting regexps.
var _fieldKeyRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// exprTarget what a condition is matched against: a variable, or the
// key of a collection.
type exprTarget struct {
	variable   crslang.VariableName
	collection crslang.CollectionName
	key        string
	count      bool // number of values instead of the values
}

// conjunctions flattens the expression to its disjunctive normal form:
// it matches when every condition of any conjunction matches. An empty
// expression is a single empty conjunction, matching every request.
func conjunctions(expr *rulepb.Expression) [][]*rulepb.Condition {
	return expandOr(expr.GetOrCondition())
}

func expandOr(or []*rulepb.AndCondition) [][]*rulepb.Condition {
	if len(or) == 0 {
		return [][]*rulepb.Condition{nil}
	}

	var conjs [][]*rulepb.Condition

	for _, and := range or {
		own := make([]*rulepb.Condition, 0, len(and.GetRules()))
		for _, r := range and.GetRules() {
			own = append(own, r.GetCondition())
		}

		for _, nested := range expandOr(and.GetOrCondition()) {
			conj := make([]*rulepb.Condition, 0, len(own)+len(nested))
			conjs = append(conjs, append(append(conj, own...), nested...))
		}
	}

	return conjs
}

// exprChains translates the expression to SecLang: one chain of rules
// per conjunction, a chain matching when all its rules match.
func exprChains(expr *rulepb.Expression) ([][]*crslang.SecRule, error) {
	conjs := conjunctions(expr)
	chains := make([][]*crslang.SecRule, 0, len(conjs))

	for _, conj := range conjs {
		var chain []*crslang.SecRule

		for _, c := range conj {
			links, err := conditionLinks(c)
			if err != nil {
				return nil, err
			}

			chain = append(chain, links...)
		}

		chains = append(chains, chain)
	}

	return chains, nil
}

// conditionLinks the rules that all match when the condition does.
func conditionLinks(c *rulepb.Condition) ([]*crslang.SecRule, error) {
	switch c.GetSource() {
	case rulepb.FieldSource_FIELD_SOURCE_PATH:
		return stringLink(
			exprTarget{variable: crslang.REQUEST_FILENAME}, c,
		)
	case rulepb.FieldSource_FIELD_SOURCE_HOST:
		return stringLink(exprTarget{variable: crslang.SERVER_NAME}, c)
	case rulepb.FieldSource_FIELD_SOURCE_METHOD:
		return valueLink(exprTarget{variable: crslang.REQUEST_METHOD}, c)
	case rulepb.FieldSource_FIELD_SOURCE_IP:
		return ipLink(c)
	case rulepb.FieldSource_FIELD_SOURCE_HEADER:
		return fieldLinks(crslang.REQUEST_HEADERS, "header", c)
	case rulepb.FieldSource_FIELD_SOURCE_QUERY:
		return fieldLinks(crslang.ARGS_GET, "query", c)
	}

	// The body, the JA4 fingerprint and the TLS state are not variables
	// of the rule engine in the request headers phase.
	return nil, fmt.Errorf(
		"%w: expression source %s", ErrInvalidSetting, c.GetSource(),
	)
}

// fieldLinks a header or a query argument: without a key the condition
// is on the presence of the fields named by the value. A rule has
// nothing to match on a missing field, so a negated comparison of its
// value (ne, not in) only matches when the field is present.
func fieldLinks(
	collection crslang.CollectionName, name string, c *rulepb.Condition,
) ([]*crslang.SecRule, error) {
	if key := c.GetKey(); key != "" && key != name {
		if !_fieldKeyRe.MatchString(key) {
			return nil, fmt.Errorf(
				"%w: expression key %q", ErrInvalidSetting, key,
			)
		}

		return valueLink(exprTarget{collection: collection, key: key}, c)
	}

	return presenceLinks(collection, c)
}

func presenceLinks(
	collection crslang.CollectionName, c *rulepb.Condition,
) ([]*crslang.SecRule, error) {
	keys, err := presenceKeys(c)
	if err != nil {
		return nil, err
	}

	// "@gt 0" = present, "@eq 0" = absent.
	op := crslang.Operator{Name: crslang.Gt, Value: "0"}
	if negated(c.GetOperator()) {
		op.Name = crslang.Eq
	}

	links := make([]*crslang.SecRule, 0, len(keys))
	for _, key := range keys {
		target := exprTarget{collection: collection, key: key, count: true}
		links = append(links, newLink(target, op))
	}

	return links, nil
}

// presenceKeys a list is only in (all present) or not in (all absent).
func presenceKeys(c *rulepb.Condition) ([]string, error) {
	keys := []string{c.GetValue().GetStringValue()}

	if list := c.GetValue().GetListValue(); list != nil {
		if !membership(c.GetOperator()) {
			return nil, operatorError(c)
		}

		keys = stringValues(list)
	}

	if len(keys) == 0 {
		return nil, operatorError(c)
	}

	for _, key := range keys {
		if !_fieldKeyRe.MatchString(key) {
			return nil, fmt.Errorf(
				"%w: expression key %q", ErrInvalidSetting, key,
			)
		}
	}

	return keys, nil
}

// valueLink a value compared to a string, or to the members of a list.
func valueLink(
	target exprTarget, c *rulepb.Condition,
) ([]*crslang.SecRule, error) {
	list := c.GetValue().GetListValue()
	if list == nil && !membership(c.GetOperator()) {
		return stringLink(target, c)
	}

	values := stringValues(list)
	if len(values) == 0 {
		return nil, operatorError(c)
	}

	quoted := make([]string, 0, len(values))
	for _, v := range values {
		if err := validateOperand(v); err != nil {
			return nil, err
		}

		quoted = append(quoted, regexp.QuoteMeta(v))
	}

	// \A and \z: the rx operator of the rule engine is multi-line.
	op := crslang.Operator{
		Name:   crslang.Rx,
		Value:  `\A(?:` + strings.Join(quoted, "|") + `)\z`,
		Negate: negated(c.GetOperator()),
	}

	return []*crslang.SecRule{newLink(target, op)}, nil
}

func stringLink(
	target exprTarget, c *rulepb.Condition,
) ([]*crslang.SecRule, error) {
	value := c.GetValue().GetStringValue()
	if err := validateOperand(value); err != nil {
		return nil, err
	}

	op := crslang.Operator{Value: value}

	switch c.GetOperator() {
	case rulepb.Operator_OPERATOR_EQ:
		op.Name = crslang.StrEq
	case rulepb.Operator_OPERATOR_NE:
		op.Name, op.Negate = crslang.StrEq, true
	case rulepb.Operator_OPERATOR_CONTAINS:
		op.Name = crslang.Contains
	case rulepb.Operator_OPERATOR_PREFIX:
		op.Name = crslang.BeginsWith
	case rulepb.Operator_OPERATOR_SUFFIX:
		op.Name = crslang.EndsWith
	case rulepb.Operator_OPERATOR_MATCHES:
		if _, err := regexp.Compile(value); err != nil {
			return nil, fmt.Errorf(
				"%w: expression regex %q", ErrInvalidSetting, value,
			)
		}

		op.Name = crslang.Rx
	default:
		return nil, operatorError(c)
	}

	return []*crslang.SecRule{newLink(target, op)}, nil
}

// ipLink the client address against addresses and CIDR blocks.
func ipLink(c *rulepb.Condition) ([]*crslang.SecRule, error) {
	values := []string{c.GetValue().GetStringValue()}
	if membership(c.GetOperator()) {
		values = stringValues(c.GetValue().GetListValue())
	}

	switch c.GetOperator() {
	case rulepb.Operator_OPERATOR_EQ, rulepb.Operator_OPERATOR_NE,
		rulepb.Operator_OPERATOR_IN, rulepb.Operator_OPERATOR_NOT_IN:
	default:
		return nil, operatorError(c)
	}

	if len(values) == 0 {
		return nil, operatorError(c)
	}

	for _, v := range values {
		if !validIP(v) {
			return nil, fmt.Errorf(
				"%w: expression address %q", ErrInvalidSetting, v,
			)
		}
	}

	op := crslang.Operator{
		Name:   crslang.IpMatch,
		Value:  strings.Join(values, ","),
		Negate: negated(c.GetOperator()),
	}
	target := exprTarget{variable: crslang.REMOTE_ADDR}

	return []*crslang.SecRule{newLink(target, op)}, nil
}

func newLink(target exprTarget, op crslang.Operator) *crslang.SecRule {
	rule := crslang.NewSecRule()
	rule.Operator = op
	rule.Transformations.Transformations = []crslang.Transformation{
		crslang.None,
	}

	if target.key == "" {
		rule.Variables = []crslang.Variable{{Name: target.variable}}
		return rule
	}

	rule.Collections = []crslang.Collection{{
		Name:      target.collection,
		Arguments: []string{target.key},
		Count:     target.count,
	}}

	return rule
}

func membership(op rulepb.Operator) bool {
	return op == rulepb.Operator_OPERATOR_IN ||
		op == rulepb.Operator_OPERATOR_NOT_IN
}

func negated(op rulepb.Operator) bool {
	return op == rulepb.Operator_OPERATOR_NE ||
		op == rulepb.Operator_OPERATOR_NOT_IN
}

func stringValues(list *structpb.ListValue) []string {
	values := make([]string, 0, len(list.GetValues()))
	for _, v := range list.GetValues() {
		values = append(values, v.GetStringValue())
	}

	return values
}

func validIP(s string) bool {
	if _, err := netip.ParseAddr(s); err == nil {
		return true
	}

	_, err := netip.ParsePrefix(s)

	return err == nil
}

// validateOperand rejects what could escape the quoting of a directive
// or expand a macro.
func validateOperand(s string) error {
	bad := s == "" || strings.HasSuffix(s, `\`) ||
		strings.Contains(s, "%{") ||
		strings.ContainsFunc(s, func(r rune) bool {
			return r == '"' || r < 0x20 || r == 0x7f
		})
	if bad {
		return fmt.Errorf("%w: expression value %q", ErrInvalidSetting, s)
	}

	return nil
}

func operatorError(c *rulepb.Condition) error {
	return fmt.Errorf("%w: expression %s %s on %s",
		ErrInvalidSetting, c.GetOperator(), valueKind(c), c.GetSource(),
	)
}

func valueKind(c *rulepb.Condition) string {
	if c.GetValue().GetListValue() != nil {
		return "list"
	}

	return "value"
}
