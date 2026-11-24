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

// listNode the node of field compared by op to the values: one of the
// values equal to it, or none of them.
func listNode(field scripting.Cond, op rulepb.Operator,
	values []*structpb.Value) (scripting.Node, error) {

	oneOf, ok := _listOps[field.Source][op]
	if !ok || len(values) == 0 {
		return scripting.Node{}, unsupported("list with %s", op)
	}

	cmp := scripting.OpNe
	if oneOf {
		cmp = scripting.OpEq
	}

	nodes := make([]scripting.Node, 0, len(values))
	for _, v := range values {
		s, ok := v.GetKind().(*structpb.Value_StringValue)
		if !ok {
			return scripting.Node{}, unsupported("list value %v", v)
		}

		c := field
		c.Op, c.Value = cmp, s.StringValue
		if c.Source == scripting.SourceIP {
			// a slash: a CIDR prefix, IN ip_in, NOT_IN not ip_in
			c.Op, _ = ipOp(cmp, s.StringValue)
		}

		nodes = append(nodes, scripting.Leaf(c))
	}

	if oneOf {
		return scripting.Any(nodes...), nil
	}

	return scripting.All(nodes...), nil
}
