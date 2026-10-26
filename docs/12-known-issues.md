# 12. Known issues

This list was compiled while reading the code at commit `47e5f876` and
reviewed again at `d64307d3`. Each entry gives the location and the
consequence so it can be turned into an issue. Entries marked *(needs
verification)* are inferred from the code and have not been reproduced.
Remove an entry once it is fixed.

## 12.1 Logic bugs (affect behavior)

| # | Location | Problem | Consequence |
|---|---|---|---|
| 2 | [controlplane/apps/security/v1/repos/secrule/secrule.go](../staging/src/github.com/sentinez/controlplane/apps/security/v1/repos/secrule/secrule.go) `Update` | Only updates `name`, `description`, `expr`, `status` | Changing `priority`/`action` through the API is not persisted |
| 3 | [pkg/cluster/cluster.go](../pkg/cluster/cluster.go) `newConfig` | When both membership and discovery addresses are valid, the function `return conf` before setting `conf.Started` | `dictReady` is never signaled → DMaps are never created; after 1024 `Put`s the caller goroutine blocks |
| 7 | [contrib/httphz/proxy/ws_reverse_proxy.go](../staging/src/github.com/sentinez/contrib/httphz/proxy/ws_reverse_proxy.go) | `p.target += string(uri)` mutates a field of a shared proxy | The WebSocket target grows with every request, and there is a data race |
| 8 | [contrib/httphz/server.go](../staging/src/github.com/sentinez/contrib/httphz/server.go) `Handle` | The `fn = s.chains[i](fn)` loop runs **inside** the handler and reassigns a captured variable | Each request wraps the middleware once more (not visible yet because the edge never calls `Use`) |
| 9 | [internal/memory/routes/routes.go](../internal/memory/routes/routes.go) | Always uses `proxyPass[0]`; no empty check; ignores `balanceStrategy` | A location without upstreams panics; no load balancing |
| 11 | [internal/handlers/logging/logging.go](../internal/handlers/logging/logging.go) | Calls `bpf.LookupBandwidth` on every request, but the eBPF objects are only loaded in the dataplane process | Always an error/0, costs time, and logs at `Infof` on every request |
| 12 | [internal/dmz/dataplane/bpf/security.go](../internal/dmz/dataplane/bpf/security.go) `BlockCIDR` | Stores `ip` with `binary.BigEndian.Uint32` while XDP looks up with the raw `saddr` (network byte order in memory) | On little-endian hosts the keys don't match, so blocked CIDRs have no effect *(needs verification)*. The function has no callers yet |

## 12.2 Security

| # | Location | Problem |
|---|---|---|
| S1 | [pkg/apps/gateway/apiserver/middleware/middleware.go](../pkg/apps/gateway/apiserver/middleware/middleware.go) `AllowCORS` | Reflects any `Origin` together with `Allow-Credentials: true` (the code itself says not to use this in production) |
| S2 | Control-plane handlers | Authorization must be called in each handler; only `ListAccounts` and `ListActivities` do it. Tenant, Security, `ListUsers`, `DeleteUser`, … require no token |
| S3 | IAM `loginAdmin` | Compares the admin password in plaintext, not in constant time; no limit on failed admin logins |
| S4 | apiserver `Logging` middleware | Reads the whole body and debug-logs it on error — can write passwords from `/iam/login` to the logs |
| S5 | [cmd/szedge/v1/server/pprof.go](../cmd/szedge/v1/server/pprof.go) | pprof is always open on `defaults.PprofAddress` (`0.0.0.0:6060`; compose also publishes this port) |
| S6 | [pkg/network/httpx/context.go](../pkg/network/httpx/context.go), realtime WebSocket | `CheckOrigin` always returns `true` |
| S7 | `cmd/szedge/v1/Dockerfile` | The image bakes in the cert/key pair from the source tree |
| S8 | `httpx.Context.RequestIP` | Trusts client-supplied `X-Forwarded-For`/`X-Real-IP` with no trusted-proxy list → IP spoofing to evade IP rules/rate limits (`Quic` and `Standard` modes, i.e. the default) |

## 12.3 Unfinished features

- No sync channel from control plane → edge: SecRules/tenants in the DB do
  not affect the edge; the edge only reads `proxy.yaml` at startup and does
  not reload.
- `szcentraldata`: empty `main()`. `AnalyticService` is not mounted in the
  apiserver. Consul discovery (`WithDiscorvery`, `VisitToEndpoint`) exists
  but is not enabled.
- `EdgeService.EvaluateRuleset` evaluates
  an empty `Expression` (always matches) and ignores `ruleset_id`.
- Waiting room (`ROM`), `shared/szqueue`, `shared/sztopic`, `shared/szstorage/{keyval,
  clickhouse,elastic}`, `eventpub/eventsub` are empty packages.
- `DataPlaneService` has no handler; `wsz.WebSocket.Shutdown` is a no-op.
- IAM uniqueness (one user per email, one account per provider per user,
  `LOCAL` usernames) is only checked in the service before inserting; the
  tables have no unique index (`dbx.Index` is single-column, non-unique), so
  two concurrent sign-ups can both pass.
- IAM: a user who signed up with a passkey cannot add a password (there is
  no "link" RPC for `LOCAL`). The console has `PasskeyLink()` in
  `lib/api/iam/passkey.ts` but no page calls it.
- IAM `UpdateUser` changes `User.email` without updating the email on the
  user's accounts, and fails unless some account already has that email.
- Passkey sessions and pending registrations are kept in process memory:
  they are lost on restart and not shared between apiserver replicas.
- `SENZ_TIMESCALE_URI`, `SENZ_CLICKHOUSE_URI` are declared but unused.
