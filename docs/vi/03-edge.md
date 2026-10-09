# 3. Edge proxy (`szedge`)

Edge là thành phần chịu tải chính: nhận HTTPS từ Internet, áp các lớp bảo vệ,
rồi reverse-proxy tới origin của tenant.

Mã nguồn liên quan:

| Đường dẫn | Vai trò |
|---|---|
| [cmd/szedge/v1/main.go](../../cmd/szedge/v1/main.go) | Entry point, bật pprof `:6060` |
| [pkg/apps/dmz/edge/](../../pkg/apps/dmz/edge/) | `Server` (Start/Shutdown), config, flags, nạp YAML |
| [internal/dmz/edge/engine/](../../internal/dmz/edge/engine/engine.go) | Constructor HTTP engine, chọn qua `--engine`: `Quic` (mặc định, HTTP/3 + HTTP/1.1/2), `Hertz` hoặc `Standard` (`net/http`) |
| [internal/funcs/](../../internal/funcs/) | Các node của chain middleware |
| [internal/dmz/edge/transport/](../../internal/dmz/edge/transport/) | Hook TLS (JA4) và OnConnect |
| [internal/memory/](../../internal/memory/) | Bộ nhớ runtime theo namespace (route, proxy, rule, limiter, WAF, CDN) |
| [internal/cluster/](../../internal/cluster/) | Olric embedded + `DMap[T]` generic |
| [contrib/httphz](../../staging/src/github.com/sentinez/contrib/httphz/) | Server Hertz + `corehttp.Context` + reverse proxy |
| [pkg/network/](../../pkg/network/) | Listener TCP (`StdListener`/`StdConn`, có ID) và QUIC (`QuicListener`), transport `net/http` |
| [pkg/network/httpx/](../../pkg/network/httpx/) | `Context` + `ReverseProxy` `net/http` dùng chung; server ở `std/` (`stdhttpx`) và `quic/` (`quichttpx`) |

## 3.1 Khởi động

```
main: runner.Main(NewApp(edgeServer), NewApp(grpcServer))
 ├─ edgeServer
 │   ├─ Inject(config.Config, edgeyaml.LoadSetting, edge.NewServer)
 │   └─ Serve(server.Start)
 └─ grpcServer
     ├─ Inject(config.Config, edge.NewService)
     └─ Serve(service.Start)   // gRPC on defaults.EdgeAddress

edge.NewServer(conf, setting)
 ├─ corecmn.NormalizeEdgeSetting(setting)   // *Lite (YAML) → runtime proto
 ├─ memory.NewMemStore(setting)             // singleton, lưu Setting vào map + DMap
 └─ switch conf.Flag.Engine                 // --engine
     hertz → engine.Hertz(conf)            → httphz.NewServer + WithOnStdConnect(OnHertzConnect)
     quic  → engine.Quic(conf)             → quichttpx.NewServer + WithOnStdConnect(OnStandardConnect)
     std, khác → engine.Standard(conf)     → stdhttpx.NewServer + WithOnStdConnect(OnStandardConnect)

Server.Start()
 ├─ initialize()
 │   ├─ mem.Start(conf)        → cluster.Start: olric.New + db.Start (goroutine)
 │   ├─ mem.LoadServer(server) → với mỗi Setting: route, reverse proxy, CDN, SecRule, limiter, WAF
 │   └─ server.Handle(http.Init(conf, mem).Handle)
 ├─ newTLSConfig(cert, key)  → {GetConfigForClient: transport.TLSConfig, MinVersion: TLS1.3, Certificates}
 ├─ network.StdListen(SENZ_ADDRESS | 0.0.0.0:443)   // TCP
 ├─ network.QuicListen(cùng addr, WithTLSConfig)    // UDP
 └─ server.ListenAndServe(addr,
        WithCertificate(cert, key), WithTLSConfig(tlsConf),
        WithStdListener(tcp), WithQuicListener(udp), + options (OnStdConnect))
```

