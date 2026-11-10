# 8. API & Protobuf

The `github.com/sentinez/sentinez/api` module ([api/](../api/)) is the
source of truth for every data type: service config, edge config, rules,
log events and the REST/gRPC API.

## 8.1 Proto layout

```
api/proto/sentinez/
├── apps/<domain>/v1/        # domain services (REST via grpc-gateway)
│   ├── <domain>.proto       #   service + request/response
│   └── model.proto          #   DB model (x_message.database_model)
├── dmz/
│   ├── edge/v1/             # EdgeService, Setting (proxy.yaml), Context
│   └── dataplane/v1/        # DataPlaneService
├── gateway/
│   ├── apiserver/v1/        # x_meta only
│   └── realtime/v1/         # x_meta only
└── types/                   # shared types
    ├── v1/                  # known.proto (enums), model.proto (Pages, Metadata), options.proto
    ├── setting/v1/          # Config, Flag, Senz enum (environment variables)
    ├── rule/v1/             # Expression, Condition, Operator, Action
    ├── secrule/v1/          # SecRule runtime/Lite, WAF Event
    ├── cdn/v1/              # CDN rule
    ├── coreruleset/v1/      # CoreRulesets (rule parser output)
    ├── net/v1/              # Transport (conn_id, SNI)
    └── net/http/v1/         # Request, RequestEvent, Event
```

Go packages: `github.com/sentinez/sentinez/api/proto/sentinez/<path>;<alias>pb`
(e.g. `settingpb`, `rulepb`, `edgepb`, `iampb`).

`buf lint` rules ([api/buf.yaml](../api/buf.yaml)): `STANDARD`, enum zero
values suffixed `_UNSPECIFIED`, services suffixed `Service`, request and
response types must differ.

## 8.2 Custom options (`types/v1/options.proto`)

| Extension | On | Fields | Purpose |
|---|---|---|---|
| `x_meta` (52001) | File | `service_name`, `service_zone`, `service_key` | Service identity: console banner, logger name, Consul key, `GetMeta<Name>()` |
| `x_message` (51001) | Message | `database_model`, `export_field`, `scanner` | Generates column-name constants `<Msg>_<Field>` / `X<Msg>_<Field>`; `scanner` also generates the row scanners below |
| `x_field` (53001) | Field | `flatten`, `ignore`, `codec`, `nullable` | How a field of a database model maps to columns, see below |
| `x_method` (50001) | RPC | `ignore`, `control_planes[]` | Generates `Get<Service><Method>() *XMethod` for authorization |

`Zone`: `DEMILITARIZED` (edge, dataplane), `INTERNAL` (domain services),
`PRIVATE_API` (apiserver, realtime), `PUBLIC_API`.

These options are processed by the `protoc-gen-go-senz` plugin
([tools/internal/senz/senz.go](../staging/src/github.com/sentinez/tools/internal/senz/senz.go)),
which emits `*_senz.pb.go` files.

### Generated row scanners

For a message with `x_message.scanner = true`
([tools/internal/senz/scanner.go](../staging/src/github.com/sentinez/tools/internal/senz/scanner.go)),
every nested message field is flattened into `<field>_<nested>` columns
(e.g. `metadata_created_at`, `owner_tenant_id`), and the plugin generates:

| Function | Does |
|---|---|
| `<Msg>Columns() []string` | The stored columns, in scan order |
| `<Msg>Values(m) ([]any, error)` | The values of `m` in the same order, for an insert |
| `ScanOne<Msg>(row) (*<Msg>, error)` | Scans one row (any `Scan(dest ...any) error`: `pgx.Row`, `dbx.Row`) |
| `Scan<Msg>(rows) ([]*<Msg>, error)` | Scans every row (`Next`/`Scan`/`Err`: `pgx.Rows`, `dbx.Rows`) |

How a field is scanned, set with `x_field`:

| Field | Column |
|---|---|
| scalar, enum | scanned straight into the field |
| `google.protobuf.Timestamp` / `Duration` | `time.Time` / `time.Duration`, converted |
| other message, no codec | flattened into its fields' columns |
| `codec = COLUMN_CODEC_PROTO` | one column of binary protobuf (`proto.Unmarshal`), e.g. `Resource.server` |
| `codec = COLUMN_CODEC_PROTOJSON` | one column of protobuf JSON (`protojson.Unmarshal`, unknown fields discarded), e.g. `CoreRuleset.content` |
| `codec = COLUMN_CODEC_DRIVER` | one column the driver scans into the field (e.g. JSONB), e.g. `SecRule.expr`, `SecRuleEvent.action` |
| `ignore = true` | not stored, left out (e.g. `SecRule.action_value`, `expr_preview`) |
| `nullable = true` | a singular scalar whose column may be `NULL`, scanned as the zero value (e.g. `Owner.tenant_id`, `Tenant.name`) |

