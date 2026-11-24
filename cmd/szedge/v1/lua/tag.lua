---@stage http.request/router
---@param ctx sz.Context
function TagEdge(ctx)
    log.debug("tag", ctx:host(), ctx:path())
    ctx:set_header("X-Edge", "sentinez")
end
