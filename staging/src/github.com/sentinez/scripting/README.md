# github.com/sentinez/scripting

> ⚠️ **This is an automatically published [staged repository](https://github.com/sentinez/scripting) for Sentinez**.   
> Contributions, including issues and pull requests, should be made to the main Sentinez repository: [https://github.com/sentinez/sentinez](https://github.com/sentinez/sentinez).  
> This repository is read-only for importing, and not used for direct contributions.

=======================================================================

This module runs the Lua scripts of the edge in a sandbox.

`scripting` compiles a script (`Compile`) and calls its hooks (`Program.Call`).
A `---@stage <phase>/<stage>` line hooks the global function it
precedes, whatever its name, to a stage of the edge handler chain, as
`SEC("fentry/...")` does in eBPF:

```lua
---@stage http.request/secrule
---@stage http.request/secruleset
---@param ctx sz.Context
function deny_debug(ctx)
  if ctx:header("X-Debug") == "1" then
    ctx:respond(403, "no")      -- ends the request, http.request only
  end
end

---@stage http.request/router
---@stage http.response/router
---@param ctx sz.Context
function tag(ctx)
  ctx:set_header("X-Edge", "sentinez")
end
```

- `http.request/<stage>`: runs before the stage; `ctx:respond` ends the request
  there and the stage is skipped.
- `http.response/<stage>`: runs once the stage, and the ones after it, returned.
  The response is written: it only observes.
- One stage and phase per line; a function may have several lines.
- The lines must precede a global function (`function name(ctx)`), other
  comments aside, or the script does not compile; so does a line of an
  unknown phase or stage. A function without one is not a hook, whatever
  its name.
- The hooks of a stage run in source order, until one responds.

The stages, in the order of the edge chain
(`internal/handlers/handlers.go`):

| Stage (plugin id) | Plugin |
|---|---|
| `caching` | `contents.Caching` |
| `secratelimit` | `secure.SecLimiter` |
| `waitingroom` | `room.WaitingRoom` |
| `staticassets` | `contents.StaticAssets` |
| `secrule` | `secure.SecRule` |
| `secruleset` | `secure.SecRuleset` (OWASP CRS) |
| `router` | `routing.StandardRouter`, to the upstream |

The stages before the namespace of the request is known
(`tracing`, `logging`, `domain`) have no hooks.
`scripting.Stages` returns the ids, the ones `Has` and `Call` take; a test
of the edge (`TestScriptStagesAreTheChain`) checks them against the chain.

| `ctx` | |
|---|---|
| `method()`, `path()`, `host()`, `ip()` | the request |
| `header(k)`, `query(k)` | a header or query parameter, "" if absent |
| `has_header(k)`, `has_query(k)` | whether the request has it, even empty |
| `body()`, `ja4()` | the request body, the JA4 fingerprint ("" without TLS) |
| `tls()` | whether the request came over TLS |
| `ip_in(cidr)` | whether the client IP is in the prefix, e.g. `"10.0.0.0/8"` |
| `set_header(k, v)` | a header of the request sent upstream |
| `set_response_header(k, v)` | a header of the response |
| `respond(status, body)` | answers the request (100 to 599) |

The global table `log` writes lines to the logs of the request, with its id
and the name of the script: the arguments, joined by tabs like `print`.

| `log` | |
|---|---|
| `log.info(...)` | a line at the info level |
| `log.debug(...)` | a line at the debug level, written only when the edge logs debug |

`log:info(...)` and `log:debug(...)` work too: called with a colon, the
`log` table passed as first argument is left out.

The global table `re` has `re.match(s, pattern)`: whether `s` holds a match
of `pattern`, a regular expression of the RE2 syntax of Go (not a Lua
pattern); an invalid one raises an error. The first 256 patterns stay
compiled.

```lua
---@stage http.request/secrule
function audit(ctx)
  log.debug("path", ctx:path(), "ip", ctx:ip())
  if ctx:header("X-Debug") == "1" then
    log:info("denied debug request from", ctx:ip())
    ctx:respond(403, "no")
  end
end
```

A line is cut at 1024 bytes and a hook call logs 16 lines at most. In the
script body, outside of a hook, there is no request: the line is dropped.

## Rule scripts

A secrule or a rate limiter of the edge may be a Lua source in place of its
expression. It has no `---@stage` line: it defines one global function, with
a read-only `ctx` (`set_header`, `set_response_header` and `respond` raise
an error). The edge does not call it: it converts the script to its
expression when it loads (`ParseMatch`/`ParseKey`, below), so the script
must be of the declarative form.

A secrule defines `match(ctx)`. It returns `nil` or `false` when the request
does not match, else the action, as an eBPF program returns `XDP_DROP`:

```lua
---@param ctx sz.Context
function match(ctx)
  if ctx:header("X-Debug") == "1" then
    return SZ_BLOCK
  end
  if ctx:path() == "/old" then
    return {type = SZ_REDIRECT, params = {location = "/new"}}
  end
end
```

| Constant | Value |
|---|---|
| `SZ_BLOCK` | `"block"` |
| `SZ_CHALLENGE` | `"challenge"` |
| `SZ_LOG` | `"log"` |
| `SZ_MODIFY_HEADER` | `"modify_header"` |
| `SZ_REDIRECT` | `"redirect"` |
| `SZ_ROUTE_TO` | `"route_to"` |
| `SZ_SET_TAG` | `"set_tag"` |

The constants hold the action names, so `return "block"` is the same. A
table `{type = SZ_REDIRECT, params = {...}}` adds the params of the action.

A rate limiter defines `key(ctx)`: the key to count the request under (a
string), `true` for the client IP, `nil` or `false` not to count it. The
edge holds the declarative form only, `return true`: it counts under the
client IP.

```lua
function key(ctx)
  local k = ctx:header("X-Api-Key")
  if k ~= "" then return k end
end
```

A host that runs the rule as Lua compiles it with `Compile`, checks it with
`Hooked` (false) and `HasFunc(scripting.MatchFunc)` (or `KeyFunc`), then
calls `Program.Invoke(ctx, "match", ex)` for each request: the result is a
`Value` (`KindNil`, `KindBool`, `KindString`, `KindTable`) free of the Lua
types.

`ParseMatch` and `ParseKey` read a script of the declarative form, branches
of `if <conditions> then return <action> end`, into a `Rule`; `MatchSource`
and `KeySource` write a `Rule` back as Lua (`return SZ_BLOCK`). The
condition of a `Rule` is a `Node` tree: a `Leaf` of one `Cond`, or a group
of `All` or `Any` of its nodes (`MaxConditions` conditions at most, no group
multiplied out). The builders keep the canonical form a script is read to:
nested groups of one kind merged, conditions before groups.
`CondSource` writes the condition alone, with the parentheses an `or` in an
`and` needs. A condition
combines with `and`, `or`, `not` and parentheses the comparisons below,
`field` one of `ctx:path()`, `ctx:method()`, `ctx:host()`, `ctx:ip()`,
`ctx:header("k")`, `ctx:query("k")`, `ctx:body()`, `ctx:ja4()`:

| Lua | `Cond` |
|---|---|
| `field == "v"`, `field ~= "v"` | `OpEq`, `OpNe` |
| `field:find("v", 1, true)` | `OpContains` |
| `field:sub(1, #v) == "v"`, `field:sub(-#v) == "v"` | `OpPrefix`, `OpSuffix` |
| `tonumber(field) > 10` (`>=`, `<`, `<=`) | `OpGt`, `OpGte`, `OpLt`, `OpLte` |
| `re.match(field, "^/a")` | `OpMatches` |
| `ctx:ip_in("10.0.0.0/8")`, `not ctx:ip_in(...)` | `OpWithin`, `OpNotWithin` |
| `ctx:has_header("k")`, `ctx:has_query("k")`, with `not` | `OpExists`, `OpAbsent` |
| `ctx:tls()`, `not ctx:tls()` | `SourceTLS`, `OpEq`/`OpNe` `"true"` |

`not` applies to the equalities, `ip_in`, `has_*` and `tls` only. A script
out of that form returns `ErrNotDeclarative`. The edge converts its protobuf expressions with them
(`core/szrule/rulescript`), and its rule scripts to expressions: it refuses
a rule script out of that form.

## Sandbox

The sandbox opens `base`, `table`, `string` and `math` only: there is no
`os`, `io`, `load` or `require`. Each call is limited to 50ms (`WithTimeout`),
and a script that fails or times out is skipped by the edge.

## Usage

```sh
go get github.com/sentinez/scripting
```

```go
prog, err := scripting.Compile("audit.lua", source,
	scripting.WithTimeout(20*time.Millisecond))
if err != nil {
	return err // syntax error, misplaced ---@stage, or a failing script body
}

// around the stage whose id is "secrule" (secure.SecRule):
if prog.Has(phase.Request, "secrule") {
	responded, err := prog.Call(ctx, phase.Request, "secrule", ex)
	// responded: a hook answered, skip the stage
}
```

- `Compile` parses the script, reads its `---@stage` lines and runs the body
  once, so a broken script fails there, not at the first request.
- A `Program` is safe for concurrent use: each call takes a Lua state of its
  own, kept for reuse unless the call failed or timed out.
- `Call` runs the hooks of the stage in source order until one responds, and
  reports whether one did. The hooks share the time limit (50ms by default).
- `Has` is a cheap check to skip the call when the script hooks nothing there.

The host implements `exchange.Exchange` (the request, `SetHeader`,
`SetResponseHeader`, `String` to answer it, `Log`) and calls
`Program.Call(ctx, phase.Request, stage, ex)` around each stage.

## Packages

| Package | |
|---|---|
| `scripting` | `Compile`, `Program` (`Has`, `Call`, `Invoke`, `HasFunc`, `Hooked`, `Name`, `Source`), `Value`, `WithTimeout` and `Stages`; the rule forms `Rule`, `ParseMatch`, `MatchSource`, `ParseKey`, `KeySource` |
| `scripting/exchange` | `Exchange`, the request the host hands a hook, and the log `Level` |
| `scripting/phase` | `Phase`: `Request` (`http.request`), `Response` (`http.response`) |
| `internal/hooks` | parses the `---@stage` lines into the hooks of a script; the stage table |
| `internal/api` | the `ctx` and `log` APIs and the `SZ_` action constants; `Run` calls the hooks of a stage in a Lua state, `Invoke` the function of a rule |
| `internal/sandbox` | the Lua state: the libraries a script may use, no more |

## Editor support

`meta/sentinez.lua` declares `ctx` (`sz.Context`), `log`, the stages
(`sz.Stage`), the action constants (`sz.Action`) and the rule functions
(`sz.Match`, `sz.Key`) for lua-language-server
(LuaLS / EmmyLua): annotate a hook with `---@param ctx sz.Context` for completion
and type checks. `.luarc.json` loads it for this directory; elsewhere add the
directory to `workspace.library`. `TestMetaMatchesContext` keeps it in step with
the methods of `ctx`, the `log` functions, the stages and the actions.
