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

// Package scripts holds the compiled Lua scripts of each namespace.
package scripts

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szrule/rulescript"
	"github.com/sentinez/pkg/szsync"
	"github.com/sentinez/scripting"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"google.golang.org/protobuf/proto"
)

var (
	// ErrExprAndScript a rule has an expr and a script.
	ErrExprAndScript = errors.New("expr and script are exclusive")
)

// Scripts the compiled scripts by namespace.
type Scripts struct {
	sets *szsync.Map[string, []*scripting.Program]
}

func New() *Scripts {
	return &Scripts{sets: szsync.NewMap[string, []*scripting.Program]()}
}

// Store compiles the enabled scripts of namespace, in their given order,
// and replaces what it held. A script whose name and source did not change
// is kept as compiled, with its warm Lua states. A script that does not
// compile fails the whole call and leaves the previous scripts in place.
func (s *Scripts) Store(namespace string, scripts []*edgepb.Script) error {
	held := s.Load(namespace)
	programs := make([]*scripting.Program, 0, len(scripts))

	for _, sc := range scripts {
		if sc.GetDisabled() {
			continue
		}

		p, err := compiled(held, cmp.Or(sc.GetName(), sc.GetId()),
			sc.GetSource())
		if err != nil {
			return fmt.Errorf("scripts: %w", err)
		}

		programs = append(programs, p)
	}

	if len(programs) == 0 {
		s.sets.Delete(namespace)

		return nil
	}

	s.sets.Store(namespace, programs)

	return nil
}

// compiled the program of held with name and source, or a new one.
func compiled(held []*scripting.Program, name,
	source string) (*scripting.Program, error) {

	i := slices.IndexFunc(held, func(p *scripting.Program) bool {
		return p.Name() == name && p.Source() == source
	})
	if i >= 0 {
		return held[i], nil
	}

	return scripting.Compile(name, source)
}

// Load returns the scripts of namespace. The result is shared and must not
// be modified.
func (s *Scripts) Load(namespace string) []*scripting.Program {
	programs, _ := s.sets.Load(namespace)

	return programs
}

// LoadContext returns the scripts of the namespace of the request.
func (s *Scripts) LoadContext(ctx szhttp.Context) []*scripting.Program {
	if s == nil {
		return nil
	}

	return s.Load(ctx.X().GetNamespace())
}

// Delete removes what namespace holds.
func (s *Scripts) Delete(namespace string) {
	s.sets.Delete(namespace)
}

// Rule r with its secrule script converted by rulescript.FromScript to the
// expr and action it declares: the runtime evaluates every rule by szrule
// and runs no Lua. r itself when it has no script. A rule with an expr too
// fails with ErrExprAndScript, a script outside of the declarative form
// with an error wrapping scripting.ErrNotDeclarative.
func Rule(r *secrulepb.SecRule) (*secrulepb.SecRule, error) {
	if err := checkScript(r); err != nil {
		return nil, err
	}

	if r.GetScript() == "" {
		return r, nil
	}

	expr, action, err := rulescript.FromScript(r.GetScript())
	if err != nil {
		return nil, err
	}

	return converted(r, expr, action), nil
}

// KeyRule r with its rate limiter script converted like Rule, by
// rulescript.FromKeyScript: the requests its expr matches are counted
// under the client IP.
func KeyRule(r *secrulepb.SecRule) (*secrulepb.SecRule, error) {
	if err := checkScript(r); err != nil {
		return nil, err
	}

	if r.GetScript() == "" {
		return r, nil
	}

	expr, err := rulescript.FromKeyScript(r.GetScript())
	if err != nil {
		return nil, err
	}

	return converted(r, expr, r.GetAction()), nil
}

func checkScript(r *secrulepb.SecRule) error {
	if r.GetScript() != "" && len(r.GetExpr().GetOrCondition()) > 0 {
		return ErrExprAndScript
	}

	return nil
}

// converted a copy of r with expr and action in place of its script.
func converted(r *secrulepb.SecRule, expr *rulepb.Expression,
	action *rulepb.Action) *secrulepb.SecRule {

	c := proto.CloneOf(r)
	c.Script, c.Expr, c.Action = "", expr, action

	return c
}
