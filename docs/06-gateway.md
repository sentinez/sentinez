# 6. Gateway: API server & realtime

## 6.1 API server (`szapiserver`)

| Path | Role |
|---|---|
| [cmd/szapiserver/main.go](../cmd/szapiserver/main.go) | Injects `config.Config`, `grpcgateway.NewServer`, `apiserver.New` |
| [pkg/apps/gateway/apiserver/apiserver.go](../pkg/apps/gateway/apiserver/apiserver.go) | `Server.Start/Shutdown`, `Visit`, `VisitToEndpoint` |
| [pkg/apps/gateway/apiserver/initialize.go](../pkg/apps/gateway/apiserver/initialize.go) | Middleware, Swagger, service registration |
| [pkg/apps/gateway/apiserver/controlplane/v1/](../pkg/apps/gateway/apiserver/controlplane/v1/) | Registrars for Greeter, IAM, Security, Tenant |
| [pkg/apps/gateway/apiserver/middleware/](../pkg/apps/gateway/apiserver/middleware/) | `Logging`, `AllowCORS` |
| [pkg/apps/gateway/apiserver/handlers/openapi_swagger.go](../pkg/apps/gateway/apiserver/handlers/openapi_swagger.go) | Serves OpenAPI JSON + Swagger UI |
| [core/grpc/gateway/](../staging/src/github.com/sentinez/core/grpc/gateway/) | `XServer` (HTTP mux + runtime mux), `ServiceRegistrar` |

### Initialization

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

A `ServiceRegistrar` can attach a service to `runtime.ServeMux` in two ways:

| Method | Mechanism | In use |
|---|---|---|
| `Accept(ctx, server)` | `RegisterXxxHandlerServer` — calls the Go handler directly in the same process | **Yes** (`Visit`) |
| `AcceptFromEndpoint(ctx, server, conf)` | Every 10s asks Consul (`discovery.Discover(serviceKey)`) and calls `RegisterXxxHandlerFromEndpoint` over gRPC | No (`VisitToEndpoint` is never called) |

So the apiserver is currently a **monolith**: IAM/Tenant/Security/Greeter run
inside the apiserver process and share one PostgreSQL pool.
`AnalyticService` has a handler and an OpenAPI spec but is **not registered**
in the apiserver.

### HTTP request pipeline

```
http.Server.Handler = extendHeader( Logging( AllowCORS( httpMux ) ) )
httpMux:
  /api/      → OpenAPI JSON (api/docs/v1, embedded)
  /swagger   → 301 /swagger/
  /swagger/  → Swagger UI (api/docs/swagger, embedded); *.json → 307 /api/...
  /          → grpc-gateway runtime.ServeMux
```

- `Logging`: reads the whole body (so it can be logged on error), calls the
  handler, debug-logs when status ≥ 400, then writes an `httppb.Event` (kind
  `LOG_TYPE_HTTP`) through `zlog`.
- `AllowCORS`: reflects any `Origin`, sets `Allow-Credentials: true`, returns
  `204` for preflight with methods `GET, HEAD, POST, PUT, DELETE` and headers
  `Content-Type, Accept, Authorization`.
- The gateway error handler maps `errorx.ErrNotFound/ErrInvalidData/
  ErrUnimplemented` to the matching gRPC status before converting to HTTP.

grpc-gateway forwards the `Authorization` header as gRPC metadata
`authorization`; handlers read it with `headers.GetAuth(ctx)`.

## 6.2 Realtime (`szrealtime`)

| Path | Role |
|---|---|
| [pkg/apps/gateway/realtime/realtime.go](../pkg/apps/gateway/realtime/realtime.go) | Registers route `/ws` and listens |
| [pkg/apps/gateway/realtime/handlers/realtime.go](../pkg/apps/gateway/realtime/handlers/realtime.go) | WebSocket upgrade, message read loop |
| [pkg/network/wsz/](../pkg/network/wsz/) | `WebSocket` server (on the default `net/http` mux) and client `Manager` |

Flow: a client connects to `GET /ws?id=<clientId>` → upgrade
(gorilla/websocket, any origin accepted) → `Manager.AddClient(id, conn)` →
every received message is broadcast to all clients as `From <id>: <msg>` →
on read error, `RemoveClient`. A missing `id` → `InvalidData` error.

`wsz.WebSocket.ListenAndServe` registers handlers on `http.DefaultServeMux`,
and `Shutdown` is currently a no-op.

## 6.3 Service discovery (Consul)

[api/client/](../api/client/):

| Package | Role |
|---|---|
| `consul` | Wrapper over `hashicorp/consul/api`: `Register` (HTTP check), `RegisterWithTTL`, `SendHeartbeat`, `Discover` (healthy only) |
| `resolver` | Caches instances for 10s per name, round-robin selection; per-key circuit breakers |
| `discovery` | Singleton `GetDiscovery(opt)`: `Register` (ID = UUID, TTL), `Heartbeat`, `Discover` → `host:port` |
| `connection` | `Conn(addr)` insecure gRPC; `BufConn(bufLis)` via `bufconn` |
| `local` | `RegisterServiceServer` runs an in-memory gRPC server (`bufconn`, 1 MiB) per service key |
| [client.go](../api/client/client.go) | `NewIAM` (via Consul), `NewLocalIAM`, `NewLocalGreeter`, `NewLocalEdgeEngine` (via bufconn) |

On the server side, `coregrpc.Server.Serve` calls `Register` (TTL 15s,
retries through a circuit breaker, heartbeat every 10s) **only if** the
server was built with `coregrpc.WithDiscorvery(true)` — no service enables it
yet.

IAM uses `client.NewLocalGreeter` to call Greeter over bufconn in its
`Status` RPC.
