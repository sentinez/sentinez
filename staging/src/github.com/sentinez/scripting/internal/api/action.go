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
	"slices"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// _actions the actions a secrule script may return, by name or by their
// global constant (SZ_BLOCK, as an eBPF program returns XDP_DROP).
var _actions = []string{
	"block", "challenge", "log", "modify_header", "redirect", "route_to",
	"set_tag",
}

// Actions the names of the actions a secrule script may return, sorted.
func Actions() []string { return slices.Clone(_actions) }

// IsAction reports whether a secrule script may return the action name.
func IsAction(name string) bool { return slices.Contains(_actions, name) }

// ActionConst the global constant of the action name: SZ_ and the name in
// upper case. Its value is the name.
func ActionConst(name string) string {
	return "SZ_" + strings.ToUpper(name)
}

// registerActions adds the global constant of every action to L.
func registerActions(L *lua.LState) {
	for _, name := range _actions {
		L.SetGlobal(ActionConst(name), lua.LString(name))
	}
}
