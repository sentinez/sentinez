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
| `x_message` (51001) | Message | `database_model`, `export_field` | Generates column-name constants `<Msg>_<Field>` / `X<Msg>_<Field>` |
| `x_method` (50001) | RPC | `ignore`, `control_planes[]` | Generates `Get<Service><Method>() *XMethod` for authorization |

`Zone`: `DEMILITARIZED` (edge, dataplane), `INTERNAL` (domain services),
`PRIVATE_API` (apiserver, realtime), `PUBLIC_API`.

These options are processed by the `protoc-gen-go-senz` plugin
([tools/internal/senz/senz.go](../staging/src/github.com/sentinez/tools/internal/senz/senz.go)),
which emits `*_senz.pb.go` files.

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
| `protoc-gen-openapiv2` | `api/docs/v1/<domain>.swagger.json` |

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
| GET | `/analytic/activities` | `AnalyticService.ListActivities` | ignore *(not mounted)* |

"Declared permission" is the option in the proto; enforcement depends on the
handler (see
[07-controlplane.md](07-controlplane.md#72-authentication--authorization)).

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
ContextExtra{namespace, ...matched ids/names}}` — per-request edge state,
attached to `corehttp.Context`.

### Log events

| Type | Logger kind | Emitted by |
|---|---|---|
| `net.http.v1.Event` | `LOG_TYPE_HTTP` | Edge `LOG` node, apiserver `Logging` middleware |
| `secrule.v1.Event` | `LOG_TYPE_WAF` | Edge `WAF` node when a transaction is interrupted |

Events are reused through pools in `pkg/pools/{request,ruleevent}`;
`LogCloser` returns the object to its pool after logging.
