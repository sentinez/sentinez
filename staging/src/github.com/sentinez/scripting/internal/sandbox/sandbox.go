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

package sandbox

import (
	lua "github.com/yuin/gopher-lua"
)

const (
	_callStackSize = 64
	_registrySize  = 1024
	_registryMax   = 8192
)

// _removed the functions of the base library that load code, read files
// or reach outside of the script.
var _removed = []string{
	"dofile", "loadfile", "load", "loadstring", "require", "module",
	"collectgarbage", "getfenv", "setfenv", "newproxy",
}

// NewState a state with the base, table, string and math libraries only.
func NewState() *lua.LState {
	L := lua.NewState(lua.Options{
		SkipOpenLibs:    true,
		CallStackSize:   _callStackSize,
		RegistrySize:    _registrySize,
		RegistryMaxSize: _registryMax,
	})

	for _, lib := range []struct {
		name string
		open lua.LGFunction
	}{
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
	} {
		L.Push(L.NewFunction(lib.open))
		L.Push(lua.LString(lib.name))
		L.Call(1, 0)
	}

	for _, name := range _removed {
		L.SetGlobal(name, lua.LNil)
	}

	return L
}
