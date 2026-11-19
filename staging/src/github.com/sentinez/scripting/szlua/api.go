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
	"strings"

	lua "github.com/yuin/gopher-lua"
)

const (
	_ctxType = "szlua.ctx"

	// Longest log line of a script, in bytes: the rest is cut.
	_maxLogLen = 1024
	// Most log lines of a hook call: the others are dropped.
	_maxLogLines = 16
)

// call the state of one hook call, the Value of the ctx userdata and of
// the context of the Lua state, for the log table.
type call struct {
	ex        Exchange
	script    string
	phase     Phase
	responded bool
	logged    int
}

// callKey the key of the call in the context of the Lua state.
type callKey struct{}

// _logFuncs the functions of the log table, one by level, as
// meta/sentinez.lua declares them.
var _logFuncs = map[string]Level{
	LevelDebug.String(): LevelDebug,
	LevelInfo.String():  LevelInfo,
}

// _ctxMethods the methods of ctx, as meta/sentinez.lua declares them.
var _ctxMethods = map[string]lua.LGFunction{
	"method":              reader(method),
	"path":                reader(path),
	"host":                reader(host),
	"ip":                  reader(ip),
	"header":              keyed(header),
	"query":               keyed(query),
	"set_header":          setter(setHeader),
	"set_response_header": setter(setResponseHeader),
	"respond":             respond,
}

// registerCtx adds the ctx type to L.
func registerCtx(L *lua.LState) {
	mt := L.NewTypeMetatable(_ctxType)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), _ctxMethods))
}

func method(c *call) []byte { return c.ex.Method() }
func path(c *call) []byte   { return c.ex.Path() }
func host(c *call) []byte   { return c.ex.Host() }
func ip(c *call) []byte     { return c.ex.RequestIP() }

func header(c *call, k string) []byte { return c.ex.Header([]byte(k)) }
func query(c *call, k string) []byte  { return c.ex.Query([]byte(k)) }

func setHeader(c *call, k, v string)         { c.ex.SetHeader(k, v) }
func setResponseHeader(c *call, k, v string) { c.ex.SetResponseHeader(k, v) }

// registerLog adds the global table log to L: log.debug(...) and
// log.info(...), or log:debug(...) and log:info(...).
func registerLog(L *lua.LState) {
	tbl := L.NewTable()
	for name, level := range _logFuncs {
		L.SetField(tbl, name, L.NewClosure(logAt(level), tbl))
	}

	L.SetGlobal("log", tbl)
}

// logAt a function of the log table: its arguments, as tostring prints
// them, joined by tabs as print does. Called with a colon, the first
// argument is the log table, its upvalue: it is left out. Outside of a
// hook call, while the script body runs, the line is dropped: there is no
// request to log it with.
func logAt(level Level) lua.LGFunction {
	return func(L *lua.LState) int {
		c := currentCall(L.Context())
		if c == nil || c.logged >= _maxLogLines {
			return 0
		}

		first := 1
		if L.GetTop() > 0 && L.Get(1) == L.Get(lua.UpvalueIndex(1)) {
			first = 2
		}

		c.logged++
		c.ex.Log(level, c.script, logLine(L, first))

		return 0
	}
}

func currentCall(ctx context.Context) *call {
	if ctx == nil {
		return nil
	}

	c, _ := ctx.Value(callKey{}).(*call)

	return c
}

// logLine the arguments of L from first on.
func logLine(L *lua.LState, first int) string {
	var b strings.Builder
	for i := first; i <= L.GetTop(); i++ {
		if i > first {
			b.WriteByte('\t')
		}

		b.WriteString(L.ToStringMeta(L.Get(i)).String())
		if b.Len() > _maxLogLen {
			break
		}
	}

	s := b.String()
	if len(s) > _maxLogLen {
		s = strings.ToValidUTF8(s[:_maxLogLen], "")
	}

	return s
}

func callOf(L *lua.LState) *call {
	ud := L.CheckUserData(1)
	c, ok := ud.Value.(*call)
	if !ok {
		L.ArgError(1, "ctx expected")
	}

	return c
}

func reader(get func(c *call) []byte) lua.LGFunction {
	return func(L *lua.LState) int {
		L.Push(lua.LString(get(callOf(L))))

		return 1
	}
}

func keyed(get func(c *call, k string) []byte) lua.LGFunction {
	return func(L *lua.LState) int {
		c := callOf(L)
		L.Push(lua.LString(get(c, L.CheckString(2))))

		return 1
	}
}

func setter(set func(c *call, k, v string)) lua.LGFunction {
	return func(L *lua.LState) int {
		c := callOf(L)
		set(c, L.CheckString(2), L.CheckString(3))

		return 0
	}
}

// respond ctx:respond(status, body) answers the request. An after hook
// runs once the stage answered: it raises an error.
func respond(L *lua.LState) int {
	c := callOf(L)
	if c.phase == After {
		L.RaiseError("respond: the stage already ran")
	}

	status := L.CheckInt(2)
	if status < 100 || status > 599 {
		L.ArgError(2, "status must be 100 to 599")
	}

	if err := c.ex.String(status, []byte(L.OptString(3, ""))); err != nil {
		L.RaiseError("respond: %v", err)
	}

	c.responded = true

	return 0
}
