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
// back. A condition a script cannot write (a regex, a CIDR, the body...)
// returns an error wrapping scripting.ErrNotDeclarative.
func ToScript(expr *rulepb.Expression, action *rulepb.Action,
) (string, error) {
	anyOf, err := branchesOf(expr)
	if err != nil {
		return "", err
	}

	a, err := actionOf(action)
	if err != nil {
		return "", err
	}

	return scripting.MatchSource(scripting.Rule{Any: anyOf, Action: a})
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

	return exprOf(r.Any), action, nil
}

// ToKeyScript the rate limiter script counting, under the client IP, the
// requests expr matches: FromKeyScript reads expr back.
func ToKeyScript(expr *rulepb.Expression) (string, error) {
	anyOf, err := branchesOf(expr)
	if err != nil {
		return "", err
	}

	return scripting.KeySource(anyOf)
}

// FromKeyScript the expr of a rate limiter script of the declarative
// form (scripting.ParseKey).
func FromKeyScript(source string) (*rulepb.Expression, error) {
	anyOf, err := scripting.ParseKey(source)
	if err != nil {
		return nil, err
	}

	return exprOf(anyOf), nil
}

// exprOf the expression of the branches anyOf: an and condition each.
func exprOf(anyOf [][]scripting.Cond) *rulepb.Expression {
	expr := &rulepb.Expression{
		OrCondition: make([]*rulepb.AndCondition, 0, len(anyOf)),
	}

	for _, branch := range anyOf {
		and := &rulepb.AndCondition{
			Rules: make([]*rulepb.Rule, 0, len(branch)),
		}

		for _, c := range branch {
			and.Rules = append(and.Rules, &rulepb.Rule{Condition: condPb(c)})
		}

		expr.OrCondition = append(expr.OrCondition, and)
	}

	return expr
}

// actionOf the script action of a.
func actionOf(a *rulepb.Action) (scripting.Action, error) {
	name := strings.ToLower(
		strings.TrimPrefix(a.GetType().String(), _actionPrefix))
	if !scripting.IsAction(name) {
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
