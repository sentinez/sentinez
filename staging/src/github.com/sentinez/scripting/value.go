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
	"errors"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

const (
	_maxDepth = 4
	_maxItems = 64
)

var errTooLarge = errors.New("table too deep or too large")

// Kind is the type of the Value a rule returned.
type Kind int

const (
	// KindNil nil, or no result.
	KindNil Kind = iota + 1
	// KindBool a boolean.
	KindBool
	// KindString a string.
	KindString
	// KindTable a table with string keys.
	KindTable
)

// Value the result of a rule function, free of the Lua types.
type Value struct {
	Kind Kind
	Bool bool
	Str  string
	// Table the fields of a table: string, float64, bool or
	// map[string]any, as structpb.NewValue takes them.
	Table map[string]any
}

// convert the Lua value v: a function, userdata or a table with other
// than string keys (or an array) is an error.
func convert(v lua.LValue) (Value, error) {
	switch x := v.(type) {
	case *lua.LNilType:
		return Value{Kind: KindNil}, nil
	case lua.LBool:
		return Value{Kind: KindBool, Bool: bool(x)}, nil
	case lua.LString:
		return Value{Kind: KindString, Str: string(x)}, nil
	case *lua.LTable:
		m, err := toMap(x, 0)
		if err != nil {
			return Value{}, err
		}

		return Value{Kind: KindTable, Table: m}, nil
	}

	return Value{}, fmt.Errorf("unsupported result of type %s", v.Type())
}

func toMap(t *lua.LTable, depth int) (map[string]any, error) {
	if depth >= _maxDepth || t.Len() > 0 {
		return nil, errTooLarge
	}

	m := make(map[string]any)

	var err error
	t.ForEach(func(k, v lua.LValue) {
		if err != nil {
			return
		}

		ks, ok := k.(lua.LString)
		if !ok || len(m) >= _maxItems {
			err = errTooLarge
			return
		}

		m[string(ks)], err = toAny(v, depth)
	})

	return m, err
}

func toAny(v lua.LValue, depth int) (any, error) {
	switch x := v.(type) {
	case lua.LBool:
		return bool(x), nil
	case lua.LString:
		return string(x), nil
	case lua.LNumber:
		return float64(x), nil
	case *lua.LTable:
		return toMap(x, depth+1)
	}

	return nil, fmt.Errorf("unsupported field of type %s", v.Type())
}
