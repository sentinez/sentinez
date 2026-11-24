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

package hooks

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/sentinez/scripting/phase"
	"github.com/yuin/gopher-lua/ast"
)

// stage a stage of the edge chain a hook may annotate.
type stage struct {
	// name the name of the annotations: ---@stage http:request/<name>.
	name string
	// id the id of the plugin of the stage, Stage of the chain.
	id string
}

// _stages the stages of the edge chain a hook may annotate, in chain
// order. The stages before the namespace of the request is known (TRC,
// LOG, DMA) have no hooks. The edge tests the ids against its chain.
var _stages = []stage{
	{name: "cache", id: "CDN"},        // cdn.Cache
	{name: "sec_limiter", id: "LMT"},  // secure.SecLimiter
	{name: "waiting_room", id: "ROM"}, // room.WaitingRoom
	{name: "static", id: "STC"},       // static.Static
	{name: "sec_rule", id: "RUL"},     // secure.SecRule
	{name: "sec_ruleset", id: "RST"},  // secure.SecRuleset
	{name: "router", id: "ROU"},       // routing.StandardRouter
}

// Stages returns the ids of the stages a hook may annotate, in chain
// order.
func Stages() []string {
	ids := make([]string, 0, len(_stages))
	for _, s := range _stages {
		ids = append(ids, s.id)
	}

	return ids
}

// StageName the name of the stage id, the id when it is none.
func StageName(id string) string {
	i := slices.IndexFunc(_stages, func(s stage) bool { return s.id == id })
	if i < 0 {
		return id
	}

	return _stages[i].name
}

// StageNames returns the names of the stages, in chain order.
func StageNames() []string {
	names := make([]string, 0, len(_stages))
	for _, s := range _stages {
		names = append(names, s.name)
	}

	return names
}

// hook a stage, by id, and the phase of it a function runs at.
type hook struct {
	phase phase.Phase
	stage string
}

// Set the global functions each hook runs, in source order.
type Set struct {
	funcs map[hook][]string
}

// Funcs returns the functions hooking the stage id at the phase at, in
// source order.
func (s Set) Funcs(at phase.Phase, id string) []string {
	return s.funcs[hook{phase: at, stage: id}]
}

// Any reports whether a function hooks any stage.
func (s Set) Any() bool { return len(s.funcs) > 0 }

// Has reports whether a function hooks the stage id at the phase at.
func (s Set) Has(at phase.Phase, id string) bool {
	return len(s.Funcs(at, id)) > 0
}

// _annotation ---@stage <phase>/<stage>
var _annotation = regexp.MustCompile(`^\s*---\s*@stage\b(.*)$`)

// Parse the global functions each hook of source runs, in source
// order. A ---@stage phase/stage comment hooks the global function it
// precedes, other comments aside; a function may have several:
//
//	---@stage http:request/sec_rule
//	---@stage http:response/router
//	function check(ctx) end
func Parse(source string, chunk []ast.Stmt) (Set, error) {
	funcs := globalFuncs(chunk)
	hooks := make(map[hook][]string)

	var p pending
	for i, line := range strings.Split(source, "\n") {
		s := strings.TrimSpace(line)
		if s != "" && !strings.HasPrefix(s, "--") {
			if err := p.bind(hooks, funcs[i+1]); err != nil {
				return Set{}, err
			}

			continue
		}

		h, ok, err := parseStage(s)
		if err != nil {
			return Set{}, fmt.Errorf("line %d: %w", i+1, err)
		}

		if ok {
			p.add(h, i+1)
		}
	}

	return Set{funcs: hooks}, p.bind(hooks, "")
}

// pending the hooks of the comments above a line of code.
type pending struct {
	hooks []hook
	line  int // of the first of them, 1-based
}

func (p *pending) add(h hook, line int) {
	if len(p.hooks) == 0 {
		p.line = line
	}

	p.hooks = append(p.hooks, h)
}

// parseStage the hook of the comment s, false when it is no @stage.
func parseStage(s string) (hook, bool, error) {
	m := _annotation.FindStringSubmatch(s)
	if m == nil {
		return hook{}, false, nil
	}

	ph, name, ok := strings.Cut(strings.TrimSpace(m[1]), "/")
	if !ok || strings.ContainsAny(name, " \t") {
		return hook{}, false, fmt.Errorf("@stage takes phase/stage, "+
			"e.g. http:request/sec_rule, got %q", strings.TrimSpace(m[1]))
	}

	h := hook{phase: phase.Request}
	switch ph {
	case "http:request":
	case "http:response":
		h.phase = phase.Response
	default:
		return hook{}, false, fmt.Errorf(
			"@stage: phase %q, http:request or http:response", ph)
	}

	i := slices.IndexFunc(_stages, func(s stage) bool {
		return s.name == name
	})
	if i < 0 {
		return hook{}, false, fmt.Errorf(
			"@stage: unknown stage %q, one of %s", name,
			strings.Join(StageNames(), ", "))
	}

	h.stage = _stages[i].id

	return h, true, nil
}

// bind the pending hooks to the global function fn, "" when the code
// after them is no global function, and clears them.
func (p *pending) bind(hooks map[hook][]string, fn string) error {
	if len(p.hooks) == 0 {
		return nil
	}

	if fn == "" {
		return fmt.Errorf("line %d: @stage must precede a global "+
			"function: function name(ctx)", p.line)
	}

	for _, h := range p.hooks {
		if !slices.Contains(hooks[h], fn) {
			hooks[h] = append(hooks[h], fn)
		}
	}

	p.hooks = nil

	return nil
}

// globalFuncs the global functions defined at the top of chunk, by the
// line of their function keyword.
func globalFuncs(chunk []ast.Stmt) map[int]string {
	funcs := make(map[int]string)

	for _, stmt := range chunk {
		def, ok := stmt.(*ast.FuncDefStmt)
		if !ok || def.Name.Receiver != nil {
			continue
		}

		if id, ok := def.Name.Func.(*ast.IdentExpr); ok {
			funcs[def.Line()] = id.Value
		}
	}

	return funcs
}
