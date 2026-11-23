# 12. Known issues

This list was compiled while reading the code at commit `47e5f876` and
reviewed again at `d64307d3` and `83b43464`. Each entry gives the location and the
consequence so it can be turned into an issue. Entries marked *(needs
verification)* are inferred from the code and have not been reproduced.
Remove an entry once it is fixed.

## 12.1 Logic bugs (affect behavior)

| # | Location | Problem |
|---|---|---|
| L1 ([#395](https://github.com/sentinez/sentinez/issues/395)) | [controlplane/pkg/schema](../staging/src/github.com/sentinez/controlplane/pkg/schema/) since `28dd2173` *(needs verification)* | The timestamp columns of the iam, tenant, security and activities tables were renamed (`created_at` → `metadata_created_at` / `owner_created_at`) and `security.secrules.action` went from `JSONB` to `TEXT`, but the schema sync never renames or retypes: on a database created before, the generated scanners select columns that do not exist. Drop the tables (dev) or migrate them by hand |

## 12.2 Security

| # | Location | Problem |
|---|---|---|
| S4 ([#388](https://github.com/sentinez/sentinez/issues/388)) | apiserver `Logging` middleware | Reads the whole body and debug-logs it on error — can write passwords from `/iam/login` to the logs. The body is also read with no size limit, before auth (memory exhaustion); `middleware/injection.go` does the same |
| S5 ([#389](https://github.com/sentinez/sentinez/issues/389)) | [cmd/szedge/v1/server/pprof.go](../cmd/szedge/v1/server/pprof.go) | pprof is always open on `defaults.PprofAddress` (`0.0.0.0:6060`; compose also publishes this port) |
| S6 ([#390](https://github.com/sentinez/sentinez/issues/390)) | [pkg/network/httpx/context.go](../pkg/network/httpx/context.go), realtime WebSocket | `CheckOrigin` always returns `true` |
| S7 ([#391](https://github.com/sentinez/sentinez/issues/391)) | `cmd/szedge/v1/Dockerfile` | The image bakes in the cert/key pair from the source tree |
| S8 ([#392](https://github.com/sentinez/sentinez/issues/392)) | `httpx.Context.RequestIP` | Trusts client-supplied `X-Forwarded-For`/`X-Real-IP` with no trusted-proxy list → IP spoofing to evade IP rules/rate limits (`Quic` and `Standard` modes, i.e. the default) |

## 12.3 Unfinished features

- [#393](https://github.com/sentinez/sentinez/issues/393) A standalone edge gets no control plane changes: they reach the
  edge only through `szcentraldata`, which the zone leader of a clustered
  edge follows (Consul, `SZ_SERVICE_TOKEN`, `SZ_SYNC_KEY`). Otherwise the
  edge reads `proxy.yaml` once at startup and never reloads it.
- [#394](https://github.com/sentinez/sentinez/issues/394) Settings that do not declare `metadata.tenantId` /
  `ingress.resource.id` share the ids `tnt_default` / `rsrc_default`, so
  their events are grouped together in the event tables; a setting with the
  default tenant id is not synced between the edges.
- [#395](https://github.com/sentinez/sentinez/issues/395) Old DBs keep the redundant single `resource_id` index: the
  composite index `(resource_id, <time>)` was added to every event table, but
  schema sync never drops indexes (same cause as L1: no migrations).
- [#396](https://github.com/sentinez/sentinez/issues/396) [deploy/docker/docker-compose.yaml](../deploy/docker/docker-compose.yaml)
  creates the event topics with 3 partitions and disables topic auto-create,
  so the producer's `DefaultTopicPartitions` (6) never applies in that
  setup.
- [#397](https://github.com/sentinez/sentinez/issues/397) `szcentraldata` sync: the edge loads the tables as
  `edgepb.Setting`s rebuilt from the rows, which keep only the type of a rule
  action (not its `params`) and nothing of the CDN: the CDN of the file
  stays. The CA certificates reach the edge (sealed) but nothing on the edge
  uses them yet. The edge takes a namespace from the setting of the file when
  the tables hold the same domain.
- [#398](https://github.com/sentinez/sentinez/issues/398) The zone leader writes the zone status
  (`sentinez/dmz/edge/zone/<zone>/status`) and follows centraldata; nothing
  reads the status.
- [#399](https://github.com/sentinez/sentinez/issues/399) Unloading a namespace (deleted resource, changed
  `ingress.resource.domain`) keeps its plain reverse proxies (`plainProxy`):
  they are keyed by upstream and shared between namespaces, with no
  reference count.
- [#400](https://github.com/sentinez/sentinez/issues/400) The waiting room (`ROM`) is a pass-through and `pkg/szqueue` is
  empty.
- [#401](https://github.com/sentinez/sentinez/issues/401) `pkg/sztopic`, `pkg/szstorage/{cache/keyval,dbx/clickhouse,
  search/elastic}` are empty packages; `SZ_TIMESCALE_URI`,
  `SZ_CLICKHOUSE_URI` are declared but unused.
- [#402](https://github.com/sentinez/sentinez/issues/402) `DataPlaneService` has no handler.
- [#403](https://github.com/sentinez/sentinez/issues/403) `wsz.WebSocket.Shutdown` is a no-op (`httpx/std.Shutdown` is a
  placeholder).
- [#404](https://github.com/sentinez/sentinez/issues/404) IAM uniqueness (one user per email, one account per provider per
  user, `LOCAL` usernames) is only checked in the service before inserting;
  the tables have no unique index (schema unique indexes are not used here),
  so two concurrent sign-ups can both pass.
- [#405](https://github.com/sentinez/sentinez/issues/405) IAM: a user who signed up with a passkey cannot add a password
  (there is no "link" RPC for `LOCAL`). The console has `PasskeyLink()` in
  `lib/api/iam/passkey.ts` but no page calls it.
- [#406](https://github.com/sentinez/sentinez/issues/406) IAM `UpdateUser` changes `User.email` without updating the email
  on the user's accounts, and fails unless some account already has that
  email.
- [#407](https://github.com/sentinez/sentinez/issues/407) Passkey sessions and pending registrations are kept in process
  memory: they are lost on restart and not shared between apiserver
  replicas.

## 12.4 Release and tooling

These were found while preparing 0.0.1 ([13-release.md](13-release.md)).
R1 blocks the release.

| # | Location | Problem |
|---|---|---|
| R1 ([#383](https://github.com/sentinez/sentinez/issues/383)) | `go.mod` of `core`, `pkg`, `controlplane`, `contrib/httphz`, `tools` | The modules require `sentinez/api v0.0.0`, the zero pseudo-version `v0.0.0-00010101000000-…` of `contrib/pkg` and `pkg`, or pseudo-versions of monorepo commits. Go ignores the `replace` of a dependency, so `go get` of these modules fails outside the repo. Before tagging, bump the requires to the release version (13.2) |
| R2 ([#384](https://github.com/sentinez/sentinez/issues/384)) | [core/core.go](../staging/src/github.com/sentinez/core/core.go) | `Version` is a `const` (`v0.0.1-beta`), so `-ldflags -X` cannot set it, despite its comment; no build passes `-X` either. Every binary shows `v0.0.1-beta` in its banner |
| R3 ([#385](https://github.com/sentinez/sentinez/issues/385)) | `api/proto/**/generate.sh` that call `protoc-go-inject-tag` | The inject step lengthens the tags without realigning the comments, so the generated `*.pb.go` fail `gofmt` (`make fmt.go.check`) after each regeneration. CI runs neither `fmt.go.check` nor a build, so this goes unnoticed |
| R4 ([#386](https://github.com/sentinez/sentinez/issues/386)) | `cmd/sz*/Makefile`, `sz.<svc>.image.build` | Images are only tagged `$(TAG):latest`; a versioned tag has to be added by hand with `docker tag` |
| R5 ([#387](https://github.com/sentinez/sentinez/issues/387)) | [hack/gomodules.sh](../hack/gomodules.sh) | `bpf` and `integration` are published but not listed, so lint, `go mod tidy`, tests, CI and `make release.check` all skip them |
