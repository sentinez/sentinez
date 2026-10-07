# 8. API & Protobuf

Module `github.com/sentinez/sentinez/api` ([api/](../../api/)) là nguồn sự thật
cho mọi kiểu dữ liệu: cấu hình service, cấu hình edge, rule, event log và
API REST/gRPC.

## 8.1 Cấu trúc thư mục proto

```
api/proto/sentinez/
├── apps/<domain>/v1/        # domain service (REST qua grpc-gateway)
│   ├── <domain>.proto       #   service + request/response
│   └── model.proto          #   model DB (x_message.database_model)
├── dmz/
│   ├── edge/v1/             # EdgeService, Setting (proxy.yaml), Context
│   └── dataplane/v1/        # DataPlaneService
├── gateway/
│   ├── apiserver/v1/        # chỉ khai báo x_meta
│   └── realtime/v1/         # chỉ khai báo x_meta
└── types/                   # kiểu dùng chung
    ├── v1/                  # known.proto (enum), model.proto (Pages, Metadata), options.proto
    ├── setting/v1/          # Config, Flag, enum Senz (biến môi trường)
    ├── rule/v1/             # Expression, Condition, Operator, Action
    ├── secrule/v1/          # SecRule runtime/Lite, Event WAF
    ├── cdn/v1/              # CDN rule
    ├── coreruleset/v1/      # CoreRulesets (đầu ra rule parser)
    ├── net/v1/              # Transport (conn_id, SNI)
    └── net/http/v1/         # Request, RequestEvent, Event
```

Go package: `github.com/sentinez/sentinez/api/proto/sentinez/<path>;<alias>pb`
(vd. `settingpb`, `rulepb`, `edgepb`, `iampb`).

Quy tắc `buf lint` ([api/buf.yaml](../../api/buf.yaml)): `STANDARD`, enum zero
value hậu tố `_UNSPECIFIED`, service hậu tố `Service`, không cho request và
response trùng type.

## 8.2 Custom option (`types/v1/options.proto`)

| Extension | Đặt trên | Trường | Dùng để |
|---|---|---|---|
| `x_meta` (52001) | File | `service_name`, `service_zone`, `service_key` | Định danh service: banner console, tên logger, key Consul, `GetMeta<Name>()` |
| `x_message` (51001) | Message | `database_model`, `export_field` | Sinh hằng tên cột `<Msg>_<Field>` / `X<Msg>_<Field>` |
| `x_method` (50001) | RPC | `ignore`, `control_planes[]` | Sinh `Get<Service><Method>() *XMethod` cho phân quyền |

`Zone`: `DEMILITARIZED` (edge, dataplane), `INTERNAL` (domain service),
`PRIVATE_API` (apiserver, realtime), `PUBLIC_API`.

Plugin xử lý các option này là `protoc-gen-go-senz`
([tools/internal/senz/senz.go](../../staging/src/github.com/sentinez/tools/internal/senz/senz.go)),
sinh file `*_senz.pb.go`.

## 8.3 Sinh code

Mỗi thư mục domain có `generate.sh` gọi `protoc` với include
`api/proto`, `api/third_party/{googleapis,grpc-gateway,protovalidate}`:

| Plugin | Đầu ra |
|---|---|
| `protoc-gen-go` | `*.pb.go` |
| `protoc-gen-go-grpc` (`require_unimplemented_servers=false`) | `*_grpc.pb.go` |
| `protoc-gen-grpc-gateway` | `*.pb.gw.go` |
| `protoc-gen-validate` | `*.pb.validate.go` |
| `protoc-gen-go-senz` | `*_senz.pb.go` |
| `protoc-gen-openapiv2` | `api/docs/v1/<domain>.swagger.json` |

TypeScript cho UI: `cd api && buf generate` dùng `ts-proto`
([api/buf.gen.yaml](../../api/buf.gen.yaml)) → `ui/packages/proto/ts-proto`.

Một số trường có comment `// @gotags: yaml:"..."` để `protoc-go-inject-tag`
thêm struct tag YAML.

[api/docs.go](../../api/docs.go) nhúng `docs/v1` (OpenAPI) và `docs/swagger`
(Swagger UI tĩnh) để apiserver phục vụ.

## 8.4 Bảng REST endpoint

Base URL: địa chỉ apiserver (mặc định `:8080` trong `.env.example`). Xác thực:
header `Authorization: Bearer <JWT>`.

| Method | Path | RPC | Quyền khai báo |
|---|---|---|---|
| GET | `/greeter/say` | `GreeterService.SayHello` | ignore |
| GET | `/greeter/status` | `GreeterService.Status` | ignore |
| PUT | `/iam/login` | `IdentityAccessManagementService.Login` | ignore |
| POST | `/iam/account` | `...CreateAccount` | ignore |
| GET | `/iam/accounts` | `...ListAccounts` | ignore (handler vẫn yêu cầu token) |
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
| GET | `/analytic/activities` | `AnalyticService.ListActivities` | ignore *(chưa mount)* |

"Quyền khai báo" là option trong proto; việc thực thi phụ thuộc handler (xem
[07-controlplane.md](07-controlplane.md#72-xác-thực--phân-quyền)).

Phân trang dùng `types.v1.Pages{index, size, total}`; qua REST truyền dạng
query `page.index=1&page.size=20&page.total=true`.

## 8.5 Kiểu dữ liệu quan trọng

### `setting.v1.Config`

Cấu hình runtime truyền vào mọi constructor:

```proto
message Config {
  XMeta meta = 1;                // từ x_meta của service
  Flag flag = 2;                 // command-line flags
  map<string, string> env = 3;   // biến SENZ_*
}
```

Helper Go `conf.Get(key)` / `conf.GetDefault(key, def)` đọc từ `env`.

### `dmz.edge.v1.Setting`

Cấu hình edge cho một namespace — xem [03-edge.md](03-edge.md#35-cấu-hình-proxyyaml).

### `dmz.edge.v1.Context`

`{metadata, transport: net.v1.Transport, request: net.http.v1.Request, x:
ContextExtra{namespace, ...matched ids/names}}` — trạng thái của một request ở
edge, gắn với `corehttp.Context`.

### Event log

| Type | Logger kind | Phát bởi |
|---|---|---|
| `net.http.v1.Event` | `LOG_TYPE_HTTP` | Node `LOG` ở edge, middleware `Logging` ở apiserver |
| `secrule.v1.Event` | `LOG_TYPE_WAF` | Node `WAF` khi transaction bị interrupt |

Event được tái sử dụng qua pool trong `pkg/pools/{request,ruleevent}`;
`LogCloser` tự trả object về pool sau khi log.
