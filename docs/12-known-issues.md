# 12. Known issues

This list was compiled while reading the code at commit `47e5f876`. Each
entry gives the location and the consequence so it can be turned into an
issue. Entries marked *(needs verification)* are inferred from the code and
have not been reproduced. Remove an entry once it is fixed.

## 12.1 Logic bugs (affect behavior)

| # | Location | Problem | Consequence |
|---|---|---|---|
| 1 | [internal/memory/memory.go](../internal/memory/memory.go) `LoadRateLimiter` + [core/common/edge_normalize.go](../staging/src/github.com/sentinez/core/common/edge_normalize.go) | Limiters are only loaded when `IngressRuntime.Status == ACTIVE`, but `NormalizeEdgeSetting` only fills `IngressRuntime` for `secRules`, not `limiters`. Also, an inactive limiter triggers `return nil` (skipping all later limiters) | Rate limits configured in `proxy.yaml` **never take effect** |
| 2 | [core/limiter/ratelimiter.go](../staging/src/github.com/sentinez/core/limiter/ratelimiter.go) | `lastBlocking` is a field of `RateLimiter`, shared by every key | One IP exceeding the limit blocks **every IP** in the namespace for `timeout` |
| 3 | [controlplane/iam/v1/service/service.go](../staging/src/github.com/sentinez/controlplane/iam/v1/service/service.go) `UsernameOrEmailMustUnique`, `CreateUser`, `UpdateUser` | Uses `if !errors.Is(err, ErrNotFound) { return err }` — when a record **is** found, `err == nil`, so the function returns early with `nil` | `CreateAccount` allows duplicate usernames/emails; `CreateUser` returns `(nil, nil)` when the email exists; `UpdateUser` never updates when the account exists |
| 4 | [controlplane/security/v1/repos/secrule/secrule.go](../staging/src/github.com/sentinez/controlplane/security/v1/repos/secrule/secrule.go) `Update` | Only updates `name`, `description`, `expr`, `status` | Changing `priority`/`action` through the API is not persisted |
| 5 | [internal/cluster/cluster.go](../internal/cluster/cluster.go) `newConfig` | When both membership and discovery addresses are valid, the function `return conf` before setting `conf.Started` | `dictReady` is never signaled → DMaps are never created; after 1024 `Put`s the caller goroutine blocks |
| 6 | [cmd/szedge/v1/main.go](../cmd/szedge/v1/main.go) + [core/runner/hook.go](../staging/src/github.com/sentinez/core/runner/hook.go) | `engine.Hertz` sets `OnConnect` in one OnStart hook while `server.Start` runs in another; each hook runs in its own goroutine | Race: `Start` may read `options` before `SetOptions` runs → `opt.OnConnect` is nil and panics on the first connection *(needs verification)* |
| 7 | [core/runner/hook.go](../staging/src/github.com/sentinez/core/runner/hook.go) `OnStart` | Errors returned by `start` are only logged when they are `http.ErrServerClosed`; any other error is swallowed | A service that fails to listen (port in use, bad cert, ...) keeps "running" with no log |
| 8 | [core/runner/runner.go](../staging/src/github.com/sentinez/core/runner/runner.go) `_OTLP` | `Insecure: secure` — the value is inverted | An `https://` endpoint is used without TLS, while a plain endpoint uses TLS |
| 9 | [contrib/httphz/proxy/ws_reverse_proxy.go](../staging/src/github.com/sentinez/contrib/httphz/proxy/ws_reverse_proxy.go) | `p.target += string(uri)` mutates a field of a shared proxy | The WebSocket target grows with every request, and there is a data race |
| 10 | [contrib/httphz/server.go](../staging/src/github.com/sentinez/contrib/httphz/server.go) `Handle` | The `fn = s.chains[i](fn)` loop runs **inside** the handler and reassigns a captured variable | Each request wraps the middleware once more (not visible yet because the edge never calls `Use`) |
| 11 | [internal/memory/routes/routes.go](../internal/memory/routes/routes.go) | Always uses `proxyPass[0]`; no empty check; ignores `balanceStrategy`; path join is not normalized | A location without upstreams panics; no load balancing; may produce `//path` |
| 12 | [internal/memory/memory.go](../internal/memory/memory.go) `LoadRulesets` | Ignores the contents of `rulesets[]`; always CRS v4.16.0 + the RCE flag; the status check is commented out | The rule version/groups cannot be configured; SQLi (942) is never enabled |
| 13 | [internal/dmz/edge/http/logging/logging.go](../internal/dmz/edge/http/logging/logging.go) | Calls `bpf.LookupBandwidth` on every request, but the eBPF objects are only loaded in the dataplane process | Always an error/0, costs time, and logs at `Infof` on every request |
| 14 | [internal/bpf/security.go](../internal/bpf/security.go) `BlockCIDR` | Stores `ip` with `binary.BigEndian.Uint32` while XDP looks up with the raw `saddr` (network byte order in memory) | On little-endian hosts the keys don't match, so blocked CIDRs have no effect *(needs verification)*. The function has no callers yet |
| 15 | [shared/eventq/eventq.go](../staging/src/github.com/sentinez/shared/eventq/eventq.go) `Close`, `CloseAll` | Type-asserts to `chan string` / `chan any` instead of `chan T` | Panics for any other `T` (no callers yet) |

