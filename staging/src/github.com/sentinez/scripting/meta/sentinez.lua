---@meta sentinez

-- Type annotations of the Sentinéz edge scripts, for lua-language-server
-- (LuaLS / EmmyLua). Nothing here runs: the edge provides ctx at runtime.
-- Point the editor at this directory (workspace.library in .luarc.json)
-- and annotate each hook with ---@param ctx sz.Context.
--
-- A hook is a global function, of any name, preceded by one
-- ---@stage <phase>/<stage> line per stage and phase it runs at:
--
--   ---@stage http:request/sec_rule
--   ---@stage http:request/sec_ruleset
--   ---@param ctx sz.Context
--   function deny_debug(ctx) end
--
-- The stages are sz.Stage. http:request: the hook may answer the request with
-- ctx:respond, which skips the stage. http:response: it runs once the response is
-- written, it only observes. The hooks of a stage run in source order.

---The stages of the edge chain a hook may annotate, in chain order. The
---ones before the namespace of the request is known (TRC, LOG, DMA) have
---no hooks.
---@alias sz.Stage
---| "cache" # CDN: cdn.Cache
---| "sec_limiter" # LMT: secure.SecLimiter
---| "waiting_room" # ROM: room.WaitingRoom
---| "static" # STC: static.Static
---| "sec_rule" # RUL: secure.SecRule
---| "sec_ruleset" # RST: secure.SecRuleset (OWASP CRS)
---| "router" # ROU: routing.StandardRouter, to the upstream

---Writes lines to the logs of the request, with the request id and the
---name of the script: the arguments, as tostring prints them, joined by
---tabs like print. log.info(...) and log:info(...) are the same. A line is
---cut at 1024 bytes, and a hook call logs 16 lines at most. In the script
---body, outside of a hook, the line is dropped.
---@class sz.Log
log = {}

---A line at the debug level: only written when the edge logs debug.
---@param ... any
function log.debug(...) end

---A line at the info level.
---@param ... any
function log.info(...) end

---A hook: called with the request, its result is ignored.
---@alias sz.Hook fun(ctx: sz.Context)

---The request a hook reads and may change.
---@class sz.Context
local Context = {}

---The request method, e.g. "GET".
---@return string
function Context:method() end

---The request path, without the query.
---@return string
function Context:path() end

---The request host.
---@return string
function Context:host() end

---The client IP address.
---@return string
function Context:ip() end

---A request header, "" when absent.
---@param key string
---@return string
function Context:header(key) end

---A query parameter, "" when absent.
---@param key string
---@return string
function Context:query(key) end

---Sets a header of the request sent upstream. An invalid name is ignored.
---@param key string
---@param value string
function Context:set_header(key, value) end

---Sets a header of the response. An invalid name is ignored.
---@param key string
---@param value string
function Context:set_response_header(key, value) end

---Answers the request, ending it: the stage is skipped. In an http:response
---hook it raises an error, the response being written already.
---@param status integer 100 to 599
---@param body? string
function Context:respond(status, body) end
