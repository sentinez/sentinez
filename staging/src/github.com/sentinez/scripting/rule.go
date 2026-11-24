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

	"github.com/sentinez/scripting/internal/api"
)

// The global functions of the rule scripts.
const (
	// MatchFunc the function of a secrule script.
	MatchFunc = "match"
	// KeyFunc the function of a rate limiter script.
	KeyFunc = "key"
)

// MaxConditions the most conditions of a declarative rule, all its groups
// together.
const MaxConditions = 256

var (
	// ErrNotDeclarative a script outside of the declarative form Parse
	// reads, or a rule the form cannot write.
	ErrNotDeclarative = errors.New("not declarative")

	errNoBranch = errors.New("rule has no branch")

	errTooMany = fmt.Errorf("%w: more than %d conditions",
		ErrNotDeclarative, MaxConditions)
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
	// SourceBody ctx:body().
	SourceBody
	// SourceJA4 ctx:ja4().
	SourceJA4
	// SourceTLS ctx:tls(), a boolean: its conditions are OpEq or OpNe to
	// "true" or "false", written ctx:tls() and not ctx:tls().
	SourceTLS
)

// _sourceMethods the ctx method of each source.
var _sourceMethods = map[Source]string{
	SourcePath:   "path",
	SourceMethod: "method",
	SourceHost:   "host",
	SourceIP:     "ip",
	SourceHeader: "header",
	SourceQuery:  "query",
	SourceBody:   "body",
	SourceJA4:    "ja4",
	SourceTLS:    "tls",
}

// _presenceMethods the ctx method telling whether the request has the
// field of a keyed source.
var _presenceMethods = map[Source]string{
	SourceHeader: "has_header",
	SourceQuery:  "has_query",
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
	// OpMatches re.match(field, "value"): value a regular expression of
	// the RE2 syntax of Go.
	OpMatches
	// OpWithin ctx:ip_in("value"): SourceIP in the CIDR prefix value.
	OpWithin
	// OpNotWithin not ctx:ip_in("value").
	OpNotWithin
	// OpExists ctx:has_header(key) or ctx:has_query(key): the request has
	// the field Key of a keyed source, Value empty.
	OpExists
	// OpAbsent not ctx:has_header(key) or not ctx:has_query(key).
	OpAbsent
)

// _negations each Op a not negates, and the Op it gives.
var _negations = map[Op]Op{
	OpEq: OpNe, OpNe: OpEq,
	OpWithin: OpNotWithin, OpNotWithin: OpWithin,
	OpExists: OpAbsent, OpAbsent: OpExists,
}

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
func IsAction(name string) bool { return api.IsAction(name) }

// Node a condition of a rule: a Cond, or a group of nodes all of which
// (And) or one of which (not And) must match. Build one with Leaf, All and
// Any: they keep the groups in the canonical form the parser reads a script
// to, so a rule written and read back is equal to itself.
type Node struct {
	// And the group needs all its Nodes to match, not only one. An empty
	// And matches every request, an empty Or matches none.
	And   bool
	Nodes []Node
	// Cond the condition of a leaf, nil for a group.
	Cond *Cond
}

// Leaf the node of the condition c.
func Leaf(c Cond) Node { return Node{Cond: &c} }

// IsAnd whether n is a group needing all its nodes.
func (n Node) IsAnd() bool { return n.Cond == nil && n.And }

// IsOr whether n is a group needing one of its nodes.
func (n Node) IsOr() bool { return n.Cond == nil && !n.And }

// matchesAll whether n is the empty and group, matching every request.
func (n Node) matchesAll() bool { return n.IsAnd() && len(n.Nodes) == 0 }

// All the node matching when every one of nodes does: the and groups it
// holds are merged into it, and its conditions come before its or groups,
// as the edge evaluates them. A single node is returned as is, none match
// every request.
func All(nodes ...Node) Node {
	var leaves, groups []Node

	var add func(nodes []Node)
	add = func(nodes []Node) {
		for _, n := range nodes {
			switch {
			case n.IsAnd():
				add(n.Nodes)
			case n.Cond != nil:
				leaves = append(leaves, n)
			default:
				groups = append(groups, n)
			}
		}
	}
	add(nodes)

	return group(true, append(leaves, groups...))
}

// Any the node matching when one of nodes does: the or groups it holds are
// merged into it. A single node is returned as is, none match no request.
func Any(nodes ...Node) Node {
	var all []Node
	for _, n := range nodes {
		if n.IsOr() {
			all = append(all, n.Nodes...)
		} else {
			all = append(all, n)
		}
	}

	return group(false, all)
}

// Branches the nodes one of which n needs: those of an or group, n itself
// for any other node. A script writes an if for each.
func (n Node) Branches() []Node {
	if n.IsOr() {
		return n.Nodes
	}

	return []Node{n}
}

// group the node of nodes: the node itself when it is alone.
func group(and bool, nodes []Node) Node {
	if len(nodes) == 1 {
		return nodes[0]
	}

	return Node{And: and, Nodes: nodes}
}

// count the conditions of n.
func (n Node) count() int {
	if n.Cond != nil {
		return 1
	}

	total := 0
	for _, c := range n.Nodes {
		total += c.count()
	}

	return total
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
// Match does. A secrule takes Action, a rate limiter counts the matching
// requests under the client IP.
//
// MatchSource and KeySource write it as Lua, ParseMatch and ParseKey read
// it back from a script of that form, hand written or not:
//
//	function match(ctx)
//	  if ctx:path() == "/admin" and not (ctx:ip() == "10.0.0.1") then
//	    return SZ_BLOCK
//	  end
//	  if ctx:has_header("X-Debug") or re.match(ctx:path(), "^/wp-") then
//	    return SZ_BLOCK
//	  end
//	end
type Rule struct {
	Match  Node
	Action Action
}
