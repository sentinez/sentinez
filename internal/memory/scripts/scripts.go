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
	"errors"
	"fmt"
	"slices"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/pkg/szsync"
	"github.com/sentinez/scripting"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
)

var errRuleStage = errors.New("a rule script has no @stage hook")

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

		name := sc.GetName()
		if name == "" {
			name = sc.GetId()
		}

		p, err := compiled(held, name, sc.GetSource())
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

// Rule the program of the script of r, named by its name or else its id:
// compiled, checked to define the global function fn, its entry point,
// and to hook no stage.
func Rule(held []*scripting.Program, r *secrulepb.SecRule,
	fn string) (*scripting.Program, error) {

	name := r.GetName()
	if name == "" {
		name = r.GetId()
	}

	p, err := compiled(held, name, r.GetScript())
	if err != nil {
		return nil, err
	}

	if p.Hooked() {
		return nil, errRuleStage
	}

	if !p.HasFunc(fn) {
		return nil, fmt.Errorf("script must define %s(ctx)", fn)
	}

	return p, nil
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
