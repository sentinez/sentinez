// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rulescript

import (
	"fmt"
	"strings"

	"github.com/sentinez/scripting"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// _maxBranches the most branches of an expression in disjunctive normal
// form, as scripting bounds them.
const _maxBranches = 64

// _sources the script source of each field source; the others (body,
// JA4, TLS) have no ctx method.
var _sources = map[rulepb.FieldSource]scripting.Source{
	rulepb.FieldSource_FIELD_SOURCE_PATH:   scripting.SourcePath,
	rulepb.FieldSource_FIELD_SOURCE_METHOD: scripting.SourceMethod,
	rulepb.FieldSource_FIELD_SOURCE_HOST:   scripting.SourceHost,
	rulepb.FieldSource_FIELD_SOURCE_IP:     scripting.SourceIP,
	rulepb.FieldSource_FIELD_SOURCE_HEADER: scripting.SourceHeader,
	rulepb.FieldSource_FIELD_SOURCE_QUERY:  scripting.SourceQuery,
}

// _ops the script operator of each operator of a single value. IN, NOT_IN
// and the lists are expanded (listAlts); MATCHES, an RE2 regex, has no
// Lua equivalent.
var _ops = map[rulepb.Operator]scripting.Op{
	rulepb.Operator_OPERATOR_EQ:       scripting.OpEq,
	rulepb.Operator_OPERATOR_NE:       scripting.OpNe,
	rulepb.Operator_OPERATOR_CONTAINS: scripting.OpContains,
	rulepb.Operator_OPERATOR_PREFIX:   scripting.OpPrefix,
	rulepb.Operator_OPERATOR_SUFFIX:   scripting.OpSuffix,
	rulepb.Operator_OPERATOR_GT:       scripting.OpGt,
	rulepb.Operator_OPERATOR_GTE:      scripting.OpGte,
	rulepb.Operator_OPERATOR_LT:       scripting.OpLt,
	rulepb.Operator_OPERATOR_LTE:      scripting.OpLte,
}

// _presenceKeys the keys of the header and query conditions checking
// which fields are present (szrule), not a value.
var _presenceKeys = map[string]bool{"": true, "header": true, "query": true}

// branchesOf expr in disjunctive normal form: the edge matches a request
// when one of its or conditions does, an empty expr matching every request.
func branchesOf(expr *rulepb.Expression) ([][]scripting.Cond, error) {
	return orBranches(expr.GetOrCondition())
}

func orBranches(ors []*rulepb.AndCondition) ([][]scripting.Cond, error) {
	if len(ors) == 0 {
		return [][]scripting.Cond{nil}, nil
	}

	var out [][]scripting.Cond
	for _, and := range ors {
		branches, err := andBranches(and)
		if err != nil {
			return nil, err
		}

		out = append(out, branches...)
		if len(out) > _maxBranches {
			return nil, errTooMany
		}
	}

	return out, nil
}

// andBranches the rules of and, all of them, and one of its nested or
// conditions.
func andBranches(and *rulepb.AndCondition) ([][]scripting.Cond, error) {
	branches := [][]scripting.Cond{nil}

	for _, r := range and.GetRules() {
		alts, err := alternatives(r.GetCondition())
		if err != nil {
			return nil, err
		}

		if branches, err = product(branches, alts); err != nil {
			return nil, err
		}
	}

	if len(and.GetOrCondition()) == 0 {
		return branches, nil
	}

	nested, err := orBranches(and.GetOrCondition())
	if err != nil {
		return nil, err
	}

	return product(branches, nested)
}

var errTooMany = fmt.Errorf("%w: more than %d branches",
	scripting.ErrNotDeclarative, _maxBranches)

// product every branch of a joined to every branch of b.
func product(a, b [][]scripting.Cond) ([][]scripting.Cond, error) {
	if len(a)*len(b) > _maxBranches {
		return nil, errTooMany
	}

	out := make([][]scripting.Cond, 0, len(a)*len(b))
	for _, x := range a {
		for _, y := range b {
			out = append(out, append(append([]scripting.Cond{}, x...), y...))
		}
	}

	return out, nil
}

// alternatives the branches c expands to: one condition, or a list of
// values matched one of (EQ, IN) or none of (NE, NOT_IN).
func alternatives(c *rulepb.Condition) ([][]scripting.Cond, error) {
	if c == nil {
		return nil, errNoCondition
	}

	field, err := fieldOf(c)
	if err != nil {
		return nil, err
	}

	if list := c.GetValue().GetListValue(); list != nil {
		return listAlts(field, c.GetOperator(), list.GetValues())
	}

	return single(field, c)
}

// fieldOf the source and key c reads.
func fieldOf(c *rulepb.Condition) (scripting.Cond, error) {
	src, ok := _sources[c.GetSource()]
	if !ok {
		return scripting.Cond{}, unsupported("source %s", c.GetSource())
	}

	field := scripting.Cond{Source: src}
	if !src.Keyed() {
		return field, nil
	}

	if _presenceKeys[c.GetKey()] {
		return scripting.Cond{}, unsupported("presence of %s", c.GetSource())
	}

	field.Key = c.GetKey()

	return field, nil
}

func single(field scripting.Cond, c *rulepb.Condition,
) ([][]scripting.Cond, error) {
	v, ok := c.GetValue().GetKind().(*structpb.Value_StringValue)
	if !ok {
		return nil, unsupported("value of %s: not a string", c.GetSource())
	}

	op, ok := _ops[c.GetOperator()]
	ipOp := op == scripting.OpEq || op == scripting.OpNe
	isIP := field.Source == scripting.SourceIP
	switch {
	case !ok:
		return nil, unsupported("operator %s", c.GetOperator())
	case isIP && (!ipOp || strings.Contains(v.StringValue, "/")):
		// the edge reads a CIDR as a prefix, and compares nothing else
		return nil, unsupported("ip %s %q", c.GetOperator(), v.StringValue)
	}

	field.Op, field.Value = op, v.StringValue

	return [][]scripting.Cond{{field}}, nil
}

func unsupported(format string, args ...any) error {
	return fmt.Errorf("%w: "+format,
		append([]any{scripting.ErrNotDeclarative}, args...)...)
}
