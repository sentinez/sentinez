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

package api

import (
	"maps"
	"regexp"
	"slices"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// MaxRegexps the most patterns re keeps compiled: the others compile on
// each call.
const MaxRegexps = 256

// _reFuncs the functions of the re table, as meta/sentinez.lua declares
// them.
var _reFuncs = map[string]lua.LGFunction{
	"match": reMatch,
}

// ReFuncs returns the sorted names of the functions of the re table.
func ReFuncs() []string { return slices.Sorted(maps.Keys(_reFuncs)) }

// registerRe adds the global table re to L: re.match(s, pattern).
func registerRe(L *lua.LState) {
	L.SetGlobal("re", L.SetFuncs(L.NewTable(), _reFuncs))
}

// reMatch re.match(s, pattern) reports whether s holds a match of the
// regular expression pattern, of the RE2 syntax of Go (not a Lua
// pattern).
func reMatch(L *lua.LState) int {
	s := L.CheckString(1)

	re, err := _regexps.get(L.CheckString(2))
	if err != nil {
		L.ArgError(2, err.Error())
	}

	L.Push(lua.LBool(re.MatchString(s)))

	return 1
}

var _regexps = &regexps{m: make(map[string]*regexp.Regexp)}

// regexps the compiled patterns, MaxRegexps at most.
type regexps struct {
	mu sync.RWMutex
	m  map[string]*regexp.Regexp
}

func (r *regexps) get(pattern string) (*regexp.Regexp, error) {
	r.mu.RLock()
	re, ok := r.m[pattern]
	r.mu.RUnlock()
	if ok {
		return re, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	if len(r.m) < MaxRegexps {
		r.m[pattern] = re
	}
	r.mu.Unlock()

	return re, nil
}