Certificate được nạp vào `tls.Config` **trước** khi tạo listener: `QuicListen`
clone config (qua `http3.ConfigureTLSConfig`, hàm này cũng thêm ALPN `h3`),
nên certificate thêm vào sau sẽ không bao giờ được dùng. QUIC không có chế độ
plaintext, nên engine QUIC báo lỗi `quic: tls certificate required` khi không
cấu hình certificate.

### Engine QUIC (mặc định)

[pkg/network/httpx/quic/server.go](../../pkg/network/httpx/quic/server.go)
(`quichttpx.Server`) phục vụ trên cùng một địa chỉ:

- **HTTP/3 qua UDP** bằng `quic-go/http3` (`ServeListener`, idle timeout
  120s).
- **HTTP/1.1 và HTTP/2 qua TCP** bằng `net/http` (`ServeTLS`, idle 120s,
  read-header 10s). Mọi response TCP đều có header `Alt-Svc`
  (`http3.Server.SetQUICHeaders`) để client chuyển sang HTTP/3 ở request sau
  — client luôn vào TCP trước, nên bắt buộc phải phục vụ cả hai listener.

Hai server chạy trong một `errgroup`; khi một bên dừng thì bên kia bị đóng.
`Shutdown` tắt cả hai. Request ở cả hai giao thức đi qua cùng `http.ServeMux`
và cùng chain middleware, dùng `httpx.Context` chung. QUIC chỉ được terminate
ở edge: `AcceptReverse` dùng `network.StandardTransporter()`, nên upstream
được gọi qua TCP (HTTP/1.1 hoặc HTTP/2). `network.QuicTransporter` (một
`http3.Transport`) có sẵn nhưng chưa được dùng.

### Engine Hertz

Server Hertz ([httphz/server.go](../../staging/src/github.com/sentinez/contrib/httphz/server.go))
bật ALPN, H2C, HTTP/2 (`AddProtocol("h2", ...)`), stream body, dùng transport
`net/std` fork từ Hertz. Mọi route đều đi vào `NoRoute(handler)` — tức một
handler duy nhất cho mọi path. Chỉ dùng listener TCP (`opt.StdListener`).

## 3.2 Fingerprint & context kết nối

1. `network.StdListener.Accept` bọc `net.Conn` thành
   `*network.StdConn{Id}` (8 byte ngẫu nhiên, hex).
2. `transport.TLSConfig(chi)` (callback `GetConfigForClient`) tính
   `ja4plus.JA4(chi)` và `ja4.Set(conn.Id, fp)` — cache in-memory TTL 24h.
3. `OnHertzConnect` / `OnStandardConnect` (đăng ký qua `WithOnStdConnect`)
   bóc lớp `TLSConn → tls.Conn → network.StdConn`, đưa `netpb.Transport{ConnId, ServerName(SNI)}` vào
   `context`.
4. `httphz.NewContext` đọc transport từ context, gán
   `Request.Fingerprint = ja4.Get(connId)`.
5. `network.StdConn.Close` xoá fingerprint khỏi cache.

Luồng này chỉ áp dụng cho kết nối TCP. Với HTTP/3, `chi.Conn` không phải
`*network.StdConn` và engine QUIC không đặt hook `OnQuicConnect`, nên request
HTTP/3 không có JA4 fingerprint lẫn `netpb.Transport` (xem
[12-known-issues.md](12-known-issues.md)).

## 3.3 `corehttp.Context`

Interface trừu tượng hoá request/response
([core/http/context.go](../../staging/src/github.com/sentinez/core/http/context.go)),
gồm `RequestContext` (đọc/ghi header, query, path, body, IP, JA4, ...) +
`ResponseWriter` + `X()` trả về `edgepb.ContextExtra` (chứa `namespace`).

Có ba cài đặt:

| Cài đặt | File | Dùng khi |
|---|---|---|
| `httphz.Context` | `contrib/httphz/request_context.go` | Edge (Hertz) |
| `httpx.Context` | `pkg/network/httpx/context.go` | Edge chế độ `Quic` (HTTP/3 và TCP) và `Standard`, realtime |
| `corehttpreq.RequestContext` | `core/http/request/context_value.go` | Đánh giá rule trên `httppb.Request` (gRPC `EvaluateRuleset`) |

Context được lấy từ `sync.Pool` và trả lại trong `Close()`.

