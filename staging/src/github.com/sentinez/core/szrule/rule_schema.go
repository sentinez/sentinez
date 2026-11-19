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
	"slices"

	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
)

// The schema of the conditions: for each source, the key it takes and the
// operators and values the matchers of rule_match.go evaluate. It is the
// only place listing them: ValidateExpression checks a condition against
// it, Schema hands it to the console. Keep it in step with the matchers.

// Names of the variants of a source.
const (
	// The source has a single value, or the key names the field to read.
	_variantValue = "value"
	// Without a key: whether the request has the named headers or query
	// parameters.
	_variantPresence = "presence"

	// The keys the header and query matchers read as no key.
	_headerPresenceKey = "header"
	_queryPresenceKey  = "query"
)

type operatorSpec struct {
	op   rulepb.Operator
	kind rulepb.ValueKind
	list bool
	enum []string
}

type variant struct {
	name string
	key  rulepb.KeyMode
	ops  []operatorSpec
}

type sourceSpec struct {
	source   rulepb.FieldSource
	variants []variant
	// The key read as no key, as well as the empty one.
	presenceKey string
}

const (
	_kindString = rulepb.ValueKind_VALUE_KIND_STRING
	_kindNumber = rulepb.ValueKind_VALUE_KIND_NUMBER
	_kindRegex  = rulepb.ValueKind_VALUE_KIND_REGEX
	_kindIp     = rulepb.ValueKind_VALUE_KIND_IP
	_kindEnum   = rulepb.ValueKind_VALUE_KIND_ENUM
	_kindName   = rulepb.ValueKind_VALUE_KIND_NAME

	_keyNone     = rulepb.KeyMode_KEY_MODE_NONE
	_keyRequired = rulepb.KeyMode_KEY_MODE_REQUIRED
)

var (
	// The methods matchSourceMethod compares byte for byte.
	_methods = []string{
		"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE",
		"OPTIONS", "CONNECT", "TRACE",
	}

	// The text comparisons of matchBytes and matchString.
	_textOps = []operatorSpec{
		scalar(rulepb.Operator_OPERATOR_EQ, _kindString),
		scalar(rulepb.Operator_OPERATOR_NE, _kindString),
		scalar(rulepb.Operator_OPERATOR_CONTAINS, _kindString),
		scalar(rulepb.Operator_OPERATOR_PREFIX, _kindString),
		scalar(rulepb.Operator_OPERATOR_SUFFIX, _kindString),
		scalar(rulepb.Operator_OPERATOR_MATCHES, _kindRegex),
	}

	// The numeric comparisons of matchNumber.
	_numberOps = []operatorSpec{
		scalar(rulepb.Operator_OPERATOR_GT, _kindNumber),
		scalar(rulepb.Operator_OPERATOR_GTE, _kindNumber),
		scalar(rulepb.Operator_OPERATOR_LT, _kindNumber),
		scalar(rulepb.Operator_OPERATOR_LTE, _kindNumber),
	}

	// The sources, in the order the console lists them.
	_sources = []sourceSpec{
		{
			source:      rulepb.FieldSource_FIELD_SOURCE_HEADER,
			variants:    keyedVariants(),
			presenceKey: _headerPresenceKey,
		},
		{
			source:      rulepb.FieldSource_FIELD_SOURCE_QUERY,
			variants:    keyedVariants(),
			presenceKey: _queryPresenceKey,
		},
		plain(rulepb.FieldSource_FIELD_SOURCE_PATH, _textOps...),
		plain(rulepb.FieldSource_FIELD_SOURCE_BODY, _textOps...),
		plain(
			rulepb.FieldSource_FIELD_SOURCE_IP,
			membershipOps(_kindIp, nil)...,
		),
		plain(rulepb.FieldSource_FIELD_SOURCE_JA4, _textOps...),
		// matchSourceTLS only compares with EQ and NE
		plain(
			rulepb.FieldSource_FIELD_SOURCE_TLS,
			enumOp(rulepb.Operator_OPERATOR_EQ, "true", "false"),
			enumOp(rulepb.Operator_OPERATOR_NE, "true", "false"),
		),
		plain(
			rulepb.FieldSource_FIELD_SOURCE_METHOD,
			membershipOps(_kindEnum, _methods)...,
		),
		plain(rulepb.FieldSource_FIELD_SOURCE_HOST, _textOps...),
	}
)

