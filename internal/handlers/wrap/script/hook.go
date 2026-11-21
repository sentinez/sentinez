// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package script runs the Lua scripts of a namespace around the stages of
// the edge chain.
package script

import (
	"slices"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/scripting"
	"github.com/sentinez/scripting/phase"
	"github.com/sentinez/sentinez/internal/memory/scripts"
	"github.com/sentinez/sentinez/pkg/pools/exchange"
)

var _ chains.WrapFunc = (*Hooks)(nil).Wrap

// Hooks wraps the handlers of the plugins to call the hooks of the
// scripts of the namespace of the request.
type Hooks struct {
	store *scripts.Scripts
}

func New(store *scripts.Scripts) *Hooks {
	return &Hooks{store: store}
}

// Wrap returns h, the handler running the plugin id, with the hooks of
// the stage id around it. A stage scripts may not hook (scripting.Stages) is
// returned as is.
//
// The hooks of ---@stage http:request/<stage> run before the plugin: a
// script answering the request ends it there. The ones of
// http:response/<stage> run when the plugin, and the ones after it,
// returned: the response is written, they only observe.
func (h *Hooks) Wrap(id string, next chains.Handler) chains.Handler {
	if !slices.Contains(scripting.Stages(), id) {
		return next
	}

	return &hooked{store: h.store, stage: id, next: next}
}

type hooked struct {
	store *scripts.Scripts
	stage string
	next  chains.Handler
}

func (h *hooked) Handle(ctx szhttp.Context) error {
	programs := h.store.LoadContext(ctx)
	if !h.hooks(programs) {
		return h.next.Handle(ctx)
	}

	ex := exchange.Acquire(ctx)
	defer exchange.Release(ex)

	if h.run(ctx, programs, phase.Request, ex) {
		return nil
	}

	err := h.next.Handle(ctx)
	h.run(ctx, programs, phase.Response, ex)

	return err
}

// hooks reports whether a script of programs hooks the stage, at either
// phase: the exchange of the hooks is only acquired then.
func (h *hooked) hooks(programs []*scripting.Program) bool {
	for _, p := range programs {
		if p.Has(phase.Request, h.stage) || p.Has(phase.Response, h.stage) {
			return true
		}
	}

	return false
}

// run calls the hooks of each script at phase, and reports whether one
// answered the request. A script that fails is skipped: the request goes
// on.
func (h *hooked) run(ctx szhttp.Context, programs []*scripting.Program,
	at phase.Phase, ex *exchange.Exchange) bool {

	for _, p := range programs {
		responded, err := p.Call(ctx.Context(), at, h.stage, ex)
		if err != nil {
			szlog.Warnf("script: request %s: %v", ctx.RequestId(), err)

			continue
		}

		if responded {
			return true
		}
	}

	return false
}