`flatten` (older) only splits the field constants of a nested message;
with `scanner` every nested message is flattened anyway. The plugin fails
on a oneof member, on `nullable` set on a non-scalar, and on a `codec` set
on a non-message field. Tests:
[tools/example/example_scanner_test.go](../staging/src/github.com/sentinez/tools/example/example_scanner_test.go).

## 8.3 Code generation

Each domain directory has a `generate.sh` that runs `protoc` with include
paths `api/proto` and `api/third_party/{googleapis,grpc-gateway,protovalidate}`:

| Plugin | Output |
|---|---|
| `protoc-gen-go` | `*.pb.go` |
| `protoc-gen-go-grpc` (`require_unimplemented_servers=false`) | `*_grpc.pb.go` |
| `protoc-gen-grpc-gateway` | `*.pb.gw.go` |
| `protoc-gen-validate` | `*.pb.validate.go` |
| `protoc-gen-go-senz` | `*_senz.pb.go` |
| `protoc-gen-go-vtproto` (`features=marshal+unmarshal+size`) | `*_vtproto.pb.go` |
| `protoc-gen-openapiv2` | `api/docs/v1/<domain>.swagger.json` |

The `pool` vtproto feature is on for the pooled events only, each named by
a `--go-vtproto_opt=pool=<import path>.<Message>` in its `generate.sh`:
`secrule.v1.SecRuleEvent`, `coreruleset.v1.CoreRuleEvent`,
`net.http.v1.IngressEvent` and `rule.v1.MatchedRules`. It generates `ResetVT`
(`Reset` that keeps the capacity of the repeated fields), and
`<Msg>FromVTPool`/`ReturnToVTPool` on a pool of the message.

TypeScript for the UI: `cd api && buf generate` with `ts-proto`
([api/buf.gen.yaml](../api/buf.gen.yaml)) → `ui/packages/proto/ts-proto`.

Some fields carry `// @gotags: yaml:"..."` comments so that
`protoc-go-inject-tag` adds YAML struct tags.

[api/docs.go](../api/docs.go) embeds `docs/v1` (OpenAPI) and `docs/swagger`
(static Swagger UI) for the apiserver to serve.

## 8.4 REST endpoint table

Base URL: the apiserver address (`:8080` in `.env.example`). Auth: header
`Authorization: Bearer <JWT>`.

| Method | Path | RPC | Declared permission |
|---|---|---|---|
| GET | `/greeter/say` | `GreeterService.SayHello` | ignore |
| GET | `/greeter/status` | `GreeterService.Status` | ignore |
| PUT | `/iam/login` | `IdentityAccessManagementService.Login` | ignore |
| POST | `/iam/account` | `...CreateAccount` | ignore |
| GET | `/iam/accounts` | `...ListAccounts` | ignore (handler still requires a token) |
| GET | `/iam/users` | `...ListUsers` | PORTAL, ADMIN |
| GET | `/iam/user` | `...GetUser` | ignore |
| DELETE | `/iam/user/{id}` | `...DeleteUser` | ignore |
| GET | `/iam/status` | `...Status` | ignore |
| GET | `/iam/passkey/register/challenge` | `...PasskeyRegisterChallenge` | ignore |
| GET | `/iam/passkey/link/challenge` | `...PasskeyLinkChallenge` | ignore (service requires a token) |
| POST | `/iam/passkey/register/verify` | `...PasskeyRegisterVerify` | ignore |
| GET | `/iam/passkey/login/challenge` | `...PasskeyLoginChallenge` | ignore |
| PUT | `/iam/passkey/login/verify` | `...PasskeyLoginVerify` | ignore |
| GET | `/tenant/status` | `TenantService.Status` | ignore |
| GET | `/tenant/resources` | `TenantService.ListResource` | PORTAL, ADMIN |
| POST | `/tenant/resource` | `TenantService.CreateResource` | PORTAL, ADMIN |
| PUT | `/tenant/resource` | `TenantService.UpdateResource` | PORTAL, ADMIN |
| DELETE | `/tenant/resource` | `TenantService.DeleteResource` | PORTAL, ADMIN |
| GET | `/tenant/resource` | `TenantService.GetResource` | PORTAL, ADMIN |
| GET | `/tenant/resource/{resource_domain}` | `TenantService.GetResourceByDomain` | PORTAL, ADMIN |
| POST | `/security/secrule` | `SecurityService.CreateSecRule` | — |
| GET | `/security/secrule/{id}` | `SecurityService.GetSecRule` | — |
| PUT | `/security/secrule/{id}` | `SecurityService.UpdateSecRule` | — |
| DELETE | `/security/secrule/{id}` | `SecurityService.DeleteSecRule` | — |
| GET | `/security/secrules` | `SecurityService.ListSecRules` | — |
| GET | `/security/status` | `SecurityService.Status` | ignore |
| GET | `/analytic/activities` | `AnalyticService.ListActivities` | ignore |
| GET | `/analytic/events/http-ingress` | `AnalyticService.ListHttpIngressEvents` | ignore |
| GET | `/analytic/events/http-egress` | `AnalyticService.ListHttpEgressEvents` | ignore |
| GET | `/analytic/events/secrule` | `AnalyticService.ListSecRuleEvents` | ignore |
| GET | `/analytic/events/ratelimit` | `AnalyticService.ListRateLimitEvents` | ignore |
| GET | `/analytic/events/coreruleset` | `AnalyticService.ListCoreRuleEvents` | ignore |
| GET | `/analytic/events/histogram` | `AnalyticService.GetEventHistogram` | ignore |
| GET | `/analytic/events/top-values` | `AnalyticService.GetEventTopValues` | ignore |
| GET | `/analytic/events/http-egress/latency` | `AnalyticService.GetEgressLatency` | ignore |

