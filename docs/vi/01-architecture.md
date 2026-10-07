# 1. Tổng quan kiến trúc

## 1.1 Sentinéz là gì

Sentinéz là một reverse proxy bảo mật đặt trước origin server của khách hàng
(tenant). Mỗi request đi qua edge sẽ được gắn request ID, xác định tenant theo
subdomain, kiểm tra cache, rate limit, rule tuỳ biến (SecRule), bộ OWASP CRS
(Coraza WAF), rồi mới được chuyển tới origin. Phía sau là các service quản trị
(control plane) cung cấp REST API cho web console.

Phiên bản hiện tại: `v0.0.1-beta`
([core/core.go](../../staging/src/github.com/sentinez/core/core.go)). README gốc
cảnh báo chưa đảm bảo tương thích ngược trước v1.0.0.

## 1.2 Sơ đồ thành phần

```
                    Internet
                       │
             ┌─────────▼─────────┐
             │  szdataplane (XDP)│  eBPF: đếm bandwidth / chặn IP (netns "gateway")
             └─────────┬─────────┘
                       │
             ┌─────────▼─────────┐        Olric cluster (memberlist)
             │   szedge (HTTPS)  │◄──────► chia sẻ edge Setting giữa các node
             │  Hertz + chain    │
             └─────────┬─────────┘
                       │ reverse proxy
                       ▼
                 Origin của tenant


  Web console (Next.js) ──REST──► szapiserver (grpc-gateway, :8080)
                                     │ in-process gRPC handler
                                     ├─ IAM        ─┐
                                     ├─ Tenant      │  PostgreSQL
                                     ├─ Security    ├─►(bảng <env>_sentinez_*)
                                     └─ Greeter    ─┘
                    szrealtime (WebSocket /ws) — broadcast tin nhắn
                    Consul — service discovery (tuỳ chọn)
```

## 1.3 Các binary (`cmd/`)

| Binary | Entry point | Service type (`runner.App[T]`) | Vai trò |
|---|---|---|---|
| `szapiserver` | [cmd/szapiserver/main.go](../../cmd/szapiserver/main.go) | `apiserver.Server` | HTTP gateway REST → gRPC handler của control plane, phục vụ Swagger UI |
| `szrealtime` | [cmd/szrealtime/main.go](../../cmd/szrealtime/main.go) | `realtime.Realtime` | WebSocket server, broadcast tin nhắn giữa các client |
| `szedge/v1` | [cmd/szedge/v1/main.go](../../cmd/szedge/v1/main.go) | `edge.Server` | Edge proxy HTTPS (TLS 1.3, HTTP/2) chạy chain bảo mật |
| `szdataplane/v1` | [cmd/szdataplane/v1/main.go](../../cmd/szdataplane/v1/main.go) | `dataplane.Server` | Gắn chương trình XDP vào `veth0`, mở gRPC server |
| `szgreeter/v1` | [cmd/szgreeter/v1/main.go](../../cmd/szgreeter/v1/main.go) | `greeter.Greeter` | Service gRPC mẫu (SayHello/Status) |
| `szcentraldata/v1` | [cmd/szcentraldata/v1/main.go](../../cmd/szcentraldata/v1/main.go) | — | `main()` rỗng, chưa triển khai |

Mọi binary (trừ `szcentraldata`) đều theo một khuôn:

```go
func main() {
    runner.New(config.Config(), core.Code).Main(runner.NewApp(app))
}

func app(c *runner.Context[T]) {
    c.Inject(config.Config, <constructor>...)   // fx.Provide
    c.Serve(func(ctx, *T) error { ... })        // vòng lặp chính, block
    c.OnStop(func(ctx, *T) error { ... })
}
```

Một binary có thể chạy nhiều app bằng cách truyền nhiều `runner.NewApp(...)`
vào `Main`; `szedge` chạy HTTPS proxy và gRPC API của nó theo cách này.

