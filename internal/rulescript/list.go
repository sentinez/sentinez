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
	"strings"

	"github.com/sentinez/scripting"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// _listOps the operators taking a list, by source: whether the request
// must equal one of its values (true) or none of them (false), as szrule
// matches them.
var _listOps = map[scripting.Source]map[rulepb.Operator]bool{
	scripting.SourceIP:     {_in: true, _notIn: false},
	scripting.SourceMethod: {_in: true, _notIn: false},
	scripting.SourceHeader: {_eq: true, _in: true, _ne: false, _notIn: false},
	scripting.SourceQuery:  {_eq: true, _in: true, _ne: false, _notIn: false},
}

const (
	_eq    = rulepb.Operator_OPERATOR_EQ
	_ne    = rulepb.Operator_OPERATOR_NE
	_in    = rulepb.Operator_OPERATOR_IN
	_notIn = rulepb.Operator_OPERATOR_NOT_IN
)

// listAlts the branches of field compared by op to the values: one per
// value equal to it, or a single one different from all of them.
func listAlts(field scripting.Cond, op rulepb.Operator,
	values []*structpb.Value) ([][]scripting.Cond, error) {

	oneOf, ok := _listOps[field.Source][op]
	if !ok || len(values) == 0 {
		return nil, unsupported("list with %s", op)
	}

	conds := make([]scripting.Cond, 0, len(values))
	for _, v := range values {
		s, ok := v.GetKind().(*structpb.Value_StringValue)
		isCIDR := ok && field.Source == scripting.SourceIP &&
			strings.Contains(s.StringValue, "/")
		if !ok || isCIDR {
			return nil, unsupported("list value %v", v)
		}

		c := field
		c.Op, c.Value = scripting.OpNe, s.StringValue
		if oneOf {
			c.Op = scripting.OpEq
		}

		conds = append(conds, c)
	}

	if !oneOf {
		return [][]scripting.Cond{conds}, nil
	}

	alts := make([][]scripting.Cond, 0, len(conds))
	for _, c := range conds {
		alts = append(alts, []scripting.Cond{c})
	}

	return alts, nil
}

// condPb the condition of c.
func condPb(c scripting.Cond) *rulepb.Condition {
	cond := &rulepb.Condition{
		Key:   c.Key,
		Value: structpb.NewStringValue(c.Value),
	}

	for src, s := range _sources {
		if s == c.Source {
			cond.Source = src
		}
	}

	for op, o := range _ops {
		if o == c.Op {
			cond.Operator = op
		}
	}

	return cond
}
