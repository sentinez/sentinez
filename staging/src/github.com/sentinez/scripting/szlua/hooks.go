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

package szlua

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/yuin/gopher-lua/ast"
)

// Phase is when a hook runs, relative to its stage.
type Phase int

const (
	// Before the hook runs before the stage, and may answer the
	// request: the stage is then skipped.
	Before Phase = iota + 1
	// After the hook runs once the stage, and the ones after it,
	// returned: the response is written, the hook only observes.
	After
)

func (p Phase) String() string {
	switch p {
	case Before:
		return "before"
	case After:
		return "after"
	}

	return "Phase(" + strconv.Itoa(int(p)) + ")"
}

// _stages the stages of the edge chain a hook may annotate, by the id of
// their plugin, in chain order: CDN cache, LMT rate limiter, ROM waiting
// room, STC static files, RUL secrules, RST rulesets (OWASP CRS), ROU
// router. The stages before the namespace of the request is known (TRC,
// LOG, DMA) have no hooks. The edge tests it against its chain.
var _stages = []string{"CDN", "LMT", "ROM", "STC", "RUL", "RST", "ROU"}

// Stages returns the stages a hook may annotate.
func Stages() []string { return slices.Clone(_stages) }

// hook a stage and the phase of it a function runs at.
type hook struct {
	phase Phase
	stage string
}

// _annotation ---@module <stage>... or ---@types before|after...
var _annotation = regexp.MustCompile(`^\s*---\s*@(module|types)\b(.*)$`)

// block the annotations of the comments preceding a line of code.
type block struct {
	line   int // of its first annotation, 1-based
	stages []string
	phases []Phase
	// unknown the first @module name that is no stage.
	unknown string
}

// parseHooks the global functions each hook of source runs, in source
// order. The comments preceding a global function name the stages it
// hooks with @module and when with @types, other comments aside:
//
//	---@module RUL
//	---@module 'RST'
//	---@types before
//	function check(ctx) end
//
// A name may be quoted, as LuaLS writes it. A @module naming no
// stage, without @types, is the LuaLS one: it is left alone.
func parseHooks(source string,
	chunk []ast.Stmt) (map[hook][]string, error) {

	funcs := globalFuncs(chunk)
	hooks := make(map[hook][]string)

	var b block
	for i, line := range strings.Split(source, "\n") {
		s := strings.TrimSpace(line)
		if s == "" || strings.HasPrefix(s, "--") {
			if err := b.read(s, i+1); err != nil {
				return nil, err
			}

			continue
		}

		if err := b.bind(hooks, funcs[i+1]); err != nil {
			return nil, err
		}

		b = block{}
	}

	return hooks, b.bind(hooks, "")
}

// read adds the annotation of the comment s, of line n, to b.
func (b *block) read(s string, n int) error {
	m := _annotation.FindStringSubmatch(s)
	if m == nil {
		return nil
	}

	names := strings.Fields(strings.NewReplacer(`'`, " ", `"`, " ").
		Replace(m[2]))
	if len(names) == 0 {
		return fmt.Errorf("line %d: @%s names nothing", n, m[1])
	}

	if b.line == 0 {
		b.line = n
	}

	if m[1] == "types" {
		return b.readPhases(names, n)
	}

	for _, name := range names {
		if slices.Contains(_stages, name) {
			b.stages = append(b.stages, name)
		} else if b.unknown == "" {
			b.unknown = name
		}
	}

	return nil
}

func (b *block) readPhases(names []string, n int) error {
	for _, name := range names {
		switch name {
		case "before":
			b.phases = append(b.phases, Before)
		case "after":
			b.phases = append(b.phases, After)
		default:
			return fmt.Errorf("line %d: @types %q: before or after",
				n, name)
		}
	}

	return nil
}

// bind the hooks of b to the global function fn, "" when the code after b
// is no global function.
func (b *block) bind(hooks map[hook][]string, fn string) error {
	if len(b.phases) == 0 && len(b.stages) == 0 {
		return nil
	}

	if err := b.check(fn); err != nil {
		return fmt.Errorf("line %d: %w", b.line, err)
	}

	for _, p := range b.phases {
		for _, c := range b.stages {
			h := hook{phase: p, stage: c}
			if !slices.Contains(hooks[h], fn) {
				hooks[h] = append(hooks[h], fn)
			}
		}
	}

	return nil
}

func (b *block) check(fn string) error {
	switch {
	case len(b.phases) == 0:
		return fmt.Errorf("@module %s needs @types before or after",
			strings.Join(b.stages, " "))
	case b.unknown != "":
		return fmt.Errorf("@module: unknown stage %q, one of %s",
			b.unknown, strings.Join(_stages, ", "))
	case len(b.stages) == 0:
		return fmt.Errorf("@types needs a @module stage, one of %s",
			strings.Join(_stages, ", "))
	case fn == "":
		return errors.New("a hook must be a global function: " +
			"function name(ctx)")
	}

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