"Declared permission" is the option in the proto; enforcement depends on the
handler (see
[07-controlplane.md](07-controlplane.md#72-authentication--authorization)).

`GetEgressLatency` queries latency aggregated by upstream over `EventSelection`
(returning total requests, avg, p50, p95, p99, and max microseconds). `EventSelection`
is the shared message representing resource, request, filter, and time bounds
criteria.

`GetEventHistogramResponse`, `GetUniqueVisitorHistogramResponse`, and
`ListVisitorsResponse` return `from`, `to`, and `truncated` (true when the range
holds more than 1000 buckets and earlier buckets were truncated).

Paging uses `types.v1.Pages{index, size, total}`; over REST pass it as query
`page.index=1&page.size=20&page.total=true`.

## 8.5 Key data types

### `setting.v1.Config`

The runtime config passed to every constructor:

```proto
message Config {
  XMeta meta = 1;                // from the service's x_meta
  Flag flag = 2;                 // command-line flags
  map<string, string> env = 3;   // SENZ_* variables
}
```

Go helpers `conf.Get(key)` / `conf.GetDefault(key, def)` read from `env`.

### `dmz.edge.v1.Setting`

Edge config for one namespace — see
[03-edge.md](03-edge.md#35-proxyyaml-configuration).

### `dmz.edge.v1.Context`

`{metadata, transport: net.v1.Transport, request: net.http.v1.Request, x:
ContextExtra{namespace, ...matched ids/names, upstream,
upstream_timestamp, upstream_latency_micros}}` — per-request edge state,
attached to `szhttp.Context`. The router (`ROU`) sets the three `upstream`
fields around the reverse proxy call, for the egress log.

### Log events

| Type | Logger kind | Emitted by |
|---|---|---|
| `net.http.v1.IngressEvent` | `LOG_TYPE_HTTP_INGRESS` | Edge `LOG` node, apiserver `Logging` middleware |
| `net.http.v1.EgressEvent` | `LOG_TYPE_HTTP_EGRESS` | Edge `LOG` node, for a request proxied to its upstream origin; same `id` as its ingress event. Field `headers` records the request headers set on the upstream |
| `secrule.v1.Event` | `LOG_TYPE_CORE_RULESETS` | Edge `WAF` node when a transaction is interrupted |

`ContextExtra` (`dmz/edge/v1/context.proto`) records per-request edge metadata
including `set_headers` (field 20), populated via `szhttp.SetUpstreamHeader`
when the edge mutates or injects headers forwarded to an origin.

Events are reused through pools in
`pkg/pools/{ingress,egress,secrule,coreruleset,matchedrule}`: each wrapper keeps
its message, and `Release` calls `ResetVT` so the repeated fields keep
their capacity. `LogCloser` returns the object to its pool after logging.
The collector has its own pools, see
[07-controlplane.md](07-controlplane.md#event-pools).
