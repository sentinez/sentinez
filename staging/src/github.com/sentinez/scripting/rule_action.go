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

package scripting

import (
	"strconv"

	"github.com/yuin/gopher-lua/ast"
)

// parseAction the action e returns: a constant (SZ_BLOCK), a name, or
// {type = SZ_REDIRECT, params = t}.
func parseAction(e ast.Expr) (Action, error) {
	if t, ok := e.(*ast.TableExpr); ok {
		return tableAction(t)
	}

	name, err := actionName(e)
	if err != nil {
		return Action{}, lineErr(e, "return an action or action table")
	}

	if !IsAction(name) {
		return Action{}, lineErr(e, "unknown action "+strconv.Quote(name))
	}

	return Action{Name: name}, nil
}

func tableAction(t *ast.TableExpr) (Action, error) {
	var a Action
	for _, f := range t.Fields {
		key, ok := f.Key.(*ast.StringExpr)
		if !ok {
			return Action{}, lineErr(t, "an action table has named fields")
		}

		var err error
		switch key.Value {
		case "type":
			a.Name, err = actionName(f.Value)
		case "params":
			a.Params, err = paramsOf(f.Value, 0)
		default:
			err = lineErr(t, "unknown action field "+key.Value)
		}

		if err != nil {
			return Action{}, err
		}
	}

	if !IsAction(a.Name) {
		return Action{}, lineErr(t, "unknown action "+strconv.Quote(a.Name))
	}

	return a, nil
}

// actionName the name of the action e names: SZ_BLOCK or "block".
func actionName(e ast.Expr) (string, error) {
	switch x := e.(type) {
	case *ast.IdentExpr:
		if name, ok := _actionConsts[x.Value]; ok {
			return name, nil
		}
	case *ast.StringExpr:
		return x.Value, nil
	}

	return "", lineErr(e, "the action is an SZ_ constant or a string")
}

// paramsOf the fields of the table constructor e.
func paramsOf(e ast.Expr, depth int) (map[string]any, error) {
	t, ok := e.(*ast.TableExpr)
	if !ok || depth >= _maxDepth || len(t.Fields) > _maxItems {
		return nil, lineErr(e, "params is a table of named fields")
	}

	m := make(map[string]any, len(t.Fields))
	for _, f := range t.Fields {
		key, ok := f.Key.(*ast.StringExpr)
		if !ok {
			return nil, lineErr(e, "params is a table of named fields")
		}

		v, err := paramValue(f.Value, depth)
		if err != nil {
			return nil, err
		}

		m[key.Value] = v
	}

	return m, nil
}

func paramValue(e ast.Expr, depth int) (any, error) {
	switch x := e.(type) {
	case *ast.StringExpr:
		return x.Value, nil
	case *ast.TrueExpr:
		return true, nil
	case *ast.FalseExpr:
		return false, nil
	case *ast.NumberExpr:
		return number(e, x.Value)
	case *ast.UnaryMinusOpExpr:
		if n, ok := x.Expr.(*ast.NumberExpr); ok {
			return number(e, "-"+n.Value)
		}
	case *ast.TableExpr:
		return paramsOf(x, depth+1)
	}

	return nil, lineErr(e, "a param is a string, number, bool or table")
}

func number(e ast.Expr, s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, lineErr(e, "unsupported number "+s)
	}

	return f, nil
}