## 12.2 Security

| # | Location | Problem |
|---|---|---|
| S1 | [pkg/apps/gateway/apiserver/middleware/middleware.go](../pkg/apps/gateway/apiserver/middleware/middleware.go) `AllowCORS` | Reflects any `Origin` together with `Allow-Credentials: true` (the code itself says not to use this in production) |
| S2 | Control-plane handlers | Authorization must be called in each handler; only `ListAccounts` and `ListActivities` do it. Tenant, Security, `ListUsers`, `DeleteUser`, … require no token |
| S3 | IAM `loginAdmin` | Compares the admin password in plaintext, not in constant time; no limit on failed admin logins |
| S4 | apiserver `Logging` middleware | Reads the whole body and debug-logs it on error — can write passwords from `/iam/login` to the logs |
| S5 | [cmd/szedge/v1/main.go](../cmd/szedge/v1/main.go) `init` | pprof is always open on `:6060` (compose also publishes this port) |
| S6 | [pkg/network/httpx/std/context.go](../pkg/network/httpx/std/context.go), realtime WebSocket | `CheckOrigin` always returns `true` |
| S7 | `cmd/szedge/v1/Dockerfile` | The image bakes in the cert/key pair from the source tree |
| S8 | `stdhttpx.Context.RequestIP` | Trusts client-supplied `X-Forwarded-For`/`X-Real-IP` with no trusted-proxy list → IP spoofing to evade IP rules/rate limits (`Standard` mode) |

## 12.3 Unfinished features

- No sync channel from control plane → edge: SecRules/tenants in the DB do
  not affect the edge; the edge only reads `proxy.yaml` at startup and does
  not reload.
- `szcentraldata`: empty `main()`. `AnalyticService` is not mounted in the
  apiserver. Consul discovery (`WithDiscorvery`, `VisitToEndpoint`) exists
  but is not enabled.
- The `EdgeService` gRPC server is not started; `EvaluateRuleset` evaluates
  an empty `Expression` (always matches) and ignores `ruleset_id`.
- SecRule actions other than `BLOCK` are not executed; operators
  `GT/GTE/LT/LTE` are not implemented.
- Waiting room (`ROM`), `pkg/queue`, `shared/topic`, `core/storage/{keyval,
  clickhouse,kafka,elastic}`, `eventpub/eventsub` are empty packages.
- Only one CDN rule and one limiter per namespace.
- `DataPlaneService` has no handler; `wsz.WebSocket.Shutdown` and
  `stdhttpx.Shutdown` are no-ops.
- `SENZ_TIMESCALE_URI`, `SENZ_CLICKHOUSE_URI` are declared but unused.

## 12.4 Minor issues / code hygiene

- ID prefixes don't match between code and proto constraints: the repo
  generates `senz.security.sec_rules.<uuid>` but the proto requires
  `senz.security.secrules.`; users get `senz.iam.users.` but
  `GetUserRequest/UpdateUserRequest/DeleteUserRequest` require `senz.users.`.
  Not visible today because the gRPC server doesn't validate requests.
- `corehttp.Forbidden/NotFound/...` render under one global `sync.Mutex`,
  serializing every error response.
- `preflightHandler` sets `Access-Control-Max-Age` after `WriteHeader` (no
  effect); the grpc-gateway `extendHeader` sets the `Server` header after
  the handler has already written the response.
- `GenerateRulesets` writes `WAF.conf.lock` to the working directory every
  time a WAF is built.
- Stale default paths: greeter `--env_file` (`./cmd/mesh/greeter/v1/.env`)
  and realtime (`./cmd/realtime/.env`); `hack/mkcert-local.sh` moves certs
  into `./cmd/edge/v1/`.
- The tenant repo's `scan` ignores `rows.Err()`; `UpdateResource` does not
  update `resource_setting`.
- `LoadRateLimiter` calls `zlog.Fatalf` on a bad duration (kills the whole
  edge).
- CI only lints/tests the root module; staging modules are only linted
  locally via `make lint`.
