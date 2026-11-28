# 6. Gateway: API server & realtime

## 6.1 API server (`szapiserver`)

| Path | Role |
|---|---|
| [cmd/szapiserver/main.go](../cmd/szapiserver/main.go) | Injects `config.Config`, `grpcgateway.NewServer`, `apiserver.New` |
| [pkg/apps/gateway/apiserver/apiserver.go](../pkg/apps/gateway/apiserver/apiserver.go) | `Server.Start/Shutdown`, `Visit`, `VisitToEndpoint` |
| [pkg/apps/gateway/apiserver/initialize.go](../pkg/apps/gateway/apiserver/initialize.go) | Middleware, Swagger, service registration |
| [internal/apps/apiserver/controlplane/v1/](../internal/apps/apiserver/controlplane/v1/) | Registrars for Analytic, Communication, Greeter, IAM, Security, Tenant |
| [internal/apps/apiserver/middleware/](../internal/apps/apiserver/middleware/) | `Logging`, `CORS`, `SwaggerDefaultRoot`, `Injection` (SQLi/XSS), `Scope` |
| [internal/apps/apiserver/handlers/openapi_swagger.go](../internal/apps/apiserver/handlers/openapi_swagger.go) | Serves OpenAPI JSON + Swagger UI |
| [internal/apps/apiserver/handlers/sse.go](../internal/apps/apiserver/handlers/sse.go) | Relays `CommunicationService.ServerStreamNotification` as server-sent events at `defaults.SSEPath` (see [Notification stream](#notification-stream-sse)) |
| [core/szgrpc/grpcgateway/](../staging/src/github.com/sentinez/core/szgrpc/grpcgateway/) | `XServer` (HTTP mux + runtime mux, HTTP/1.1/2 + HTTP/3), `ServiceRegistrar` |

### Initialization

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
// returns (handler, error); a repo init error aborts startup.
srv.Visit(ctx,
    controlplane.NewGreeter(greeterfac.NewDefaultHandler(conf)),
    controlplane.NewIAM(iamfac.NewDefaultHandler(ctx, conf)),
    controlplane.NewSecurity(security),
    controlplane.NewTenant(tenant),
    controlplane.NewAnalytic(analytic),
    // the handler of useNotifications, also behind the SSE route
    controlplane.NewCommunication(communication),
)
```

`Injection` runs only under the `Scope` prefixes; `/security` is left out
because SecRule payloads legitimately carry attack patterns.

`XServer.ListenAndServe(addr, opts...)` listens on `SZ_ADDRESS` (default
`defaults.APIServerAddress`, `:9000` inside `grpcgateway`). Listeners and TLS
config come from the `szhttp` options when set (`WithStdListener`,
`WithQuicListener`, `WithTLSConfig`). With a certificate
(`--cert_file`/`--cert_key_file`, or one in the TLS config) it serves HTTPS
over TCP and HTTP/3 over UDP on the same address, advertising HTTP/3 with
`Alt-Svc`; without one it serves plain HTTP. `Shutdown` stops both.

A `ServiceRegistrar` can attach a service to `runtime.ServeMux` in two ways:

| Method | Mechanism | In use |
|---|---|---|
| `Accept(ctx, server)` | Registers the Go handler on an in-process `grpcgateway.NewLocalConn()` and calls `RegisterXxxHandlerClient`, so gRPC interceptors (validation) still run | **Yes** (`Visit`) |
| `AcceptFromEndpoint(ctx, server, conf)` | Every 10s asks Consul (`discovery.Discover(serviceKey)`) and calls `RegisterXxxHandlerFromEndpoint` over gRPC | No (`VisitToEndpoint` is never called) |

So the apiserver is currently a **monolith**: IAM/Tenant/Security/Analytic/
Communication/Greeter run inside the apiserver process and share one
PostgreSQL pool.

### HTTP request pipeline

```
http.Server.Handler = extendHeader( Logging( CORS( httpMux ) ) )
httpMux:
  /api/      → OpenAPI JSON (api/docs/v1, embedded)
  /swagger   → 301 /swagger/
  /swagger/  → Swagger UI (api/docs/swagger, embedded); *.json → 307 /api/...
  GET /communication/notifications/stream → notification SSE (sse.go)
  /          → grpc-gateway runtime.ServeMux
```

- `Logging`: reads the whole body (so it can be logged on error), calls the
  handler, debug-logs when status ≥ 400, then writes an `httppb.IngressEvent` (kind
  `LOG_TYPE_HTTP_INGRESS`) through `szlog`.
- `CORS(origins)`: only an `Origin` listed in `SZ_CLIENT_ORIGIN` (one, or
  several separated by commas) is reflected with `Allow-Credentials: true`;
  its preflight gets `204` with methods `GET, HEAD, POST, PUT, DELETE` and
  headers `Content-Type, Accept, Authorization`. Any other origin gets no
  CORS header, and so does every origin when the variable is unset.
- The gateway error handler maps `szerror.ErrNotFound/ErrInvalidData/
  ErrUnimplemented` to the matching gRPC status before converting to HTTP.

grpc-gateway forwards the `Authorization` header as gRPC metadata
`authorization`; handlers read it with `headers.GetAuth(ctx)`.

### Notification stream (SSE)

`GET /communication/notifications/stream[?tenantIds=…]`
(`defaults.SSEPath`, registered by `handlers.RegisterServerSentEvents` in
`useNotifications` of `initialize.go`):

- The handler forwards the `Authorization` header as metadata and opens
  `ServerStreamNotification` through `client.NewLocalCommunication`: a
  bufconn server with `interceptor.ServerOptions()`, so the stream is
  authorized by `AuthorizeStream` as a networked call would be.
  `grpcgateway.LocalConn` cannot carry it (no streaming), hence the
  separate path: it has no audit interceptor nor panic recovery yet
  ([#433](https://github.com/sentinez/sentinez/issues/433)).
- It waits for the stream headers before answering: a refused call becomes
  its HTTP status (401, 403…) through `runtime.HTTPStatusFromCode`.
- Then `200 text/event-stream`, `Cache-Control: no-cache, no-transform`
  (a compressing proxy, e.g. the console's, would otherwise hold the events
  in its gzip buffer) and `X-Accel-Buffering: no`. Each notification is an
  event `id: <id>` / `event: notification` / `data: <protojson>`; an idle
  stream gets a `: ping` comment every 15s.
- `/communication` is under the `Injection` scope, which only inspects the
  request, so the stream still flushes.

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

[core/szgrpc/client/](../staging/src/github.com/sentinez/core/szgrpc/client/):

| Package | Role |
|---|---|
| `consul` | Wrapper over `hashicorp/consul/api`: `Register` (HTTP check), `RegisterWithTTL`, `SendHeartbeat`, `Discover` (healthy only) |
| `resolver` | Caches instances for 10s per name, round-robin selection; per-key circuit breakers |
| `discovery` | Singleton `GetDiscovery(opt)`: `Register` (ID = UUID, TTL), `Heartbeat`, `Discover` → `host:port` |
| `connection` | `Conn(addr)` insecure gRPC; `BufConn(bufLis)` via `bufconn` |
| `local` | `RegisterServiceServer` runs an in-memory gRPC server (`bufconn`, 1 MiB) per service key, with optional server options (those of the first call of a key) |
| [client.go](../staging/src/github.com/sentinez/core/szgrpc/client/client.go) | `NewIAM`, `NewCommunication` (via Consul), `NewLocalIAM`, `NewLocalGreeter`, `NewLocalCommunication`, `NewLocalRevProxyService` (via bufconn; `NewLocalCommunication` takes the server options, e.g. `interceptor.ServerOptions()`) |

On the server side, `szgrpc.Server.Serve` calls `Register` (TTL 15s,
retries through a circuit breaker, heartbeat every 10s) **only if** the
server was built with `szgrpc.WithDiscorvery(true)` — no service enables it
yet.

IAM uses `client.NewLocalGreeter` to call Greeter over bufconn in its
`Status` RPC.