// Schema returns the conditions ValidateExpression accepts. The result is
// a new message on each call: the caller may change it.
func Schema() *rulepb.ConditionSchema {
	sources := make([]*rulepb.SourceSpec, 0, len(_sources))
	for _, s := range _sources {
		sources = append(sources, s.proto())
	}

	return &rulepb.ConditionSchema{Sources: sources}
}

func (s sourceSpec) proto() *rulepb.SourceSpec {
	variants := make([]*rulepb.SourceVariant, 0, len(s.variants))
	for _, v := range s.variants {
		ops := make([]*rulepb.OperatorSpec, 0, len(v.ops))
		for _, o := range v.ops {
			ops = append(ops, &rulepb.OperatorSpec{
				Operator:   o.op,
				ValueKind:  o.kind,
				List:       o.list,
				EnumValues: slices.Clone(o.enum),
			})
		}

		variants = append(variants, &rulepb.SourceVariant{
			Name:      v.name,
			KeyMode:   v.key,
			Operators: ops,
		})
	}

	return &rulepb.SourceSpec{Source: s.source, Variants: variants}
}

// keyedVariants the variants of the header and query sources: the value
// of a named field, or the presence of fields when there is no key.
func keyedVariants() []variant {
	value := slices.Concat(_textOps, _numberOps, []operatorSpec{
		list(rulepb.Operator_OPERATOR_IN, _kindString, nil),
		list(rulepb.Operator_OPERATOR_NOT_IN, _kindString, nil),
	})

	return []variant{
		{name: _variantValue, key: _keyRequired, ops: value},
		{
			name: _variantPresence,
			key:  _keyNone,
			ops:  membershipOps(_kindName, nil),
		},
	}
}

// plain a source without a key: it has a single variant.
func plain(source rulepb.FieldSource, ops ...operatorSpec) sourceSpec {
	return sourceSpec{source: source, variants: valueVariant(ops...)}
}

// valueVariant the single variant of a source without a key.
func valueVariant(ops ...operatorSpec) []variant {
	return []variant{{name: _variantValue, key: _keyNone, ops: ops}}
}

// membershipOps EQ and NE to a value, IN and NOT_IN a list of values.
func membershipOps(kind rulepb.ValueKind, enum []string) []operatorSpec {
	return []operatorSpec{
		{op: rulepb.Operator_OPERATOR_EQ, kind: kind, enum: enum},
		{op: rulepb.Operator_OPERATOR_NE, kind: kind, enum: enum},
		list(rulepb.Operator_OPERATOR_IN, kind, enum),
		list(rulepb.Operator_OPERATOR_NOT_IN, kind, enum),
	}
}

func scalar(op rulepb.Operator, kind rulepb.ValueKind) operatorSpec {
	return operatorSpec{op: op, kind: kind}
}

func list(op rulepb.Operator, kind rulepb.ValueKind,
	enum []string) operatorSpec {

	return operatorSpec{op: op, kind: kind, list: true, enum: enum}
}

func enumOp(op rulepb.Operator, values ...string) operatorSpec {
	return operatorSpec{op: op, kind: _kindEnum, enum: values}
}

func findSource(source rulepb.FieldSource) (sourceSpec, bool) {
	for _, s := range _sources {
		if s.source == source {
			return s, true
		}
	}

	return sourceSpec{}, false
}

func (v variant) operator(op rulepb.Operator) (operatorSpec, bool) {
	for _, o := range v.ops {
		if o.op == op {
			return o, true
		}
	}

	return operatorSpec{}, false
}
