// Copyright 2026 Sentinéz Labs.
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

package funcs

import (
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/internal/funcs/cdn"
	"github.com/sentinez/sentinez/internal/funcs/logging"
	"github.com/sentinez/sentinez/internal/funcs/room"
	"github.com/sentinez/sentinez/internal/funcs/routing"
	"github.com/sentinez/sentinez/internal/funcs/secure"
	"github.com/sentinez/sentinez/internal/funcs/static"
	"github.com/sentinez/sentinez/internal/funcs/trace"
	"github.com/sentinez/sentinez/internal/memory"
)

// New builds the edge chain: the plugins, in request order, provision
// themselves from mem and opts (see option.WithLevel, option.WithExporter,
// option.WithHostname).
func New(
	mem *memory.MemStore, opts ...option.Option) (*chains.Chain, error) {

	return chains.Build(
		envplugin.NewEnv(mem, opts...), Stage, plugins()...)
}

// plugins returns the edge chain plugins, in request order.
func plugins() []chains.Plugin {
	return []chains.Plugin{
		&trace.Tracer{},
		&logging.Logging{},
		&secure.DomainBased{},
		&cdn.Public{},
		&cdn.Cache{},
		&secure.Limiter{},
		&room.WaitingRoom{},
		&static.Static{},
		&secure.SecRule{},
		&secure.Rulesets{},
		&routing.StandardRouter{},
	}
}
