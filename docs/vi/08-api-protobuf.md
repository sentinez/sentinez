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
| `x_message` (51001) | Message | `database_model`, `export_field`, `scanner` | Sinh hằng tên cột `<Msg>_<Field>` / `X<Msg>_<Field>`; `scanner` sinh thêm row scanner bên dưới |
| `x_field` (53001) | Field | `flatten`, `ignore`, `codec`, `nullable` | Cách một field của model DB ánh xạ ra cột, xem bên dưới |
| `x_method` (50001) | RPC | `ignore`, `control_planes[]` | Sinh `Get<Service><Method>() *XMethod` cho phân quyền |

`Zone`: `DEMILITARIZED` (edge, dataplane), `INTERNAL` (domain service),
`PRIVATE_API` (apiserver, realtime), `PUBLIC_API`.

Plugin xử lý các option này là `protoc-gen-go-senz`
([tools/internal/senz/senz.go](../../staging/src/github.com/sentinez/tools/internal/senz/senz.go)),
sinh file `*_senz.pb.go`.

### Row scanner được sinh

Với message có `x_message.scanner = true`
([tools/internal/senz/scanner.go](../../staging/src/github.com/sentinez/tools/internal/senz/scanner.go)),
mọi field kiểu message lồng nhau được flatten thành cột `<field>_<nested>`
(ví dụ `metadata_created_at`, `owner_tenant_id`), và plugin sinh:

| Hàm | Làm gì |
|---|---|
| `<Msg>Columns() []string` | Các cột được lưu, theo thứ tự scan |
| `<Msg>Values(m) ([]any, error)` | Giá trị của `m` theo cùng thứ tự, để insert |
| `ScanOne<Msg>(row) (*<Msg>, error)` | Scan một dòng (bất kỳ `Scan(dest ...any) error`: `pgx.Row`, `dbx.Row`) |
| `Scan<Msg>(rows) ([]*<Msg>, error)` | Scan mọi dòng (`Next`/`Scan`/`Err`: `pgx.Rows`, `dbx.Rows`) |

Cách scan một field, đặt bằng `x_field`:

| Field | Cột |
|---|---|
| scalar, enum | scan thẳng vào field |
| `google.protobuf.Timestamp` / `Duration` | `time.Time` / `time.Duration`, rồi chuyển đổi |
| message khác, không codec | flatten thành các cột của field con |
| `codec = COLUMN_CODEC_PROTO` | một cột protobuf nhị phân (`proto.Unmarshal`), ví dụ `Resource.server` |
| `codec = COLUMN_CODEC_PROTOJSON` | một cột protobuf JSON (`protojson.Unmarshal`, bỏ field lạ), ví dụ `CoreRuleset.content` |
| `codec = COLUMN_CODEC_DRIVER` | một cột driver tự scan vào field (ví dụ JSONB), ví dụ `SecRule.expr`, `SecRuleEvent.action` |
| `ignore = true` | không lưu, bị bỏ qua (ví dụ `SecRule.action_value`, `expr_preview`) |
| `nullable = true` | scalar đơn có cột có thể `NULL`, scan thành zero value (ví dụ `Owner.tenant_id`, `Tenant.name`) |

`flatten` (có từ trước) chỉ tách hằng tên cột của message lồng nhau; với
`scanner` thì mọi message lồng nhau đều được flatten. Plugin báo lỗi khi
gặp field thuộc oneof, `nullable` trên field không phải scalar, và `codec`
trên field không phải message. Test:
[tools/example/example_scanner_test.go](../../staging/src/github.com/sentinez/tools/example/example_scanner_test.go).

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
| `protoc-gen-go-vtproto` (`features=marshal+unmarshal+size`) | `*_vtproto.pb.go` |
| `protoc-gen-openapiv2` | `api/docs/v1/<domain>.swagger.json` |

Feature `pool` của vtproto chỉ bật cho các event được pool, mỗi message ghi
bằng một `--go-vtproto_opt=pool=<import path>.<Message>` trong
`generate.sh` của nó: `secrule.v1.SecRuleEvent`, `coreruleset.v1.CoreRuleEvent`,
`net.http.v1.IngressEvent` và `rule.v1.MatchedRules`. Feature này sinh `ResetVT`
(`Reset` nhưng giữ dung lượng các trường repeated), cùng
`<Msg>FromVTPool`/`ReturnToVTPool` trên một pool của message.

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
| GET | `/iam/passkey/link/challenge` | `...PasskeyLinkChallenge` | ignore (service yêu cầu token) |
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

"Quyền khai báo" là option trong proto; việc thực thi phụ thuộc handler (xem
[07-controlplane.md](07-controlplane.md#72-xác-thực--phân-quyền)).

`GetEgressLatency` truy vấn độ trễ gom nhóm theo upstream qua `EventSelection`
(trả về tổng request, avg, p50, p95, p99 và max micro giây). `EventSelection`
là message dùng chung biểu diễn các tiêu chí lọc resource, request, filter
và khoảng thời gian.

`GetEventHistogramResponse`, `GetUniqueVisitorHistogramResponse`, và
`ListVisitorsResponse` trả về `from`, `to`, và `truncated` (true khi khoảng
thời gian vượt quá 1000 bucket và các bucket trước đó đã bị cắt bớt).

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
ContextExtra{namespace, ...matched ids/names, upstream,
upstream_timestamp, upstream_latency_micros}}` — trạng thái của một request ở
edge, gắn với `szhttp.Context`. Router (`ROU`) ghi ba trường `upstream` quanh
lời gọi reverse proxy, để log egress dùng.

### Event log

| Type | Logger kind | Phát bởi |
|---|---|---|
| `net.http.v1.IngressEvent` | `LOG_TYPE_HTTP_INGRESS` | Node `LOG` ở edge, middleware `Logging` ở apiserver |
| `net.http.v1.EgressEvent` | `LOG_TYPE_HTTP_EGRESS` | Node `LOG` ở edge, với request được proxy tới upstream origin; cùng `id` với event ingress của nó. Trường `headers` ghi nhận các header được gửi tới upstream |
| `secrule.v1.Event` | `LOG_TYPE_CORE_RULESETS` | Node `WAF` khi transaction bị interrupt |

`ContextExtra` (`dmz/edge/v1/context.proto`) lưu metadata theo từng request của edge,
bao gồm `set_headers` (trường 20), được điền qua `szhttp.SetUpstreamHeader` khi
edge thêm hoặc sửa đổi header chuyển tiếp tới origin.

Event được tái sử dụng qua pool trong
`pkg/pools/{ingress,egress,secrule,coreruleset,matchedrule}`: mỗi wrapper giữ sẵn
message của nó, và `Release` gọi `ResetVT` nên các trường repeated giữ
dung lượng. `LogCloser` tự trả object về pool sau khi log. Collector có
pool riêng, xem [07-controlplane.md](07-controlplane.md#pool-event).
