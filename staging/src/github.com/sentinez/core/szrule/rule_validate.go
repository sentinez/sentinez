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

package szrule

import (
	"errors"
	"fmt"
	"math"
	"net/netip"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/sentinez/contrib/pkg/bytesconv"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// ErrInvalidCondition a condition of an expression is outside of Schema.
var ErrInvalidCondition = errors.New("invalid condition")

const (
	// Groups nested deeper are rejected: the evaluator recurses on them.
	_maxDepth = 8
	// Longest key or value, in bytes.
	_maxValueLen = 4096
	// Most items of a list value.
	_maxListLen = 256
)

// ValidateExpression checks every condition of expr against Schema. The
// error wraps ErrInvalidCondition and starts with the JSON path of the
// first invalid condition, e.g. "orCondition[0].rules[1]". An empty
// expression is valid: it matches every request.
func ValidateExpression(expr *rulepb.Expression) error {
	return validateOr(expr.GetOrCondition(), "orCondition", 1)
}

// ValidateCondition checks c against Schema: a known source, the key of
// one of its variants, an operator of that variant and a value of the
// kind of that operator. The error wraps ErrInvalidCondition.
func ValidateCondition(c *rulepb.Condition) error {
	if c == nil {
		return conditionErrorf("missing")
	}

	src, ok := findSource(c.GetSource())
	if !ok {
		return conditionErrorf("unsupported source %q",
			c.GetSource().String())
	}

	v, err := src.variantFor(c.GetKey())
	if err != nil {
		return err
	}

	op, ok := v.operator(c.GetOperator())
	if !ok {
		return conditionErrorf("operator %q does not apply to %s",
			operatorName(c.GetOperator()), describe(src, v))
	}

	return op.validate(src.source, c.GetValue())
}

func validateOr(or []*rulepb.AndCondition, path string, depth int) error {
	if len(or) > 0 && depth > _maxDepth {
		return fmt.Errorf("%s: %w: groups nested deeper than %d",
			path, ErrInvalidCondition, _maxDepth)
	}

	for i, and := range or {
		p := path + "[" + strconv.Itoa(i) + "]"
		if err := validateAnd(and, p, depth); err != nil {
			return err
		}
	}

	return nil
}

func validateAnd(and *rulepb.AndCondition, path string, depth int) error {
	if len(and.GetRules()) == 0 && len(and.GetOrCondition()) == 0 {
		return fmt.Errorf("%s: %w: empty group", path, ErrInvalidCondition)
	}

	for i, rule := range and.GetRules() {
		if err := ValidateCondition(rule.GetCondition()); err != nil {
			return fmt.Errorf("%s.rules[%d]: %w", path, i, err)
		}
	}

	return validateOr(and.GetOrCondition(), path+".orCondition", depth+1)
}

// variantFor the variant of the key: the one of its key mode, checked
// when the variant names a field. The empty key and the presence key of
// the source (header, query) read no field.
func (s sourceSpec) variantFor(key string) (variant, error) {
	mode := _keyRequired
	if key == "" || key == s.presenceKey {
		mode = _keyNone
	}

	i := slices.IndexFunc(s.variants, func(v variant) bool {
		return v.key == mode
	})
	switch {
	case i < 0 && key == "":
		return variant{}, conditionErrorf("%s requires a key",
			sourceName(s.source))
	case i < 0:
		return variant{}, conditionErrorf("%s takes no key, got %q",
			sourceName(s.source), key)
	}

	v := s.variants[i]
	if mode == _keyNone {
		return v, nil
	}

	if err := validName(s.source, key); err != nil {
		return v, conditionErrorf("key: %v", err)
	}

	return v, nil
}

// validate the value of a condition: a string, or a list of strings for
// a list operator, the type the matchers read.
func (o operatorSpec) validate(source rulepb.FieldSource,
	value *structpb.Value) error {

	var err error
	if o.list {
		err = o.checkList(source, value)
	} else {
		err = o.checkScalar(source, value)
	}

	if err != nil {
		return conditionErrorf("operator %q: %v", operatorName(o.op), err)
	}

	return nil
}

func (o operatorSpec) checkScalar(source rulepb.FieldSource,
	value *structpb.Value) error {

	s, ok := value.GetKind().(*structpb.Value_StringValue)
	if !ok {
		return errors.New("the value must be a string")
	}

	return o.check(source, s.StringValue)
}

func (o operatorSpec) checkList(source rulepb.FieldSource,
	value *structpb.Value) error {

	list, ok := value.GetKind().(*structpb.Value_ListValue)
	if !ok {
		return errors.New("the value must be a list of strings")
	}

	items := list.ListValue.GetValues()
	if len(items) == 0 || len(items) > _maxListLen {
		return fmt.Errorf("the list must have 1 to %d values, got %d",
			_maxListLen, len(items))
	}

	for i, item := range items {
		s, ok := item.GetKind().(*structpb.Value_StringValue)
		if !ok {
			return fmt.Errorf("value %d is not a string", i)
		}

		if err := o.check(source, s.StringValue); err != nil {
			return fmt.Errorf("value %d: %w", i, err)
		}
	}

	return nil
}

func (o operatorSpec) check(source rulepb.FieldSource, s string) error {
	if s == "" || len(s) > _maxValueLen {
		return fmt.Errorf("the value must be 1 to %d bytes", _maxValueLen)
	}

	return o.checkKind(source, s)
}

func (o operatorSpec) checkKind(source rulepb.FieldSource, s string) error {
	switch o.kind {
	case _kindNumber:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return fmt.Errorf("%q is not a number", s)
		}
	case _kindRegex:
		if _, err := regexp.Compile(s); err != nil {
			return fmt.Errorf("%q is not a regular expression", s)
		}
	case _kindIp:
		if !ValidIP(s) {
			return fmt.Errorf("%q is not an IP address or prefix", s)
		}
	case _kindEnum:
		if !slices.Contains(o.enum, s) {
			return fmt.Errorf("%q is not one of %s",
				s, strings.Join(o.enum, ", "))
		}
	case _kindName:
		return validName(source, s)
	}

	return nil
}

