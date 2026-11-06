# 12. Known issues

This list was compiled while reading the code at commit `47e5f876` and
reviewed again at `d64307d3`. Each entry gives the location and the
consequence so it can be turned into an issue. Entries marked *(needs
verification)* are inferred from the code and have not been reproduced.
Remove an entry once it is fixed.

## 12.1 Logic bugs (affect behavior)

None known.

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
- The edge does not fill `resource_id`/`tenant_id` in its rule events, so
  the collector event tables have them empty and `eventfilter.Filter`'s
  `ResourceIds` matches nothing yet. The event repositories' `Get`/`List`
  have no caller (no analytic read path yet).
- [deploy/docker/telemetry/docker-compose.yaml](../deploy/docker/telemetry/docker-compose.yaml)
  creates the event topics with 3 partitions and disables topic auto-create,
  so the producer's `DefaultTopicPartitions` (6) never applies in that
  setup.
- `szcentraldata`: empty `main()`. Consul discovery (`WithDiscorvery`,
  `VisitToEndpoint`) exists but is not enabled.
- The zone leader only writes the zone status
  (`sentinez/dmz/edge/zone/<zone>/status`); `Elector.Publish` and
  `leader.All` are ready for more zone-wide records, but none exists yet.
  Nothing reads the status either.
- Unloading a namespace (deleted resource, renamed `server.name`) keeps
  its reverse proxies: they are keyed by upstream and shared between
  namespaces, with no reference count.
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
