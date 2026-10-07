# 9. Framework & thư viện dùng chung

## 9.1 `core/runner`

[core/runner](../../staging/src/github.com/sentinez/core/runner/) bọc
`uber/fx` để chuẩn hoá vòng đời service và chạy nhiều app trong cùng một
binary.

| API | Mô tả |
|---|---|
| `New(conf, scope)` | Tạo console logger cho gRPC, đặt log level theo `--log_level`, cài OTLP exporter; trả về `*Runner` |
| `Runner.Run(ctx, apps...)` | Khởi động các app theo thứ tự, chạy đến khi `ctx` kết thúc, nhận `SIGINT/SIGTERM` hoặc một `Serve` lỗi, rồi dừng chúng theo thứ tự ngược; trả về các lỗi đã gộp |
| `Runner.Main(apps...)` | `Run` với `context.Background()`; lỗi → `zlog.Fatal` |
| `NewApp[T](setup...)` | Tạo `*App`; `T` được suy ra từ `setup func(*Context[T])` |
| `Context.Inject(fns...)` | `fx.Provide` |
| `Context.Invoke(fn)` | `fx.Invoke` |
| `Context.OnStart(func(ctx, *T) error)` | Hook khởi động; chạy đồng bộ theo thứ tự đăng ký và không được block |
| `Context.Serve(func(ctx, *T) error)` | Vòng lặp chính (block), chạy sau khi mọi app đã khởi động; `http.ErrServerClosed` được coi là thoát bình thường, lỗi khác dừng toàn bộ runner |
| `Context.OnStop(func(ctx, *T) error)` | Hook dừng; chạy theo thứ tự đăng ký ngược |

Mỗi app có fx container riêng, nên các app trong cùng process có thể cùng
provide một kiểu (ví dụ `config.Config`). Hook dừng có timeout 15 s. Ở mode
khác `dev`, log của fx bị tắt (`fx.NopLogger`).

OTLP: endpoint `SENZ_OTLP_ENDPOINT` (mặc định `localhost:4317`), service name
= `meta.service_key`. Exporter được shutdown sau khi mọi app đã dừng.

## 9.2 `core/http` và `core/http/chains`

- `corehttp.Server`: `ListenAndServe`, `Shutdown`, `Use(middleware)`,
  `Handle(RequestHandler)`, `AcceptReverse(target)`.
- `ServerOption`: `WithCertificate`, `WithTLSConfig`, `WithServerName`,
  `WithOnConnect`, `WithListener`.
- `DecoreServer(conf, s)`: decorator in banner ASCII (`console.INFO`) trước
  khi listen.
- Helper trả lỗi: `Forbidden`, `BadRequest`, `NotFound`, `TooManyRequests`,
  `InternalServerError` (render templ).
- `corechains.ChainNode{SetNext, Handle}` + `Node{next}` với `HandleNext`,
  `GetNext` — nền cho chain middleware edge.
- `core/http/variable`: biến kiểu nginx cho header proxy.
- `core/http/const`, `core/common/bytestr`: hằng header/giá trị (dạng string
  và `[]byte`).

## 9.3 `core/grpc`

- `coregrpc.New(opts...)`: `WithXMeta`, `WithDiscorvery(bool)`,
  `WithGRPCServerOption(...)`. `Serve(conf)` listen TCP `SENZ_ADDRESS`, in
  banner, tuỳ chọn đăng ký Consul. `BufServe(bufLis)` phục vụ qua bufconn.
  `Shutdown` = `GracefulStop`.
- `core/grpc/gateway`: xem [06-gateway.md](06-gateway.md).

## 9.4 Các gói `core` khác

