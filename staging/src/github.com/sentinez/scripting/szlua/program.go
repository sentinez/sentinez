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
	"context"
	"fmt"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

const (
	_defaultLimit  = 50 * time.Millisecond
	_maxIdleStates = 64
)

// Option configures a Program.
type Option interface{ apply(*Program) }

type optionFunc func(*Program)

func (f optionFunc) apply(p *Program) { f(p) }

// WithTimeout limits the time of each Call, 50ms by default.
func WithTimeout(d time.Duration) Option {
	return optionFunc(func(p *Program) { p.timeout = d })
}

// Program is a compiled script. It is safe for concurrent use: a call
// takes a Lua state of its own.
type Program struct {
	name    string
	source  string
	proto   *lua.FunctionProto
	hooks   map[hook][]string
	timeout time.Duration
	idle    chan *lua.LState
}

// Compile parses source, reads the hooks its annotations declare and runs
// it once. A syntax error, a misplaced annotation or an error of the
// script body is returned here, not at the first request.
func Compile(name, source string, opts ...Option) (*Program, error) {
	p, err := compile(name, source)
	if err != nil {
		return nil, fmt.Errorf("script %s: %w", name, err)
	}

	for _, o := range opts {
		o.apply(p)
	}

	L, err := p.fresh()
	if err != nil {
		return nil, fmt.Errorf("script %s: %w", name, err)
	}

	p.release(L)

	return p, nil
}

func compile(name, source string) (*Program, error) {
	chunk, err := parse.Parse(strings.NewReader(source), name)
	if err != nil {
		return nil, err
	}

	hooks, err := parseHooks(source, chunk)
	if err != nil {
		return nil, err
	}

	proto, err := lua.Compile(chunk, name)
	if err != nil {
		return nil, err
	}

	return &Program{
		name:    name,
		source:  source,
		proto:   proto,
		hooks:   hooks,
		timeout: _defaultLimit,
		idle:    make(chan *lua.LState, _maxIdleStates),
	}, nil
}

// Name returns the name the script was compiled with.
func (p *Program) Name() string { return p.name }

// Source returns the source the script was compiled from.
func (p *Program) Source() string { return p.source }

// Has reports whether a function of the script hooks the stage at
// phase.
func (p *Program) Has(phase Phase, stage string) bool {
	return len(p.hooks[hook{phase: phase, stage: stage}]) > 0
}

// Call runs the functions hooking the stage at phase, in source order,
// until one answers the request, and reports whether one did. They share
// the time limit. An error, a timeout included, stops there and discards
// the Lua state.
func (p *Program) Call(ctx context.Context, phase Phase, stage string,
	ex Exchange) (bool, error) {

	funcs := p.hooks[hook{phase: phase, stage: stage}]
	if len(funcs) == 0 {
		return false, nil
	}

	L, err := p.acquire()
	if err != nil {
		return false, fmt.Errorf("script %s: %w", p.name, err)
	}

	c := &call{ex: ex, script: p.name, phase: phase}
	if err := p.run(ctx, L, funcs, c); err != nil {
		L.Close()

		return false, fmt.Errorf("script %s: %s/%s: %w",
			p.name, phase, stageName(stage), err)
	}

	p.release(L)

	return c.responded, nil
}

// run calls the global functions funcs in L under the time limit, with c
// as their ctx, until one answers the request.
func (p *Program) run(ctx context.Context, L *lua.LState, funcs []string,
	c *call) error {

	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	L.SetContext(context.WithValue(ctx, callKey{}, c))
	defer L.RemoveContext()

	ud := L.NewUserData()
	ud.Value = c
	L.SetMetatable(ud, L.GetTypeMetatable(_ctxType))

	for _, fn := range funcs {
		f, ok := L.GetGlobal(fn).(*lua.LFunction)
		if !ok {
			return fmt.Errorf("%s is not a function anymore", fn)
		}

		err := L.CallByParam(lua.P{Fn: f, Protect: true}, ud)
		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		if c.responded {
			return nil
		}
	}

	return nil
}

// fresh a state with the script body run.
func (p *Program) fresh() (*lua.LState, error) {
	L := newState()

	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	L.SetContext(ctx)
	L.Push(L.NewFunctionFromProto(p.proto))

	if err := L.PCall(0, 0, nil); err != nil {
		L.Close()

		return nil, err
	}

	L.RemoveContext()

	return L, nil
}

func (p *Program) acquire() (*lua.LState, error) {
	select {
	case L := <-p.idle:
		return L, nil
	default:
		return p.fresh()
	}
}

func (p *Program) release(L *lua.LState) {
	select {
	case p.idle <- L:
	default:
		L.Close()
	}
}