Chi tiết `runner` ở [09-core-shared.md](09-core-shared.md#91-corerunner).

## 1.4 Cấu trúc monorepo

Repo là **multi-module** dùng `replace` trong [go.mod](../../go.mod) gốc
(không có `go.work`):

| Module | Đường dẫn | Phụ thuộc vào |
|---|---|---|
| `github.com/sentinez/sentinez` (root) | `/` | tất cả module bên dưới |
| `github.com/sentinez/sentinez/api` | `api/` | — (proto + client) |
| `github.com/sentinez/core` | `staging/.../core` | `api`, `shared` |
| `github.com/sentinez/shared` | `staging/.../shared` | `api`, `core` (một số gói) |
| `github.com/sentinez/controlplane` | `staging/.../controlplane` | `api`, `core`, `shared` |
| `github.com/sentinez/contrib/httphz` | `staging/.../contrib/httphz` | `core`, `shared`, Hertz |
| `github.com/sentinez/bpf` | `staging/.../bpf` | cilium/ebpf |
| `github.com/sentinez/tools` | `staging/.../tools` | `core`, `api` |

Các module trong `staging/` được đẩy sang repo riêng
`github.com/sentinez/<tên>` bằng `git subtree split` (xem
[11-deploy-ci.md](11-deploy-ci.md#114-phát-hành-module-staging)).

Thư mục cấp root:

| Thư mục | Nội dung |
|---|---|
| `cmd/` | Entry point từng service, Dockerfile, `.env.example`, `proxy.yaml` |
| `internal/` | Code riêng của root module: `dmz/edge`, `dmz/dataplane`, `memory`, `cluster`, `bpf`, `defaults` |
| `pkg/` | Wiring service (`apps/...`), mạng (`network`), object pool (`pools`), `protocol`, `tracer`, `queue` |
| `api/` | Proto, code sinh ra, client gRPC, OpenAPI/Swagger nhúng |
| `staging/` | Các module phát hành riêng |
| `ui/` | Workspace pnpm + turbo (console Next.js, web, package dùng chung) |
| `deploy/` | Docker compose, Caddy, nginx, Prometheus/Grafana, rule CRS (`ruleroot`) |
| `hack/` | Script dev: lint, sinh rule, netns, cert, sync GitLab |
| `_submodules/` | coreruleset, googleapis, grpc-gateway, opentelemetry-proto, protovalidate |
| [sentinez.go](../../sentinez.go) | Nhúng file `*.data` của CRS v4.16.0 / v4.17.0 qua `embed.FS` |

## 1.5 Luồng dữ liệu chính

### Request của người dùng cuối (data path)

1. TCP accept tại `network.Listener` → bọc thành `network.Conn` có `Id` ngẫu
   nhiên 8 byte (hex).
2. TLS ClientHello → `transport.TLSConfig` tính JA4 fingerprint và lưu vào
   `shared/store/ja4` theo `conn.Id`.
3. `OnConnect` gắn `netpb.Transport{ConnId, ServerName}` vào `context`.
4. Hertz gọi handler → tạo `httphz.Context` (lấy từ pool), gán fingerprint.
5. Chain middleware chạy theo thứ tự `TRACE → LOG → DMA → CDN → LMT → ROM →
   STC → RUL → WAF → ROU` (xem [03-edge.md](03-edge.md)).
6. `ROU` tìm location theo longest-prefix, rewrite path, set header, gọi
   reverse proxy tới upstream.

### Thao tác quản trị (control path)

1. Console gọi REST `http://<apiserver>:8080/...` với header
   `Authorization: Bearer <JWT>`.
2. grpc-gateway `runtime.ServeMux` chuyển thành lời gọi gRPC **in-process**
   (`RegisterXxxHandlerServer`) tới handler của controlplane.
3. Handler kiểm tra quyền (`headers.GetAuth` + `perms.Allow`) với một số RPC,
   rồi gọi service → repository → PostgreSQL.

Hiện tại **chưa có kênh đẩy cấu hình từ control plane xuống edge**: edge đọc
`proxy.yaml` lúc khởi động; dữ liệu tenant/SecRule trong DB chỉ phục vụ
console. Xem [12-known-issues.md](12-known-issues.md).

## 1.6 Công nghệ chính

| Lĩnh vực | Thư viện |
|---|---|
| DI / lifecycle | `go.uber.org/fx` |
| HTTP edge | `cloudwego/hertz` (+ `hertz-contrib/http2`, `reverseproxy`), có bản `net/http` thay thế |
| REST ↔ gRPC | `grpc-ecosystem/grpc-gateway/v2` |
| WAF | `corazawaf/coraza/v3` + OWASP CRS v4.16/v4.17 |
| Fingerprint TLS | `exaring/ja4plus` |
| Phân tán state | `olric-data/olric` (embedded, memberlist) |
| eBPF | `cilium/ebpf` (`bpf2go`) |
| DB | `jackc/pgx/v5` + `Masterminds/squirrel` |
| Auth | `golang-jwt/jwt/v5` (HS256), `go-webauthn/webauthn` (passkey), bcrypt |
| Discovery | HashiCorp Consul + `sony/gobreaker` |
| Logging | `zap` + bridge OpenTelemetry (OTLP/gRPC) |
| Validation | `buf.build/go/protovalidate` |
| Template lỗi | `a-h/templ` |
| UI | Next.js 16, React 19, Tailwind 4, `ts-proto` |