`RequestIP()`: bản Hertz dùng `ClientIP()` của Hertz; bản `net/http`
(`httpx.Context`) ưu tiên `X-Forwarded-For` (phần tử đầu) → `X-Real-IP` →
`RemoteAddr`.

`httpx.Context.ResponseBody()` trả về đúng những gì reverse proxy nhận từ
upstream: `httpx.ReverseProxy` bọc `ResponseWriter` bằng `recordWriter` để ghi
lại status code và sao chép body (tối đa 4 MiB; body lớn hơn vẫn được stream
qua nhưng không được giữ lại, khi đó `ResponseBody()` trả về `nil`). Context
trong pool bỏ buffer body lớn hơn 64 KiB. `SetPath` sao chép path vì caller
có thể truyền buffer lấy từ pool.

## 3.4 Chain middleware

Định nghĩa trong
[internal/dmz/edge/http/init.go](../../internal/dmz/edge/http/init.go). Mỗi node
nhúng `*corechains.Node` (giữ `next`) và cài `Handle(ctx) error`. Node được bọc
bởi `trace.Wrap(name, node)` để gắn tên stage vào lỗi
(`tracer.StageError{Stage, Elapsed, Err}`) và log một lần duy nhất ở stage
gây lỗi.

| Thứ tự | Stage | Node | Hành vi |
|---|---|---|---|
| 0 | — | `trace.Trace` | Sinh request ID `senz:req:<xid>`, set header `X-Request-Id` |
| 1 | `LOG` | `logging.Logging` | Gọi `next` trước, sau đó log `httppb.Event` (JSON, kind `LOG_TYPE_HTTP`). Có tra bandwidth eBPF theo IP |
| 2 | `DMA` | `secure.DomainBased` | Host phải có dạng `<ns>.<SENZ_HOSTNAME>` (một cấp, bỏ port). Sai → `403`. Gán `X().Namespace = ns` |
| 3 | `CDN` | `cdn.Cache` | Nếu có CDN rule active cho namespace và expression khớp: tra cache theo key `method+scheme+host+path+query+Accept-Encoding`; hit → trả response với `X-Cache: HIT`; miss → chạy tiếp và lưu response cache được (TTL 1h, xem [04-rule-engine.md](04-rule-engine.md#44-cdn-rules)) |
| 4 | `LMT` | `ratelimiter.Limiter` | `limiter.Allow(IP)`; vượt → `429` |
| 5 | `ROM` | `room.WaitingRoom` | Placeholder, chỉ chuyển tiếp |
| 6 | `STC` | `static.Static` | Sau khi chạy tiếp, nếu path có đuôi tĩnh (`.css .js .png ...`) thì set `Cache-Control: public, max-age=3600, immutable` |
| 7 | `RUL` | `secure.SecRule` | Duyệt các SecRule active theo priority giảm dần; rule đầu tiên khớp có action `BLOCK` → `403`; action khác chỉ log debug và đi tiếp |
| 8 | `WAF` | `secure.WAF` | Coraza transaction: `ExecIngress` (header/body) → chạy tiếp → `ExecEgress` (response). Bị chặn với status 403 → trang 403. Cuối cùng log `secrulepb.Event` (`LOG_TYPE_WAF`) nếu bị interrupt |
| 9 | `ROU` | `routing.StandardRouter` | Tìm route, rewrite path, set header, gọi `ReverseProxy.Serve` |

Response lỗi (`400/403/404/429/500`) được render bằng template `templ`
([core/common/render](../../staging/src/github.com/sentinez/core/common/render/))
kèm `X-Request-ID`; nếu render lỗi thì fallback text thuần.

## 3.5 Cấu hình `proxy.yaml`

File YAML có khoá gốc `setting`, được chuyển sang JSON rồi
`protojson.Unmarshal` vào `edgepb.Setting`
([pkg/apps/dmz/edge/yaml/config.go](../../pkg/apps/dmz/edge/yaml/config.go)).
Vì vậy tên trường theo **camelCase JSON của proto** và enum viết dạng chuỗi.
Ví dụ đầy đủ: [cmd/szedge/v1/proxy.yaml](../../cmd/szedge/v1/proxy.yaml).

```yaml
setting:
  metadata: {}
  server:
    name: badcheese              # namespace = subdomain
    locations:
      - location: /api           # prefix khớp
        proxyRewrite: /          # thay prefix trước khi forward (mặc định = location)
        proxyPass:
          - server: api.example.com
            protocol: PROXY_PROTOCOL_HTTPS   # hoặc PROXY_PROTOCOL_HTTP
        balanceStrategy: BALANCE_STRATEGY_ROUND_ROBIN
        proxySetHeaders:
          X-Real-IP: $remote_addr
  security:
    rulesets:                    # có phần tử → bật CRS cho namespace
    rules:                       # SecRule (dạng Lite)
      - ingress:
          name: "Block /block"
          status: STATUS_ACTIVE
          expr:
            orCondition:
              - rules:
                  - condition:
                      source: FIELD_SOURCE_PATH
                      operator: OPERATOR_EQ
                      value: "/block"
          action:
            type: ACTION_TYPE_BLOCK
    limiters:
      - timeWindow: 5s
        maxRequests: 50
        timeout: 5s
  traffic_control: {}
  delivery:
    cdn:
      - rule:
          name: cache GET
          status: STATUS_ACTIVE
          expr: { orCondition: [ { rules: [ { condition: { source: FIELD_SOURCE_METHOD, operator: OPERATOR_EQ, value: GET } } ] } ] }
  personal: {}
```

Ghi chú:

- Phần "Lite" (`SecRuleLite`, `cdn.RuleLite`, `ExpressionLite`) dùng chuỗi
  cho `source/operator/status/value`. `NormalizeEdgeSetting` chuyển sang bản
  runtime (`SecRule`, `cdn.Rule`, `Expression`) với enum thật, `value` thành
  `structpb.Value` dạng string, và sinh ID `senz.rule.<nanoid>` /
  `senz.cond.<nanoid>`.
- Biến hỗ trợ trong `proxySetHeaders`
  ([core/http/variable](../../staging/src/github.com/sentinez/core/http/variable/variable.go)):
  `$host`, `$remote_addr`, `$scheme`, `$request_uri`,
  `$proxy_add_x_forwarded_for`. Chỉ hỗ trợ khi **toàn bộ giá trị** là một biến;
  chuỗi khác chứa `$` gây lỗi khi route. Header key không hợp lệ (RFC 7230) bị
  bỏ qua với cảnh báo.
- Edge luôn set thêm `X-Forwarded-Prefix: <location>`.

## 3.6 Routing

[internal/memory/routes/routes.go](../../internal/memory/routes/routes.go):

- `Store(server)`: sao chép location, lọc header hợp lệ, đặt `proxyRewrite`
  mặc định, **sắp xếp giảm dần theo độ dài `location`** để khớp
  longest-prefix như nginx.
- `Match(ctx)`: lấy route theo namespace, tìm location đầu tiên khớp path
  **theo ranh giới segment** (`/api` khớp `/api` và `/api/x` nhưng không khớp
  `/apix`; location kết thúc bằng `/` khớp mọi path bên dưới), đặt path mới =
  `rewrite + (path − location)` mà không lặp `/` ở chỗ nối, set header, trả về
  `proxyPass[0].server`.

Ví dụ:

| location | proxyRewrite | Request path | Path gửi đi |
|---|---|---|---|
| `/` | `/` | `/about` | `/about` |
| `/api` | `/v1` | `/api/users` | `/v1/users` |
| `/api` | `/` | `/api/users` | `/users` |
| `/api` | `/` | `/api` | `/` |
| `/api` | `/v1/` | `/api/users` | `/v1/users` |
| `/api` | — | `/apix` | không khớp `/api` |

Reverse proxy được tạo một lần cho mỗi `upstream.server` trong
`MemStore.LoadReverseProxy` qua `server.AcceptReverse(target)`:

- Hertz (`proxyhz.ReverseProxy`): 2 client (TLS / plain), director viết lại
  URI bằng `JoinURLPath` và set `Host` = host upstream. Request có
  `Upgrade: websocket` chuyển sang `WSReverseProxy`.
- `net/http` (`httpx.ReverseProxy`, dùng cho engine `Quic` và `Standard`):
  `httputil.NewSingleHostReverseProxy` với `http.RoundTripper` truyền vào
  (`network.StandardTransporter()`, dial timeout 30s). Writer được bọc để
  context thấy status và body của upstream; `Unwrap` giữ `Flush`/`Hijack`
  cho streaming và upgrade.

## 3.7 Memory store

`memory.MemStore` (singleton) gom các kho theo **namespace** (`server.name`):

| Kho | Kiểu | Nội dung |
|---|---|---|
| `settings.Setting` | `Map[string,*edgepb.Setting]` + `cluster.DMap` | Setting gốc; ghi vào Olric DMap `sentinez.dmz.edge.setting` |
| `routes.Router` | `Map[string,[]*Location]` | Location đã sắp xếp |
| `reverseproxy.ReverseProxy` | `Map[string,corehttp.ReverseProxy]` | Key = upstream server |
| `secrules.SecRule` | `Map[string,[]Entry]` | `Entry{Eval, Rule}` đã compile, lọc active, sort theo priority |
| `ratelimiter.Limiter` | `Map[string,*RateLimiter]` | Một limiter / namespace |
| `rulesets.Rulesets` | `Map[string,coraza.WAF]` | WAF instance / namespace |
| `cdnrules.Rule` | `Map[string,EvalFunc]` + `Map[string,*cdn.CDN]` | Một CDN rule / namespace |

Mọi map dùng `shared/sync.Map[K,V]` (bọc `sync.Map` có generic).

## 3.8 Cluster (Olric)

[internal/cluster](../../internal/cluster/):

- `NewCluster` tạo Olric với config `local`. Nếu có cả
  `SENZ_MEMBERSHIP_ADDRESS` và `SENZ_DISCOVERY_ADDRESS` thì đặt `Peers` và bind
  memberlist theo discovery address.
- `DMap[T]` là wrapper generic lưu proto đã `proto.Marshal`. Khi cluster chưa
  sẵn sàng, `Put` đẩy vào channel đệm 1024 phần tử; goroutine `wait()` đợi
  `dictReady` rồi tạo DMap và xả hàng đợi.
- Hiện chỉ dùng để phát tán `Setting`; chưa có chỗ đọc lại (`Get`) từ node khác.

## 3.9 gRPC EdgeService

`edgepb.EdgeService` có `Status` và `EvaluateRuleset`
([internal/dmz/edge/api/api.go](../../internal/dmz/edge/api/api.go)), đóng gói
trong `edge.Service` (`pkg/apps/dmz/edge/edge_service.go`). `main` khởi chạy nó
qua app `grpcServer` trên `defaults.EdgeAddress`; `EvaluateRuleset` đang đánh
giá một `Expression` rỗng.

## 3.10 Chọn engine

Engine được chọn lúc khởi động bằng flag `--engine` (mặc định `quic`);
`edge.NewServer` switch theo `conf.Flag.Engine`:

| `--engine` | Constructor | Server | Giao thức |
|---|---|---|---|
| `quic` (mặc định) | `engine.Quic` | `quichttpx.Server` | HTTP/3 (UDP) + HTTP/1.1/2 (TCP), bắt buộc có certificate |
| `hertz` | `engine.Hertz` | `httphz.XServer` | HTTP/1.1/2 (chỉ TCP) |
| `std` | `engine.Standard` | `stdhttpx.Server` (`net/http`) | HTTP/1.1/2 (chỉ TCP) |

```sh
./cmd/szedge/v1/bin/main --engine=hertz ...
```

Giá trị ngoài `std|hertz|quic` bị protovalidate từ chối khi parse flag; giá
trị rỗng sẽ fallback về `Standard`. Log khởi động hiển thị engine đang dùng
(`https[quic] running on 0.0.0.0:7443`). Server `stdhttpx.Server`
đặt timeout Read/Write 15s, Idle 120s khi có TLS, và dùng `ConnContext` cho
`OnStdConnect`. `Server.Start` của edge luôn mở cả listener TCP lẫn UDP, bất
kể dùng engine nào.
