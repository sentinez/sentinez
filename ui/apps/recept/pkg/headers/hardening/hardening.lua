-- Adds the security headers to every response, and logs the requests
-- that still come over plain HTTP.
--
-- @version v1.0.0
-- @keywords headers, hsts, csp

---Sets HSTS, nosniff, the frame and referrer policies on the response.
---@stage http.request/router
---@param ctx sz.Context
function security_headers(ctx)
  ctx:set_response_header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
  ctx:set_response_header("X-Content-Type-Options", "nosniff")
  ctx:set_response_header("X-Frame-Options", "DENY")
  ctx:set_response_header("Referrer-Policy", "strict-origin-when-cross-origin")
end

---Logs a request without TLS, with its path and client.
---@stage http.request/secrule
---@param ctx sz.Context
function audit_plain_http(ctx)
  if not ctx:tls() then
    log.info("plain http", ctx:path(), ctx:ip())
  end
end
