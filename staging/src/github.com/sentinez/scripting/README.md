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
---@stage http:request/sec_rule
---@stage http:request/sec_ruleset
---@param ctx sz.Context
function deny_debug(ctx)
  if ctx:header("X-Debug") == "1" then
    ctx:respond(403, "no")      -- ends the request, http:request only
  end
end

---@stage http:request/router
---@stage http:response/router
---@param ctx sz.Context
function tag(ctx)
  ctx:set_header("X-Edge", "sentinez")
end
```

- `http:request/<stage>`: runs before the stage; `ctx:respond` ends the request
  there and the stage is skipped.
- `http:response/<stage>`: runs once the stage, and the ones after it, returned.
  The response is written: it only observes.
- One stage and phase per line; a function may have several lines.
- The lines must precede a global function (`function name(ctx)`), other
  comments aside, or the script does not compile; so does a line of an
  unknown phase or stage. A function without one is not a hook, whatever
  its name.
- The hooks of a stage run in source order, until one responds.

The stages, in the order of the edge chain
(`internal/handlers/handlers.go`):

| Stage | Id | Plugin |
|---|---|---|
| `cache` | `CDN` | `cdn.Cache` |
| `sec_rate_limit` | `LMT` | `secure.Limiter` |
| `waiting_room` | `ROM` | `room.WaitingRoom` |
| `static` | `STC` | `static.Static` |
| `sec_rule` | `RUL` | `secure.SecRule` |
| `sec_ruleset` | `RST` | `secure.Rulesets` (OWASP CRS) |
| `router` | `ROU` | `routing.StandardRouter`, to the upstream |

The stages before the namespace of the request is known (`TRC`, `LOG`,
`DMA`) have no hooks. `scripting.Stages` returns the ids, the ones `Has` and
`Call` take; a test of the edge (`TestScriptStagesAreTheChain`) checks
them against the chain.

| `ctx` | |
|---|---|
| `method()`, `path()`, `host()`, `ip()` | the request |
| `header(k)`, `query(k)` | a header or query parameter, "" if absent |
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

```lua
---@stage http:request/sec_rule
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

The sandbox opens `base`, `table`, `string` and `math` only: there is no
`os`, `io`, `load` or `require`. Each call is limited to 50ms (`WithTimeout`),
and a script that fails or times out is skipped by the edge.

## Packages

| Package | |
|---|---|
| `scripting` | `Compile`, `Program` (`Has`, `Call`) and `Stages` |
| `scripting/exchange` | `Exchange`, the request the host hands a hook, and the log `Level` |
| `scripting/phase` | `Phase`: `Request` (`http:request`), `Response` (`http:response`) |
| `internal/hooks` | parses the `---@stage` lines into the hooks of a script; the stage table |
| `internal/api` | the `ctx` and `log` APIs; `Run` calls the hooks of a stage in a Lua state |
| `internal/sandbox` | the Lua state: the libraries a script may use, no more |

The host implements `exchange.Exchange` and calls
`Program.Call(ctx, phase.Request, stage, ex)` around each stage.

## Editor support

`meta/sentinez.lua` declares `ctx` (`sz.Context`) for lua-language-server
(LuaLS / EmmyLua): annotate a hook with `---@param ctx sz.Context` for completion
and type checks. `.luarc.json` loads it for this directory; elsewhere add the
directory to `workspace.library`. `TestMetaMatchesContext` keeps it in step with
the methods of `ctx`.
