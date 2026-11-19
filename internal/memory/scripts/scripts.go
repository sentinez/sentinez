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
	"fmt"
	"slices"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/pkg/szsync"
	"github.com/sentinez/scripting/szlua"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
)

// Scripts the compiled scripts by namespace.
type Scripts struct {
	sets *szsync.Map[string, []*szlua.Program]
}

func New() *Scripts {
	return &Scripts{sets: szsync.NewMap[string, []*szlua.Program]()}
}

// Store compiles the enabled scripts of namespace, in their given order,
// and replaces what it held. A script whose name and source did not change
// is kept as compiled, with its warm Lua states. A script that does not
// compile fails the whole call and leaves the previous scripts in place.
func (s *Scripts) Store(namespace string, scripts []*edgepb.Script) error {
	held := s.Load(namespace)
	programs := make([]*szlua.Program, 0, len(scripts))

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
func compiled(held []*szlua.Program, name,
	source string) (*szlua.Program, error) {

	i := slices.IndexFunc(held, func(p *szlua.Program) bool {
		return p.Name() == name && p.Source() == source
	})
	if i >= 0 {
		return held[i], nil
	}

	return szlua.Compile(name, source)
}

// Load returns the scripts of namespace. The result is shared and must not
// be modified.
func (s *Scripts) Load(namespace string) []*szlua.Program {
	programs, _ := s.sets.Load(namespace)

	return programs
}

// LoadContext returns the scripts of the namespace of the request.
func (s *Scripts) LoadContext(ctx szhttp.Context) []*szlua.Program {
	if s == nil {
		return nil
	}

	return s.Load(ctx.X().GetNamespace())
}

// Delete removes what namespace holds.
func (s *Scripts) Delete(namespace string) {
	s.sets.Delete(namespace)
}
