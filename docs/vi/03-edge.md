# 3. Edge proxy (`szedge`)

Edge là thành phần chịu tải chính: nhận HTTPS từ Internet, áp các lớp bảo vệ,
rồi reverse-proxy tới origin của tenant.

Mã nguồn liên quan:

| Đường dẫn | Vai trò |
|---|---|
| [cmd/szedge/v1/main.go](../../cmd/szedge/v1/main.go) | Entry point, bật pprof `:6060` |
| [pkg/apps/dmz/edge/](../../pkg/apps/dmz/edge/) | `Server` (Start/Shutdown), config, flags, nạp YAML |
| [internal/dmz/edge/engine/](../../internal/dmz/edge/engine/engine.go) | Chọn HTTP engine: `Hertz` (mặc định) hoặc `Standard` (`net/http`) |
| [internal/funcs/](../../internal/funcs/) | Các node của chain middleware |
| [internal/dmz/edge/transport/](../../internal/dmz/edge/transport/) | Hook TLS (JA4) và OnConnect |
| [internal/memory/](../../internal/memory/) | Bộ nhớ runtime theo namespace (route, proxy, rule, limiter, WAF, CDN) |
| [internal/cluster/](../../internal/cluster/) | Olric embedded + `DMap[T]` generic |
| [contrib/httphz](../../staging/src/github.com/sentinez/contrib/httphz/) | Server Hertz + `corehttp.Context` + reverse proxy |
| [pkg/network/](../../pkg/network/) | Listener/Conn có ID, transport `net/http`, bản `stdhttpx` |

## 3.1 Khởi động

```
main
 ├─ engine.Hertz(c)            → Inject(httphz.NewServer); OnStart: SetOptions(WithOnConnect(OnHertzConnect))
 ├─ Inject(config.Config, edgeyaml.LoadSetting, edge.New)
 └─ OnStart(server.Start)

edge.New(conf, setting, server)
 ├─ corecmn.NormalizeEdgeSetting(setting)   // *Lite (YAML) → runtime proto
 └─ memory.NewMemStore(setting)             // singleton, lưu Setting vào map + DMap

Server.Start()
 ├─ initialize()
 │   ├─ mem.Start(conf)        → cluster.Start: olric.New + db.Start (goroutine)
 │   ├─ mem.LoadServer(core)   → với mỗi Setting: route, reverse proxy, CDN, SecRule, limiter, WAF
 │   └─ core.Handle(http.Init(conf, mem).Handle)
 ├─ network.Listen(SENZ_ADDRESS | 0.0.0.0:443)
 └─ core.ListenAndServe(addr,
        WithCertificate(cert, key),
        WithTLSConfig{GetConfigForClient: transport.TLSConfig, MinVersion: TLS1.3},
        WithListener(l), + options (OnConnect))
```

Server Hertz ([httphz/server.go](../../staging/src/github.com/sentinez/contrib/httphz/server.go))
bật ALPN, H2C, HTTP/2 (`AddProtocol("h2", ...)`), stream body, dùng transport
`net/std` fork từ Hertz. Mọi route đều đi vào `NoRoute(handler)` — tức một
handler duy nhất cho mọi path.

## 3.2 Fingerprint & context kết nối

1. `network.Listener.Accept` bọc `net.Conn` thành `*network.Conn{Id}` (8 byte
   ngẫu nhiên, hex).
2. `transport.TLSConfig(chi)` (callback `GetConfigForClient`) tính
   `ja4plus.JA4(chi)` và `ja4.Set(conn.Id, fp)` — cache in-memory TTL 24h.
3. `OnHertzConnect` / `OnStandardConnect` bóc lớp `TLSConn → tls.Conn →
   network.Conn`, đưa `netpb.Transport{ConnId, ServerName(SNI)}` vào
   `context`.
4. `httphz.NewContext` đọc transport từ context, gán
   `Request.Fingerprint = ja4.Get(connId)`.
5. `network.Conn.Close` xoá fingerprint khỏi cache.

## 3.3 `corehttp.Context`

Interface trừu tượng hoá request/response
([core/http/context.go](../../staging/src/github.com/sentinez/core/http/context.go)),
gồm `RequestContext` (đọc/ghi header, query, path, body, IP, JA4, ...) +
`ResponseWriter` + `X()` trả về `edgepb.ContextExtra` (chứa `namespace`).

Có ba cài đặt:

| Cài đặt | File | Dùng khi |
|---|---|---|
| `httphz.Context` | `contrib/httphz/request_context.go` | Edge (Hertz) |
| `stdhttpx.Context` | `pkg/network/httpx/std/context.go` | Edge chế độ `Standard`, realtime |
| `corehttpreq.RequestContext` | `core/http/request/context_value.go` | Đánh giá rule trên `httppb.Request` (gRPC `EvaluateRuleset`) |

Context được lấy từ `sync.Pool` và trả lại trong `Close()`.

`RequestIP()`: bản Hertz dùng `ClientIP()` của Hertz; bản `net/http` ưu tiên
`X-Forwarded-For` (phần tử đầu) → `X-Real-IP` → `RemoteAddr`.

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
| 3 | `CDN` | `cdn.Cache` | Nếu có CDN rule active cho namespace và expression khớp: tra cache theo key `method+scheme+host+path+query`; hit → trả response với `X-Cache: HIT`; miss → chạy tiếp và lưu response `< 400` (TTL 1h) |
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
        limit: 50
        timeout: 5s
  controller:
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
- `Match(ctx)`: lấy route theo namespace, tìm location đầu tiên là prefix của
  path, đặt path mới = `rewrite + (path − location)`, set header, trả về
  `proxyPass[0].server`.

Ví dụ:

| location | proxyRewrite | Request path | Path gửi đi |
|---|---|---|---|
| `/` | `/` | `/about` | `/about` |
| `/api` | `/v1` | `/api/users` | `/v1/users` |
| `/api` | `/` | `/api/users` | `//users` (ghép thẳng, không chuẩn hoá) |

Phép ghép là nối chuỗi thuần nên `proxyRewrite: /` với location khác `/` sinh
ra `//`. Xem [12-known-issues.md](12-known-issues.md).

Reverse proxy được tạo một lần cho mỗi `upstream.server` trong
`MemStore.LoadReverseProxy` qua `server.AcceptReverse(target)`:

- Hertz (`proxyhz.ReverseProxy`): 2 client (TLS / plain), director viết lại
  URI bằng `JoinURLPath` và set `Host` = host upstream. Request có
  `Upgrade: websocket` chuyển sang `WSReverseProxy`.
- `net/http` (`stdhttpx.ReverseProxy`): `httputil.NewSingleHostReverseProxy`
  với transport timeout 30s.

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
trong `edge.Service` (`pkg/apps/dmz/edge/edge_service.go`). Hiện **không được
khởi chạy** trong `main` của edge; `EvaluateRuleset` đang đánh giá một
`Expression` rỗng.

## 3.10 Chế độ `net/http`

Để thay Hertz bằng thư viện chuẩn, đổi `engine.Hertz(c)` thành
`engine.Standard(c)` trong `main`. Server `stdhttpx.Server` đặt timeout
Read/Write 15s, Idle 120s khi có TLS, và dùng `ConnContext` cho OnConnect.
