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
	"slices"

	"github.com/sentinez/scripting/internal/api"
)

// The global functions of the rule scripts.
const (
	// MatchFunc the function of a secrule script.
	MatchFunc = "match"
	// KeyFunc the function of a rate limiter script.
	KeyFunc = "key"
)

// _maxBranches the most branches of a declarative rule, once its
// condition is put in disjunctive normal form.
const _maxBranches = 64

var (
	// ErrNotDeclarative a script outside of the declarative form Parse
	// reads, or a rule the form cannot write.
	ErrNotDeclarative = errors.New("not declarative")

	errNoBranch = errors.New("rule has no branch")
)

// Source a field of the request a condition reads, through the ctx
// method of the same name.
type Source int

const (
	// SourcePath ctx:path().
	SourcePath Source = iota + 1
	// SourceMethod ctx:method().
	SourceMethod
	// SourceHost ctx:host().
	SourceHost
	// SourceIP ctx:ip().
	SourceIP
	// SourceHeader ctx:header(key).
	SourceHeader
	// SourceQuery ctx:query(key).
	SourceQuery
)

// _sourceMethods the ctx method of each source.
var _sourceMethods = map[Source]string{
	SourcePath:   "path",
	SourceMethod: "method",
	SourceHost:   "host",
	SourceIP:     "ip",
	SourceHeader: "header",
	SourceQuery:  "query",
}

// keyed reports whether the source takes a key.
func (s Source) Keyed() bool { return s == SourceHeader || s == SourceQuery }

// Op the comparison of a condition.
type Op int

const (
	// OpEq field == "value".
	OpEq Op = iota + 1
	// OpNe field ~= "value".
	OpNe
	// OpContains field:find("value", 1, true).
	OpContains
	// OpPrefix field:sub(1, #value) == "value".
	OpPrefix
	// OpSuffix field:sub(-#value) == "value".
	OpSuffix
	// OpGt tonumber(field) > value.
	OpGt
	// OpGte tonumber(field) >= value.
	OpGte
	// OpLt tonumber(field) < value.
	OpLt
	// OpLte tonumber(field) <= value.
	OpLte
)

// _numericOps the Lua operator of each numeric comparison.
var _numericOps = map[Op]string{
	OpGt: ">", OpGte: ">=", OpLt: "<", OpLte: "<=",
}

// Cond a condition: the field Source (of Key, for a keyed source) compared
// by Op to Value. A numeric Op takes a decimal Value, e.g. "10" or "0.5".
type Cond struct {
	Source Source
	Key    string
	Op     Op
	Value  string
}

// Action the action of a secrule: its name ("block", which a script
// returns as SZ_BLOCK or "block"), and its params.
type Action struct {
	Name string
	// Params string, float64, bool or map[string]any fields.
	Params map[string]any
}

// IsAction reports whether a secrule script may return the action name.
func IsAction(name string) bool {
	return slices.Contains(api.Actions(), name)
}

// _actionConsts the action names by their global constant: SZ_BLOCK.
var _actionConsts = func() map[string]string {
	m := make(map[string]string)
	for _, name := range api.Actions() {
		m[api.ActionConst(name)] = name
	}

	return m
}()

// Rule the declarative form of a rule script: it matches a request when
// one of the branches of Any does, a branch being the conjunction of its
// conditions; an empty branch matches every request. A secrule takes
// Action, a rate limiter counts the matching requests under the client IP.
//
// MatchSource and KeySource write it as Lua, ParseMatch and ParseKey read
// it back from a script of that form, hand written or not:
//
//	function match(ctx)
//	  if ctx:path() == "/admin" and not (ctx:ip() == "10.0.0.1") then
//	    return SZ_BLOCK
//	  end
//	  if ctx:header("X-Debug") ~= "" then return SZ_BLOCK end
//	end
type Rule struct {
	Any    [][]Cond
	Action Action
}
