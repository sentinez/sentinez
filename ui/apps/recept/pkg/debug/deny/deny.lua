-- Blocks the debug requests and the probes of internal paths.
--
-- A request with `X-Debug: 1` or under `/_internal/` never reaches the
-- upstream. `/old` is sent to `/new` on the way.
--
-- @version v0.3.0
-- @keywords debug, internal

---Blocks the debug header and the internal paths, redirects /old.
---@param ctx sz.Context
function match(ctx)
  if ctx:header("X-Debug") == "1" then
    return SZ_BLOCK
  end
  if ctx:path():sub(1, 11) == "/_internal/" and not ctx:ip_in("10.0.0.0/8") then
    return SZ_BLOCK
  end
  if ctx:path() == "/old" then
    return {type = SZ_REDIRECT, params = {location = "/new"}}
  end
end
