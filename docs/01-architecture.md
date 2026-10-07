# 1. Architecture overview

## 1.1 What Sentinéz is

Sentinéz is a security reverse proxy that sits in front of a customer's
(tenant's) origin server. Each request that passes through the edge gets a
request ID, is mapped to a tenant by subdomain, and is checked against the
cache, rate limits, custom rules (SecRule) and the OWASP CRS (Coraza WAF)
before it is forwarded to the origin. Behind it, management services (the
control plane) expose a REST API for the web console.

Current version: `v0.0.1-beta`
([core/core.go](../staging/src/github.com/sentinez/core/core.go)). The root
README warns that backward compatibility is not guaranteed before v1.0.0.

## 1.2 Component diagram

```
                    Internet
                       │
             ┌─────────▼─────────┐
             │  szdataplane (XDP)│  eBPF: per-IP bandwidth / IP blocking (netns "gateway")
             └─────────┬─────────┘
                       │
             ┌─────────▼─────────┐        Olric cluster (memberlist)
             │   szedge (HTTPS)  │◄──────► shares edge Settings between nodes
             │  Hertz + chain    │
             └─────────┬─────────┘
                       │ reverse proxy
                       ▼
                 Tenant origin


  Web console (Next.js) ──REST──► szapiserver (grpc-gateway, :8080)
                                     │ in-process gRPC handlers
                                     ├─ IAM        ─┐
                                     ├─ Tenant      │  PostgreSQL
                                     ├─ Security    ├─►(tables <env>_sentinez_*)
                                     └─ Greeter    ─┘
                    szrealtime (WebSocket /ws) — message broadcast
                    Consul — service discovery (optional)
```

## 1.3 Binaries (`cmd/`)

| Binary | Entry point | Service type (`runner.App[T]`) | Role |
|---|---|---|---|
| `szapiserver` | [cmd/szapiserver/main.go](../cmd/szapiserver/main.go) | `apiserver.Server` | HTTP gateway from REST to control-plane gRPC handlers; serves Swagger UI |
| `szrealtime` | [cmd/szrealtime/main.go](../cmd/szrealtime/main.go) | `realtime.Realtime` | WebSocket server that broadcasts messages between clients |
| `szedge/v1` | [cmd/szedge/v1/main.go](../cmd/szedge/v1/main.go) | `edge.Server` | HTTPS edge proxy (TLS 1.3, HTTP/2) running the security chain |
| `szdataplane/v1` | [cmd/szdataplane/v1/main.go](../cmd/szdataplane/v1/main.go) | `dataplane.Server` | Attaches the XDP program to `veth0`, opens a gRPC server |
| `szgreeter/v1` | [cmd/szgreeter/v1/main.go](../cmd/szgreeter/v1/main.go) | `greeter.Greeter` | Sample gRPC service (SayHello/Status) |
| `szcentraldata/v1` | [cmd/szcentraldata/v1/main.go](../cmd/szcentraldata/v1/main.go) | — | Empty `main()`, not implemented yet |

Every binary (except `szcentraldata`) follows the same pattern:

```go
app := runner.NewApp[T](config.Config(), core.Code)
app.Main(func(c *runner.Context[T]) {
    c.Inject(config.Config, <constructor>...)   // fx.Provide
    c.OnStart(func(ctx, *T) error { ... })      // runs in a goroutine
    c.OnStop(func(ctx, *T) error { ... })
})
```

See [09-core-shared.md](09-core-shared.md#91-corerunner) for `runner`.

## 1.4 Monorepo layout

The repo is **multi-module**, wired with `replace` directives in the root
[go.mod](../go.mod) (there is no `go.work`):

| Module | Path | Depends on |
|---|---|---|
| `github.com/sentinez/sentinez` (root) | `/` | all modules below |
| `github.com/sentinez/sentinez/api` | `api/` | — (protos + clients) |
| `github.com/sentinez/core` | `staging/.../core` | `api`, `shared` |
| `github.com/sentinez/shared` | `staging/.../shared` | `api`, `core` (some packages) |
| `github.com/sentinez/controlplane` | `staging/.../controlplane` | `api`, `core`, `shared` |
| `github.com/sentinez/contrib/httphz` | `staging/.../contrib/httphz` | `core`, `shared`, Hertz |
| `github.com/sentinez/bpf` | `staging/.../bpf` | cilium/ebpf |
| `github.com/sentinez/tools` | `staging/.../tools` | `core`, `api` |

Modules under `staging/` are pushed to their own repos
`github.com/sentinez/<name>` with `git subtree split` (see
[11-deploy-ci.md](11-deploy-ci.md#114-publishing-staging-modules)).

Top-level directories:

| Directory | Contents |
|---|---|
| `cmd/` | Per-service entry points, Dockerfiles, `.env.example`, `proxy.yaml` |
| `internal/` | Root-module private code: `dmz/edge`, `dmz/dataplane`, `memory`, `cluster`, `bpf`, `defaults` |
| `pkg/` | Service wiring (`apps/...`), networking (`network`), object pools (`pools`), `protocol`, `tracer`, `queue` |
| `api/` | Protos, generated code, gRPC clients, embedded OpenAPI/Swagger |
| `staging/` | Separately published modules |
| `ui/` | pnpm + turbo workspace (Next.js console, web, shared packages) |
| `deploy/` | Docker Compose, Caddy, nginx, Prometheus/Grafana, CRS rules (`ruleroot`) |
| `hack/` | Dev scripts: lint, rule generation, netns, certs, GitLab sync |
| `_submodules/` | coreruleset, googleapis, grpc-gateway, opentelemetry-proto, protovalidate |
| [sentinez.go](../sentinez.go) | Embeds the CRS v4.16.0 / v4.17.0 `*.data` files via `embed.FS` |

## 1.5 Main data flows

### End-user request (data path)

1. TCP accept in `network.Listener` → wrapped as a `network.Conn` with a
   random 8-byte hex `Id`.
2. TLS ClientHello → `transport.TLSConfig` computes the JA4 fingerprint and
   stores it in `shared/store/ja4` keyed by `conn.Id`.
3. `OnConnect` puts `netpb.Transport{ConnId, ServerName}` into the
   `context`.
4. Hertz calls the handler → an `httphz.Context` is taken from a pool and
   given the fingerprint.
5. The middleware chain runs in order `TRACE → LOG → DMA → CDN → LMT → ROM →
   STC → RUL → WAF → ROU` (see [03-edge.md](03-edge.md)).
6. `ROU` finds the location by longest prefix, rewrites the path, sets
   headers and reverse-proxies to the upstream.

### Admin operation (control path)

1. The console calls REST `http://<apiserver>:8080/...` with the header
   `Authorization: Bearer <JWT>`.
2. The grpc-gateway `runtime.ServeMux` turns it into an **in-process** gRPC
   call (`RegisterXxxHandlerServer`) to the control-plane handler.
3. For some RPCs the handler checks permissions (`headers.GetAuth` +
   `perms.Allow`), then calls service → repository → PostgreSQL.

There is currently **no channel that pushes configuration from the control
plane to the edge**: the edge reads `proxy.yaml` at startup; tenant/SecRule
data in the DB only serves the console. See
[12-known-issues.md](12-known-issues.md).

## 1.6 Main technologies

| Area | Library |
|---|---|
| DI / lifecycle | `go.uber.org/fx` |
| Edge HTTP | `cloudwego/hertz` (+ `hertz-contrib/http2`, `reverseproxy`), with a `net/http` alternative |
| REST ↔ gRPC | `grpc-ecosystem/grpc-gateway/v2` |
| WAF | `corazawaf/coraza/v3` + OWASP CRS v4.16/v4.17 |
| TLS fingerprint | `exaring/ja4plus` |
| Distributed state | `olric-data/olric` (embedded, memberlist) |
| eBPF | `cilium/ebpf` (`bpf2go`) |
| DB | `jackc/pgx/v5` + `Masterminds/squirrel` |
| Auth | `golang-jwt/jwt/v5` (HS256), `go-webauthn/webauthn` (passkey), bcrypt |
| Discovery | HashiCorp Consul + `sony/gobreaker` |
| Logging | `zap` + OpenTelemetry bridge (OTLP/gRPC) |
| Validation | `buf.build/go/protovalidate` |
| Error pages | `a-h/templ` |
| UI | Next.js 16, React 19, Tailwind 4, `ts-proto` |
