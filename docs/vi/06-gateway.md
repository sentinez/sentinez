# 6. Gateway: API server & realtime

## 6.1 API server (`szapiserver`)

| Đường dẫn | Vai trò |
|---|---|
| [cmd/szapiserver/main.go](../../cmd/szapiserver/main.go) | Inject `config.Config`, `grpcgateway.NewServer`, `apiserver.New` |
| [pkg/apps/gateway/apiserver/apiserver.go](../../pkg/apps/gateway/apiserver/apiserver.go) | `Server.Start/Shutdown`, `Visit`, `VisitToEndpoint` |
| [pkg/apps/gateway/apiserver/initialize.go](../../pkg/apps/gateway/apiserver/initialize.go) | Middleware, Swagger, đăng ký service |
| [internal/apps/apiserver/controlplane/v1/](../../internal/apps/apiserver/controlplane/v1/) | Registrar cho Analytic, Communication, Greeter, IAM, Security, Tenant |
| [internal/apps/apiserver/middleware/](../../internal/apps/apiserver/middleware/) | `Logging`, `CORS`, `SwaggerDefaultRoot`, `Injection` (SQLi/XSS), `Scope` |
| [internal/apps/apiserver/handlers/openapi_swagger.go](../../internal/apps/apiserver/handlers/openapi_swagger.go) | Phục vụ OpenAPI JSON + Swagger UI |
| [internal/apps/apiserver/handlers/sse.go](../../internal/apps/apiserver/handlers/sse.go) | Chuyển `CommunicationService.ServerStreamNotification` thành server-sent events tại `defaults.SSEPath` (xem [Stream thông báo](#stream-thông-báo-sse)) |
| [core/szgrpc/grpcgateway/](../../staging/src/github.com/sentinez/core/szgrpc/grpcgateway/) | `XServer` (HTTP mux + runtime mux, HTTP/1.1/2 + HTTP/3), `ServiceRegistrar` |

### Khởi tạo

```go
srv.server.Use(middleware.Logging)
srv.server.Use(middleware.CORS(passkey.ClientOrigins(conf)))
srv.server.Use(middleware.SwaggerDefaultRoot)
srv.server.Use(middleware.Scope(
    middleware.Injection, "/iam", "/tenant", "/greeter", "/analytic",
    "/communication",
))
handlers.RegisterSwaggerRoutes(srv.server.HTTPMux(), flag)
// security, tenant, analytic: xxxfac.NewDefaultHandler(ctx, conf)
// trả về (handler, error); lỗi khởi tạo repo làm dừng startup.
srv.Visit(ctx,
    controlplane.NewGreeter(greeterfac.NewDefaultHandler(conf)),
    controlplane.NewIAM(iamfac.NewDefaultHandler(ctx, conf)),
    controlplane.NewSecurity(security),
    controlplane.NewTenant(tenant),
    controlplane.NewAnalytic(analytic),
    // handler của useNotifications, cũng nằm sau route SSE
    controlplane.NewCommunication(communication),
)
```

`Injection` chỉ chạy dưới các prefix của `Scope`; `/security` được bỏ ra vì
payload SecRule chứa pattern tấn công một cách hợp lệ.

`XServer.ListenAndServe(addr, opts...)` listen trên `SZ_ADDRESS` (mặc định
`defaults.APIServerAddress`, `:9000` bên trong `grpcgateway`). Listener và TLS
config lấy từ option `szhttp` nếu có (`WithStdListener`, `WithQuicListener`,
`WithTLSConfig`). Có certificate (`--cert_file`/`--cert_key_file`, hoặc có sẵn
trong TLS config) thì phục vụ HTTPS qua TCP và HTTP/3 qua UDP trên cùng địa
chỉ, quảng bá HTTP/3 bằng `Alt-Svc`; không có thì phục vụ HTTP thường.
`Shutdown` dừng cả hai.

`ServiceRegistrar` có hai cách gắn service vào `runtime.ServeMux`:

| Phương thức | Cơ chế | Đang dùng |
|---|---|---|
| `Accept(ctx, server)` | Đăng ký handler Go lên `grpcgateway.NewLocalConn()` trong cùng tiến trình rồi gọi `RegisterXxxHandlerClient`, nên interceptor gRPC (validate) vẫn chạy | **Có** (`Visit`) |
| `AcceptFromEndpoint(ctx, server, conf)` | Mỗi 10s hỏi Consul (`discovery.Discover(serviceKey)`) rồi `RegisterXxxHandlerFromEndpoint` qua gRPC | Chưa (`VisitToEndpoint` không được gọi) |

Như vậy apiserver hiện là **monolith**: IAM/Tenant/Security/Analytic/
Communication/Greeter chạy trong tiến trình apiserver và dùng chung pool
PostgreSQL.

### Request pipeline HTTP

```
http.Server.Handler = extendHeader( Logging( CORS( httpMux ) ) )
httpMux:
  /api/      → OpenAPI JSON (api/docs/v1, embed)
  /swagger   → 301 /swagger/
  /swagger/  → Swagger UI (api/docs/swagger, embed); *.json → 307 /api/...
  GET /communication/notifications/stream → SSE thông báo (sse.go)
  /          → grpc-gateway runtime.ServeMux
```

- `Logging`: đọc toàn bộ body (để log khi lỗi), gọi handler, log debug nếu
  status ≥ 400, rồi ghi `httppb.IngressEvent` (kind `LOG_TYPE_HTTP_INGRESS`) qua `szlog`.
- `CORS(origins)`: chỉ `Origin` nằm trong `SZ_CLIENT_ORIGIN` (một, hoặc
  nhiều giá trị cách nhau bằng dấu phẩy) mới được phản chiếu cùng
  `Allow-Credentials: true`; preflight của nó trả `204` với methods `GET,
  HEAD, POST, PUT, DELETE` và headers `Content-Type, Accept, Authorization`.
  Origin khác không có header CORS nào, và khi biến này không đặt thì mọi
  origin đều không có.
- Error handler của gateway map `szerror.ErrNotFound/ErrInvalidData/
  ErrUnimplemented` sang gRPC status tương ứng trước khi chuyển thành HTTP.

Header `Authorization` được grpc-gateway chuyển thành gRPC metadata
`authorization`, handler đọc bằng `headers.GetAuth(ctx)`.

### Stream thông báo (SSE)

`GET /communication/notifications/stream[?tenantIds=…]`
(`defaults.SSEPath`, đăng ký bởi `handlers.RegisterServerSentEvents` trong
`useNotifications` của `initialize.go`):

- Handler chuyển header `Authorization` thành metadata rồi mở
  `ServerStreamNotification` qua `client.NewLocalCommunication`: một server
  bufconn có `interceptor.ServerOptions()`, nên stream được `AuthorizeStream`
  xác thực như một lời gọi qua mạng. `grpcgateway.LocalConn` chưa chở được
  stream (không hỗ trợ streaming) nên phải đi đường riêng này, hiện chưa có
  interceptor audit và panic recovery
  ([#433](https://github.com/sentinez/sentinez/issues/433)).
- Handler chờ header của stream rồi mới trả lời: lời gọi bị từ chối thành
  HTTP status tương ứng (401, 403…) qua `runtime.HTTPStatusFromCode`.
- Sau đó trả `200 text/event-stream`, `Cache-Control: no-cache, no-transform`
  (không có thì proxy nén, ví dụ proxy của console, giữ event trong bộ đệm
  gzip) và `X-Accel-Buffering: no`. Mỗi thông báo là một event `id: <id>` /
  `event: notification` / `data: <protojson>`; stream rảnh nhận comment
  `: ping` mỗi 15s.
- `/communication` nằm trong scope của `Injection`, middleware này chỉ kiểm
  tra request nên stream vẫn flush được.

## 6.2 Realtime (`szrealtime`)

| Đường dẫn | Vai trò |
|---|---|
| [pkg/apps/gateway/realtime/realtime.go](../../pkg/apps/gateway/realtime/realtime.go) | Đăng ký route `/ws` và listen |
| [pkg/apps/gateway/realtime/handlers/realtime.go](../../pkg/apps/gateway/realtime/handlers/realtime.go) | Upgrade WebSocket, vòng đọc tin nhắn |
| [pkg/network/wsz/](../../pkg/network/wsz/) | `WebSocket` server (dựa trên `net/http` mặc định) và `Manager` client |

Luồng: client kết nối `GET /ws?id=<clientId>` → upgrade (gorilla/websocket,
chấp nhận mọi origin) → `Manager.AddClient(id, conn)` → mỗi tin nhắn nhận
được broadcast tới mọi client dưới dạng `From <id>: <msg>` → khi lỗi đọc thì
`RemoveClient`. Thiếu `id` → lỗi `InvalidData`.

`wsz.WebSocket.ListenAndServe` đăng ký handler lên `http.DefaultServeMux` và
`Shutdown` hiện là no-op.

## 6.3 Service discovery (Consul)

[core/szgrpc/client/](../../staging/src/github.com/sentinez/core/szgrpc/client/):

| Gói | Vai trò |
|---|---|
| `consul` | Wrapper `hashicorp/consul/api`: `Register` (HTTP check), `RegisterWithTTL`, `SendHeartbeat`, `Discover` (healthy only) |
| `resolver` | Cache instance 10s/tên, chọn round-robin; tạo circuit breaker theo key |
| `discovery` | Singleton `GetDiscovery(opt)`: `Register` (ID = UUID, TTL), `Heartbeat`, `Discover` → `host:port` |
| `connection` | `Conn(addr)` gRPC insecure; `BufConn(bufLis)` qua `bufconn` |
| `local` | `RegisterServiceServer` chạy một gRPC server in-memory (`bufconn`, 1 MiB) cho mỗi service key, có thể kèm server option (lấy theo lần gọi đầu của key) |
| [client.go](../../staging/src/github.com/sentinez/core/szgrpc/client/client.go) | `NewIAM`, `NewCommunication` (qua Consul), `NewLocalIAM`, `NewLocalGreeter`, `NewLocalCommunication`, `NewLocalRevProxyService` (qua bufconn; `NewLocalCommunication` nhận server option, ví dụ `interceptor.ServerOptions()`) |

Phía server, `szgrpc.Server.Serve` sẽ gọi `Register` (đăng ký TTL 15s, retry
qua circuit breaker, heartbeat mỗi 10s) **nếu** tạo server với
`szgrpc.WithDiscorvery(true)` — hiện chưa service nào bật.

IAM dùng `client.NewLocalGreeter` để gọi Greeter qua bufconn trong RPC
`Status`.
