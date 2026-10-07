# 6. Gateway: API server & realtime

## 6.1 API server (`szapiserver`)

| Đường dẫn | Vai trò |
|---|---|
| [cmd/szapiserver/main.go](../../cmd/szapiserver/main.go) | Inject `config.Config`, `grpcgateway.NewServer`, `apiserver.New` |
| [pkg/apps/gateway/apiserver/apiserver.go](../../pkg/apps/gateway/apiserver/apiserver.go) | `Server.Start/Shutdown`, `Visit`, `VisitToEndpoint` |
| [pkg/apps/gateway/apiserver/initialize.go](../../pkg/apps/gateway/apiserver/initialize.go) | Middleware, Swagger, đăng ký service |
| [pkg/apps/gateway/apiserver/controlplane/v1/](../../pkg/apps/gateway/apiserver/controlplane/v1/) | Registrar cho Greeter, IAM, Security, Tenant |
| [pkg/apps/gateway/apiserver/middleware/](../../pkg/apps/gateway/apiserver/middleware/) | `Logging`, `AllowCORS` |
| [pkg/apps/gateway/apiserver/handlers/openapi_swagger.go](../../pkg/apps/gateway/apiserver/handlers/openapi_swagger.go) | Phục vụ OpenAPI JSON + Swagger UI |
| [core/grpc/gateway/](../../staging/src/github.com/sentinez/core/grpc/gateway/) | `XServer` (HTTP mux + runtime mux), `ServiceRegistrar` |

### Khởi tạo

```go
srv.server.Use(middleware.Logging)
srv.server.Use(middleware.AllowCORS)
handlers.RegisterSwaggerRoutes(srv.server.HTTPMux(), flag)
srv.Visit(ctx,
    controlplane.NewGreeter(greeterfac.NewDefaultHandler(conf)),
    controlplane.NewIAM(iamfac.NewDefaultHandler(ctx, conf)),
    controlplane.NewSecurity(securityfac.NewDefaultHandler(ctx, conf)),
    controlplane.NewTenant(tenantfac.NewDefaultHandler(ctx, conf)),
)
```

`ServiceRegistrar` có hai cách gắn service vào `runtime.ServeMux`:

| Phương thức | Cơ chế | Đang dùng |
|---|---|---|
| `Accept(ctx, server)` | `RegisterXxxHandlerServer` — gọi handler Go trực tiếp trong cùng tiến trình | **Có** (`Visit`) |
| `AcceptFromEndpoint(ctx, server, conf)` | Mỗi 10s hỏi Consul (`discovery.Discover(serviceKey)`) rồi `RegisterXxxHandlerFromEndpoint` qua gRPC | Chưa (`VisitToEndpoint` không được gọi) |

Như vậy apiserver hiện là **monolith**: IAM/Tenant/Security/Greeter chạy trong
tiến trình apiserver và dùng chung pool PostgreSQL. `AnalyticService` có
handler và OpenAPI nhưng **chưa được đăng ký** vào apiserver.

### Request pipeline HTTP

```
http.Server.Handler = extendHeader( Logging( AllowCORS( httpMux ) ) )
httpMux:
  /api/      → OpenAPI JSON (api/docs/v1, embed)
  /swagger   → 301 /swagger/
  /swagger/  → Swagger UI (api/docs/swagger, embed); *.json → 307 /api/...
  /          → grpc-gateway runtime.ServeMux
```

- `Logging`: đọc toàn bộ body (để log khi lỗi), gọi handler, log debug nếu
  status ≥ 400, rồi ghi `httppb.Event` (kind `LOG_TYPE_HTTP`) qua `zlog`.
- `AllowCORS`: phản chiếu `Origin` bất kỳ, `Allow-Credentials: true`, trả
  `204` cho preflight với methods `GET, HEAD, POST, PUT, DELETE` và headers
  `Content-Type, Accept, Authorization`.
- Error handler của gateway map `errorx.ErrNotFound/ErrInvalidData/
  ErrUnimplemented` sang gRPC status tương ứng trước khi chuyển thành HTTP.

Header `Authorization` được grpc-gateway chuyển thành gRPC metadata
`authorization`, handler đọc bằng `headers.GetAuth(ctx)`.

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

[api/client/](../../api/client/):

| Gói | Vai trò |
|---|---|
| `consul` | Wrapper `hashicorp/consul/api`: `Register` (HTTP check), `RegisterWithTTL`, `SendHeartbeat`, `Discover` (healthy only) |
| `resolver` | Cache instance 10s/tên, chọn round-robin; tạo circuit breaker theo key |
| `discovery` | Singleton `GetDiscovery(opt)`: `Register` (ID = UUID, TTL), `Heartbeat`, `Discover` → `host:port` |
| `connection` | `Conn(addr)` gRPC insecure; `BufConn(bufLis)` qua `bufconn` |
| `local` | `RegisterServiceServer` chạy một gRPC server in-memory (`bufconn`, 1 MiB) cho mỗi service key |
| [client.go](../../api/client/client.go) | `NewIAM` (qua Consul), `NewLocalIAM`, `NewLocalGreeter`, `NewLocalEdgeEngine` (qua bufconn) |

Phía server, `coregrpc.Server.Serve` sẽ gọi `Register` (đăng ký TTL 15s, retry
qua circuit breaker, heartbeat mỗi 10s) **nếu** tạo server với
`coregrpc.WithDiscorvery(true)` — hiện chưa service nào bật.

IAM dùng `client.NewLocalGreeter` để gọi Greeter qua bufconn trong RPC
`Status`.