| Gói | Nội dung |
|---|---|
| `core` | Hằng `Name` (`sentinez/1.0`, dùng làm header `Server`), `Code` (`SENZ`), `Version`, prefix request ID, `EnvPattern` |
| `core/context` | `WithTransportValue` / `GetTransport` cho `netpb.Transport` |
| `core/common` | `NormalizeEdgeSetting` (Lite → runtime) |
| `core/common/console` | Banner ASCII (`go-figure`) in một lần khi service chạy |
| `core/common/render` | Template `templ` cho trang 400/403/404/429/500 |
| `core/storage/cache/mem` | Cache generic dựa trên `patrickmn/go-cache` |
| `core/storage/dbx` | Xem [07-controlplane.md](07-controlplane.md#71-lớp-lưu-trữ-corestoragedbx) |
| `core/storage/{cache/keyval,dbx/clickhouse,msgx/kafka,search/elastic}` | Package rỗng (dự kiến) |
| `core/limiter`, `core/rules`, `core/rulesets`, `core/modsec` | Xem [04-rule-engine.md](04-rule-engine.md) |

## 9.5 `shared`

| Gói | Nội dung |
|---|---|
| `zlog` | Logging (chi tiết dưới) |
| `config` | `LoadEnv(file, keys...)` nạp `.env` bằng godotenv rồi chụp giá trị `SENZ_*` vào map toàn cục; `GetEnv`/`SetEnv` có mutex |
| `flagx` | Flag chung `--mode`, `--log_level`; `Parse(meta)` (usage in banner); `Validate` bằng protovalidate |
| `errorx` | gRPC status chuẩn hoá — xem [07-controlplane.md](07-controlplane.md#77-mã-lỗi) |
| `perms` | `Allow(XMethod, ControlPlane)` |
| `sync` | `Map[K,V]` (generic `sync.Map`), `Pool[T]` (`NewPool`, `NewPoolCtr`) |
| `rand` | `RandomString` (crypto/rand), `NewID` (UUID), `NewNanoID`, `NewXID`, `NewTimeID` (ULID) — đều có prefix |
| `bytesconv` | Fork từ fasthttp/Hertz: `B2s`/`S2b` zero-copy, bảng tra ký tự, parse số |
| `jsonx` | `Marshal`/`Unmarshal` bằng `bytedance/sonic` |
| `protobuf` | Chuyển đổi Timestamp/Duration, `Compare` cho go-cmp, `Validate` |
| `protobuf/protox` | `Marshal`/`Unmarshal` nhị phân, `Struct(any)` → `structpb.Struct` |
| `copier` | Sao chép proto/JSON qua marshal |
| `cron` | `Start(ctx, interval, job)`: chạy job ngay rồi lặp theo ticker đến khi ctx huỷ |
| `eventq` | Pub/sub in-process theo namespace dựa trên channel (chưa có nơi dùng) |
| `store/ja4` | Cache fingerprint JA4 theo connection ID (TTL 24h) |
| `color` | Màu ANSI cho log console |
| `topic` | Package rỗng |

### `zlog`

Hai loại logger:

1. **Console (sugar)** — dùng qua hàm toàn cục `zlog.Info/Debugf/Errorf/
   Fatal...`. Encoder console có màu; level đặt bởi `SetScopeLogLevel`.
2. **Structured event** — `NewLog(name, kind, level)` / `NewLogCloser(...)`
   ghi JSON ra stdout với trường `kind` (`LogType`) và `event` (proto được
   marshal qua `protoreflect`). `LogCloser` gọi `Close()` trên object sau khi
   log (trả về pool).

Cả hai đều được tee sang bridge `otelzap`, nên khi `SetupOTLP` đã chạy, log
được xuất qua OTLP/gRPC (`NewOTLPProvider` + batch processor). Có thể gán
provider riêng cho một logger bằng `WithOTLPProvider(p)`.

Mức log: `debug=0, info=1, warn=2, error=3, fatal=4`; một bản ghi được ghi
khi `level >= verbosity`.

## 9.6 `pkg/` (root module)

| Gói | Nội dung |
|---|---|
| `pkg/apps/**/config` | `Config()` singleton: parse flag + nạp env + gắn `XMeta` |
| `pkg/apps/**/flags` | Flag riêng của service, giá trị mặc định `--env_file` |
| `pkg/network` | `Listen` (Conn có ID), `StandardTransporter` (dial timeout 30s), `GetInterface` |
| `pkg/network/httpx/std` | Cài đặt `corehttp.Server`/`Context`/`ReverseProxy` bằng `net/http` |
| `pkg/network/wsz` | WebSocket server + client manager |
| `pkg/pools/{request,ruleevent}` | Pool cho `httppb.Event`, `secrulepb.Event` |
| `pkg/protocol` | `Upstream2Target`: `PROXY_PROTOCOL_HTTP(S)` + server → URL |
| `pkg/tracer` | `StageError{Stage, Elapsed, Err}`, `Origin(err)` |
| `pkg/queue` | `Queue{}` rỗng (cho waiting room) |

## 9.7 `contrib/httphz`

Adapter Hertz cho `corehttp`:

- `NewServer(conf)`: `XServer` cài `corehttp.Server` bằng Hertz (TLS, ALPN,
  H2C, HTTP/2, stream body, custom transport).
- `Context`: cài `corehttp.Context` trên `*app.RequestContext`.
- `proxy`: reverse proxy HTTP(S) và WebSocket dựa trên
  `hertz-contrib/reverseproxy`, với option kiểu functional (`WithTLS`,
  `WithTimeout`, `WithStreamResponseBody`, ...).
- `net/std`: transport/connection fork từ Hertz (license CloudWeGo) để lộ
  `TLSConn` cho việc lấy `network.Conn` gốc.

## 9.8 `tools`

| Công cụ | Mô tả |
|---|---|
| `protoc-gen-go-senz` | Plugin protoc xử lý `x_meta`, `x_message`, `x_method` |
| `ruleparser-sentinez` | `-file <crs.conf> -out <dir>` → file Go chứa rule CRS mã hoá base64 |
| [tools.go](../../staging/src/github.com/sentinez/tools/tools.go) | Ghim phiên bản các công cụ: templ, buf, bpf2go, bombardier, protoc-go-inject-tag, grpc-gateway, openapiv2, vtproto, sqlc, mockery, protoc-gen-go(-grpc) |
