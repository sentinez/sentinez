# github.com/sentinez/scripting

> ⚠️ **This is an automatically published [staged repository](https://github.com/sentinez/scripting) for Sentinez**.   
> Contributions, including issues and pull requests, should be made to the main Sentinez repository: [https://github.com/sentinez/sentinez](https://github.com/sentinez/sentinez).  
> This repository is read-only for importing, and not used for direct contributions.

=======================================================================

This module runs the Lua scripts of the edge in a sandbox.

`szlua` compiles a script (`Compile`) and calls its hooks (`Program.Call`).
The annotations of a global function hook it, whatever its name, to
stages of the edge handler chain: `@module` names the stages, `@types` when
it runs.

```lua
---@module RUL
---@module RST
---@types before
---@param ctx sz.Context
function deny_debug(ctx)
  if ctx:header("X-Debug") == "1" then
    ctx:respond(403, "no")      -- ends the request, before only
  end
end

---@module ROU
---@types before after
---@param ctx sz.Context
function tag(ctx)
  ctx:set_header("X-Edge", "sentinez")
end
```

- `---@module <stage>...`: the stages the function hooks, on one line or
  several, quoted or not (`---@module 'RST'`).
- `---@types before`: runs before the stage; `ctx:respond` ends the
  request there and the stage is skipped.
- `---@types after`: runs once the stage, and the ones after it, returned.
  The response is written: it only observes.
- The annotations must precede a global function (`function name(ctx)`),
  other comments aside, or the script does not compile; so do a stage
  without `@types` and `@types` without a stage. A `@module` naming no
  stage, without `@types`, is the LuaLS one and is left alone.
  A function without annotations is not a hook, whatever its name.
- The hooks of a stage run in source order, until one responds.

The stages are the ids of the plugins of the edge chain
(`internal/handlers/handlers.go`), in chain order:

| Stage | Plugin |
|---|---|
| `CDN` | CDN cache |
| `LMT` | rate limiter |
| `ROM` | waiting room |
| `STC` | static files |
| `RUL` | secrules |
| `RST` | rulesets (OWASP CRS) |
| `ROU` | router, to the upstream |

The stages before the namespace of the request is known (`TRC`, `LOG`,
`DMA`) have no hooks. `szlua.Stages` lists them; a test of the edge
(`TestScriptStagesAreTheChain`) checks it against the chain.

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
---@module RUL
---@types before
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

## Editor support

`szlua/meta/sentinez.lua` declares `ctx` (`sz.Context`) for lua-language-server
(LuaLS / EmmyLua): annotate a hook with `---@param ctx sz.Context` for completion
and type checks. `.luarc.json` loads it for this directory; elsewhere add the
directory to `workspace.library`. `TestMetaMatchesContext` keeps it in step with
the methods of `ctx`.