// validName a header name is a token (RFC 9110), a query parameter any
// text without control characters.
func validName(source rulepb.FieldSource, s string) error {
	if s == "" || len(s) > _maxValueLen {
		return fmt.Errorf("the name must be 1 to %d bytes", _maxValueLen)
	}

	valid := isQueryChar
	if source == rulepb.FieldSource_FIELD_SOURCE_HEADER {
		valid = isTokenChar
	}

	if strings.IndexFunc(s, func(r rune) bool { return !valid(r) }) >= 0 {
		return fmt.Errorf("%q is not a valid %s name", s, sourceName(source))
	}

	return nil
}

// isTokenChar the characters of a header name, RFC 9110 token.
func isTokenChar(r rune) bool {
	return r < utf8.RuneSelf && bytesconv.ValidHeaderFieldNameTable[r] != 0
}

func isQueryChar(r rune) bool {
	return r >= 0x20 && r != 0x7f
}

// ValidIP reports whether s is an IP address, compared as is, or a CIDR
// prefix containing the client address, as matchIP reads it.
func ValidIP(s string) bool {
	if _, err := netip.ParseAddr(s); err == nil {
		return true
	}

	_, err := netip.ParsePrefix(s)

	return err == nil
}

// describe the source, with the variant when the source has several.
func describe(s sourceSpec, v variant) string {
	if len(s.variants) < 2 {
		return sourceName(s.source)
	}

	return sourceName(s.source) + " " + v.name
}

func sourceName(s rulepb.FieldSource) string {
	return enumName(s.String(), _sourcePrefix)
}

func operatorName(o rulepb.Operator) string {
	return enumName(o.String(), _operatorPrefix)
}

func conditionErrorf(format string, args ...any) error {
	return fmt.Errorf("%w: %s",
		ErrInvalidCondition, fmt.Sprintf(format, args...))
}
