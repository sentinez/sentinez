---@meta sentinez

-- Type annotations of the Sentinéz edge scripts, for lua-language-server
-- (LuaLS / EmmyLua). Nothing here runs: the edge provides ctx at runtime.
-- Point the editor at this directory (workspace.library in .luarc.json)
-- and annotate each hook with ---@param ctx sz.Context.
--
-- A hook is a global function, of any name, preceded by one
-- ---@stage <phase>/<stage> line per stage and phase it runs at:
--
--   ---@stage http.request/secrule
--   ---@stage http.request/secruleset
--   ---@param ctx sz.Context
--   function deny_debug(ctx) end
--
-- The stages are sz.Stage. http.request: the hook may answer the request with
-- ctx:respond, which skips the stage. http.response: it runs once the response is
-- written, it only observes. The hooks of a stage run in source order.

---The stages of the edge chain a hook may annotate, in chain order. The
---ones before the namespace of the request is known (tracing,
---logging, domain) have no hooks.
---@alias sz.Stage
---| "caching" # CDN cache
---| "secratelimit" # rate limiters
---| "waitingroom" # waiting room
---| "staticassets" # static assets
---| "secrule" # secrules
---| "secruleset" # OWASP CRS
---| "router" # to the upstream

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

---Answers the request, ending it: the stage is skipped. In an http.response
---hook it raises an error, the response being written already.
---@param status integer 100 to 599
---@param body? string
function Context:respond(status, body) end

-- Rule scripts. A secrule or a rate limiter may be a Lua source in place
-- of its expr: it defines one global function the edge calls for each
-- request. ctx is read-only there: set_header, set_response_header and
-- respond raise an error, the request is changed through the action. A
-- rule that fails or runs over the time limit is skipped. A rule script has
-- no ---@stage line.

---The name of an action, as its SZ_ constant holds it.
---@alias sz.Action
---| "block"
---| "challenge"
---| "log"
---| "modify_header"
---| "redirect"
---| "route_to"
---| "set_tag"

-- The actions a secrule script returns, as an eBPF program returns
-- XDP_DROP: return SZ_BLOCK, or {type = SZ_REDIRECT, params = {...}}. The
-- name ("block") is the same.

---@type sz.Action
SZ_BLOCK = "block"
---@type sz.Action
SZ_CHALLENGE = "challenge"
---@type sz.Action
SZ_LOG = "log"
---@type sz.Action
SZ_MODIFY_HEADER = "modify_header"
---@type sz.Action
SZ_REDIRECT = "redirect"
---@type sz.Action
SZ_ROUTE_TO = "route_to"
---@type sz.Action
SZ_SET_TAG = "set_tag"

---An action, or a table with the action and its params, as the action of
---a rule takes them.
---@alias sz.RuleAction sz.Action | { type: sz.Action, params?: table }

---The function of a secrule script: nil or false when the request does not
---match, else the action to take.
---@alias sz.Match fun(ctx: sz.Context): sz.RuleAction|nil|false

---The function of a rate limiter script: the key the request is counted
---under (256 bytes at most), true for the client IP, nil or false not to
---count it.
---@alias sz.Key fun(ctx: sz.Context): string|boolean|nil
