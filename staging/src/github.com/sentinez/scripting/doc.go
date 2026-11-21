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

// Package scripting runs the Lua scripts of the edge. A ---@stage line hooks
// the global function it precedes, whatever its name, to a stage of the
// edge handler chain, at http:request or http:response:
//
//	---@stage http:request/sec_rule
//	function deny_debug(ctx)
//	  if ctx:header("X-Debug") == "1" then ctx:respond(403) end
//	end
//
// The edge calls it around the stage with a ctx value
// to read the request, change its headers or answer it. Scripts run in a
// sandbox without os, io or any way to load code, under a time limit.
package scripting
