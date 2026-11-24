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

// Package rulescript converts the expr and action of a secrule or a rate
// limiter to the Lua script of the same rule, and back: the declarative
// form of package scripting (scripting.Rule) bridges the two.
package rulescript

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sentinez/scripting"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// _actionPrefix the prefix of the ActionType names, ACTION_TYPE_BLOCK
// for the script action "block".
const _actionPrefix = "ACTION_TYPE_"

var errNoCondition = errors.New("rule without condition")

// ToScript the secrule script of expr and action: FromScript reads them
// back. A condition a script cannot write (an invalid regex, a value that
// is not a string...) returns an error wrapping scripting.ErrNotDeclarative.
func ToScript(expr *rulepb.Expression, action *rulepb.Action,
) (string, error) {
	match, err := nodeOf(expr)
	if err != nil {
		return "", err
	}

	a, err := actionOf(action)
	if err != nil {
		return "", err
	}

	return scripting.MatchSource(scripting.Rule{Match: match, Action: a})
}

// FromScript the expr and action of a secrule script of the declarative
// form (scripting.ParseMatch), hand written or from ToScript.
func FromScript(source string) (*rulepb.Expression, *rulepb.Action,
	error) {

	r, err := scripting.ParseMatch(source)
	if err != nil {
		return nil, nil, err
	}

	action, err := actionPb(r.Action)
	if err != nil {
		return nil, nil, err
	}

	return exprOf(r.Match), action, nil
}

// ToCondition the Lua condition of expr, the if of the scripts ToScript
// and ToKeyScript write (scripting.CondSource). A condition a script
// cannot write returns an error wrapping scripting.ErrNotDeclarative.
func ToCondition(expr *rulepb.Expression) (string, error) {
	match, err := nodeOf(expr)
	if err != nil {
		return "", err
	}

	return scripting.CondSource(match)
}

// ToKeyScript the rate limiter script counting, under the client IP, the
// requests expr matches: FromKeyScript reads expr back.
func ToKeyScript(expr *rulepb.Expression) (string, error) {
	match, err := nodeOf(expr)
	if err != nil {
		return "", err
	}

	return scripting.KeySource(match)
}

// FromKeyScript the expr of a rate limiter script of the declarative
// form (scripting.ParseKey).
func FromKeyScript(source string) (*rulepb.Expression, error) {
	match, err := scripting.ParseKey(source)
	if err != nil {
		return nil, err
	}

	return exprOf(match), nil
}

// exprOf the expression of the condition match: each of its branches is
// an or condition of the expression, an and condition.
func exprOf(match scripting.Node) *rulepb.Expression {
	return &rulepb.Expression{OrCondition: andsOf(match.Branches())}
}

// andsOf the and condition of each of nodes.
func andsOf(nodes []scripting.Node) []*rulepb.AndCondition {
	ands := make([]*rulepb.AndCondition, 0, len(nodes))
	for _, n := range nodes {
		ands = append(ands, andOf(n))
	}

	return ands
}

// andOf the and condition matching as n: its conditions are the rules, its
// first or group the or condition, and each other group an and condition
// of its own, as (a or b) and (c or d) is
//
//	{or_condition: [a, b], and_condition: [{or_condition: [c, d]}]}
func andOf(n scripting.Node) *rulepb.AndCondition {
	switch {
	case n.Cond != nil:
		return &rulepb.AndCondition{Rules: rulesOf(n)}
	case n.IsOr():
		return &rulepb.AndCondition{OrCondition: andsOf(n.Nodes)}
	}

	and := &rulepb.AndCondition{}
	for _, c := range n.Nodes {
		switch {
		case c.Cond != nil:
			and.Rules = append(and.Rules, rulesOf(c)...)
		case c.IsOr() && len(and.OrCondition) == 0:
			and.OrCondition = andsOf(c.Nodes)
		default:
			and.AndCondition = append(and.AndCondition, andOf(c))
		}
	}

	return and
}

// rulesOf the rule of the leaf n.
func rulesOf(n scripting.Node) []*rulepb.Rule {
	return []*rulepb.Rule{{Condition: condPb(*n.Cond)}}
}

// ActionName the name a script returns the action type t by, ok false
// when no script action has t.
func ActionName(t rulepb.ActionType) (name string, ok bool) {
	name = strings.ToLower(strings.TrimPrefix(t.String(), _actionPrefix))

	return name, scripting.IsAction(name)
}

// actionOf the script action of a.
func actionOf(a *rulepb.Action) (scripting.Action, error) {
	name, ok := ActionName(a.GetType())
	if !ok {
		return scripting.Action{}, fmt.Errorf("%w: action %s",
			scripting.ErrNotDeclarative, a.GetType())
	}

	var params map[string]any
	if len(a.GetParams().GetFields()) > 0 {
		params = a.GetParams().AsMap()
	}

	return scripting.Action{Name: name, Params: params}, nil
}

// actionPb the action of the script action a.
func actionPb(a scripting.Action) (*rulepb.Action, error) {
	typ, ok := rulepb.ActionType_value[_actionPrefix+strings.ToUpper(a.Name)]
	if !ok {
		return nil, fmt.Errorf("%w: action %q",
			scripting.ErrNotDeclarative, a.Name)
	}

	action := &rulepb.Action{Type: rulepb.ActionType(typ)}
	if len(a.Params) == 0 {
		return action, nil
	}

	params, err := structpb.NewStruct(a.Params)
	if err != nil {
		return nil, fmt.Errorf("action params: %w", err)
	}

	action.Params = params

	return action, nil
}
