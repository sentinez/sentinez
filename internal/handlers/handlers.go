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

// Package handlers builds the edge chain: its plugins, each wrapped by a
// stage (errors tagged with the plugin id, panics recovered) and by the Lua
// hooks of the namespace of the request. Both wraps are private: the chain
// is only built here.
package handlers

import (
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/internal/handlers/contents"
	"github.com/sentinez/sentinez/internal/handlers/logging"
	"github.com/sentinez/sentinez/internal/handlers/room"
	"github.com/sentinez/sentinez/internal/handlers/routing"
	"github.com/sentinez/sentinez/internal/handlers/secure"
	"github.com/sentinez/sentinez/internal/handlers/tracing"
	"github.com/sentinez/sentinez/internal/memory"
)

// New builds the edge chain: the plugins, in request order, provision
// themselves from mem and opts (see option.WithLevel, option.WithExporter,
// option.WithSecretKey).
func New(mem *memory.MemStore, opts ...option.Option) (*chains.Chain, error) {
	hooks := newScriptHooks(mem.Scripts())
	wrap := func(id string, h chains.Handler) chains.Handler {
		return wrapStage(id, hooks.wrap(id, h))
	}

	return chains.Build(envplugin.NewEnv(mem, opts...), wrap, plugins()...)
}

// plugins returns the edge chain plugins, in request order.
func plugins() []chains.Plugin {
	return []chains.Plugin{
		&tracing.Tracer{},
		&logging.Logging{},
		&secure.DomainBased{},
		&contents.Caching{},
		&secure.SecLimiter{},
		&room.WaitingRoom{},
		&contents.StaticAssets{},
		&secure.SecRule{},
		&secure.SecRuleset{},
		&routing.StandardRouter{},
	}
}
