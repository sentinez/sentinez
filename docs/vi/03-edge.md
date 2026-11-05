# 3. Edge proxy (`szedge`)

Edge là thành phần chịu tải chính: nhận HTTPS từ Internet, áp các lớp bảo vệ,
rồi reverse-proxy tới origin của tenant.

Mã nguồn liên quan:

| Đường dẫn | Vai trò |
|---|---|
| [cmd/szedge/v1/main.go](../../cmd/szedge/v1/main.go) | Entry point, bật pprof `:6060` |
| [pkg/apps/dmz/edge/](../../pkg/apps/dmz/edge/) | `Server` (Start/Shutdown), config, flags, nạp YAML |
| [internal/dmz/edge/engine/](../../internal/dmz/edge/engine/engine.go) | Constructor HTTP engine, chọn qua `--engine`: `Quic` (mặc định, HTTP/3 + HTTP/1.1/2), `Hertz` hoặc `Standard` (`net/http`) |
| [internal/handlers/](../../internal/handlers/) | Các plugin của chain middleware |
| [internal/dmz/edge/transport/](../../internal/dmz/edge/transport/) | Hook TLS (JA4) và OnConnect |
| [internal/memory/](../../internal/memory/) | Bộ nhớ runtime theo namespace (route, proxy, rule, limiter, WAF, CDN) |
| [pkg/cluster/](../../pkg/cluster/) | Olric embedded + `SyncMap[T]` (memory đồng bộ giữa các edge) |
| [pkg/cluster/leader/](../../pkg/cluster/leader/) | Bầu leader theo zone (session lock của Consul) |
| [contrib/httphz](../../staging/src/github.com/sentinez/contrib/httphz/) | Server Hertz + `szhttp.Context` + reverse proxy |
| [pkg/network/](../../pkg/network/) | Listener TCP (`StdListener`/`StdConn`, có ID) và QUIC (`QuicListener`), transport `net/http` |
| [pkg/network/httpx/](../../pkg/network/httpx/) | `Context` + `ReverseProxy` `net/http` dùng chung; server ở `std/` (`stdhttpx`) và `quic/` (`quichttpx`) |

## 3.1 Khởi động

```
main: run() → runner.Run(ctx, NewApp(edgeServer), NewApp(grpcServer), NewApp(leader), NewApp(pprof))
 ├─ edgeServer
 │   ├─ Inject(config.Config, edgeyaml.LoadSetting, edge.NewServer)
 │   └─ Serve(server.Start)
 ├─ grpcServer
 │   ├─ Inject(config.Config, edge.NewService)
 │   └─ Serve(service.Start)   // gRPC on defaults.EdgeAddress
 └─ leader
     ├─ Inject(config.Config, leader.New)
     └─ Serve(e.Run(ctx, e.PublishStatus()))  // leader theo zone, xem 3.8

edge.NewServer(conf)
 ├─ edgeyaml.LoadSetting(conf)              // setting *Lite thô của proxy.yaml
 ├─ memory.NewMemStore(conf)                // singleton; cluster.NewCluster + SyncMap setting
 └─ switch conf.Flag.Engine                 // --engine
     hertz → engine.Hertz(conf)            → httphz.NewServer + WithOnStdConnect(OnHertzConnect)
     quic  → engine.Quic(conf)             → quichttpx.NewServer + WithOnStdConnect(OnStandardConnect)
     std, khác → engine.Standard(conf)     → stdhttpx.NewServer + WithOnStdConnect(OnStandardConnect)

Server.Start()
 ├─ initialize()
 │   ├─ mem.Sync(ctx, server, setting)  // 30s: đợi Olric, nạp setting của cluster,
 │   │                                  // chia sẻ setting này; mỗi thay đổi dựng lại route,
 │   │                                  // reverse proxy, CDN, SecRule, limiter, WAF
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

Mỗi kết nối TLS được cấp một ID trước khi handshake. `transport.TLSConfig`
(callback `GetConfigForClient`) lưu JA4 fingerprint theo ID đó vào
`shared/szstore/ja4` (in-memory, TTL 24h). Hook connect đưa cùng ID đó vào
context của request dưới dạng `netpb.Transport{ConnId, ServerName(SNI)}`, và
`httpx.NewContext` / `httphz.NewContext` gán
`Request.Fingerprint = ja4.Get(connId)`. TCP và QUIC chỉ khác nhau ở chỗ giữ
ID:

| | TCP (HTTP/1.1, HTTP/2) | QUIC (HTTP/3) |
|---|---|---|
| Tạo ID | `network.StdListener.Accept`, dạng `*network.StdConn{Id}` | `quic.Transport.ConnContext` trong `network.QuicListen`, gắn vào context của kết nối |
| `TLSConfig` đọc ID bằng `network.HelloConnId(chi)`, từ | `chi.Conn.(*network.StdConn).Id` | `network.ConnId(chi.Context())` |
| Hook connect | `OnStandardConnect` / `OnHertzConnect` (`WithOnStdConnect`) bóc lớp `TLSConn → tls.Conn → network.StdConn` | `OnQuicConnect` (`WithOnQuicConnect`) đọc `network.ConnId(conn.Context())` |
| Ký tự đầu JA4 | `t` | `q` |
| Xoá khỏi cache | `network.StdConn.Close` | `context.AfterFunc` khi quic-go huỷ context của kết nối (đóng kết nối hoặc handshake lỗi) |

ID là 8 byte ngẫu nhiên dạng hex (`network.newConnId`).

**Chi tiết QUIC**

- `QuicListen` tự tạo `quic.Transport` trên một UDP socket (thay cho
  `quic.ListenAddr`) để đặt được `ConnContext`. quic-go dẫn xuất context
  truyền cho `crypto/tls` (`tls.ClientHelloInfo.Context`) và context trả về
  từ `quic.Conn.Context` từ context mà `ConnContext` trả về, nên handshake và
  các request thấy cùng một ID. `QuicListener.Close` đóng transport (kéo theo
  listener) và UDP socket, vì quic-go không đóng socket mà nó không tự tạo.
- ID do server sinh, không phải QUIC Connection ID. quic-go chỉ đưa Destination
  Connection ID gốc (do client chọn) cho `Tracer` của qlog, mà callback này
  không sửa được context. Client cũng có thể cố tình dùng lại một DCID để ghi
  đè mục của kết nối khác. ID tự sinh không phụ thuộc địa chỉ client, nên vẫn
  giữ nguyên khi kết nối migrate.
- `ja4plus` (v0.0.3) chọn ký tự giao thức theo `LocalAddr().Network()`; với
  QUIC giá trị này là `"udp"` nên nó ghi `d` (DTLS). Với handshake `"udp"`,
  `TLSConfig` đưa cho `ja4plus` một bản sao `ClientHelloInfo` có conn báo
  `"quic"`, nên nó ghi `q` như spec JA4 yêu cầu. Edge không phục vụ DTLS.
- Test: [transport_quic_test.go](../../internal/dmz/edge/transport/transport_quic_test.go)
  chạy handshake QUIC thật và kiểm tra fingerprint `q13…`, SNI và việc xoá
  mục sau khi đóng kết nối.

## 3.3 `szhttp.Context`

Interface trừu tượng hoá request/response
([core/szhttp/context.go](../../staging/src/github.com/sentinez/core/szhttp/context.go)),
gồm `RequestContext` (đọc/ghi header, query, path, body, IP, JA4, ...) +
`ResponseWriter` + `X()` trả về `edgepb.ContextExtra` (chứa `namespace`).

Có ba cài đặt:

| Cài đặt | File | Dùng khi |
|---|---|---|
| `httphz.Context` | `contrib/httphz/request_context.go` | Edge (Hertz) |
| `httpx.Context` | `pkg/network/httpx/context.go` | Edge chế độ `Quic` (HTTP/3 và TCP) và `Standard`, realtime |
| `httpreq.RequestContext` | `core/szhttp/httpreq/context_value.go` | Đánh giá rule trên `httppb.Request` (hiện chỉ test dùng) |

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

Dựng trong [internal/handlers/handlers.go](../../internal/handlers/handlers.go) bằng
`chains.Build(envplugin.NewEnv(mem, opts...), Stage, plugins()...)`. Mỗi
plugin cài `chains.Plugin` (`ID()`, `Handle(ctx, next)`) và, nếu cần
dependency, `Provision(*envplugin.Env)`:
[Env](../../internal/envplugin/env.go) resolve option một lần và cấp các
sub-store của `MemStore`, hostname và logger sự kiện (`Env.Logger(kind)`),
nên plugin không nhận tham số constructor. Mỗi plugin được bọc bởi
`handlers.Stage` để gắn `ID()` của plugin vào lỗi
(`sztracer.StageError{Stage, Elapsed, Err}`) và log một lần duy nhất ở stage
gây lỗi; tên stage là hằng exported (vd `secure.WAFId`). Edge đóng chain
(`Chain.Close`) khi shutdown.

| Thứ tự | Stage | Plugin | Hành vi |
|---|---|---|---|
| 0 | `TRC` | `trace.Tracer` | Sinh request ID `senz:req:<xid>` nếu context chưa có |
| 1 | `LOG` | `logging.Logging` | Gọi `next` trước, sau đó log `httppb.Event` (JSON, kind `LOG_TYPE_HTTP`) |
| 2 | `DMA` | `secure.DomainBased` | Host phải có dạng `<ns>.<SENZ_HOSTNAME>` (một cấp, bỏ port). Sai → `403`. Gán `X().Namespace = ns` |
| 3 | `PUB` | `cdn.Public` | Phục vụ ảnh embed trong binary dưới `/cdn/sentinez/public/`; nạp ảnh lỗi lúc provision thì edge không khởi động |
| 4 | `CDN` | `cdn.Cache` | Nếu có CDN rule active cho namespace và expression khớp: tra cache theo key `method+scheme+host+path+query+Accept-Encoding`; hit → trả response với `X-Cache: HIT`; miss → chạy tiếp và lưu response cache được (TTL 1h, xem [04-rule-engine.md](04-rule-engine.md#44-cdn-rules)) |
| 5 | `LMT` | `secure.Limiter` | với mỗi limiter có `ingress.expr` (bắt buộc) khớp, theo priority như `RUL`: `limiter.Allow(IP)`; vượt → `429` |
| 6 | `ROM` | `room.WaitingRoom` | Placeholder, chỉ chuyển tiếp |
| 7 | `STC` | `static.Static` | Sau khi chạy tiếp, nếu path có đuôi tĩnh (`.css .js .png ...`) thì set `Cache-Control: public, max-age=3600, immutable` |
| 8 | `RUL` | `secure.SecRule` | Duyệt các SecRule active theo priority giảm dần; rule đầu tiên khớp có action `BLOCK` → `403`; action khác chỉ log debug và đi tiếp |
| 9 | `WAF` | `secure.WAF` | Coraza transaction: `ExecIngress` (header/body) → chạy tiếp → `ExecEgress` (response). Bị chặn với status 403 → trang 403. Cuối cùng log `secrulepb.Event` (`LOG_TYPE_WAF`) nếu bị interrupt |
| 10 | `ROU` | `routing.StandardRouter` | Kết thúc chain: tìm route, rewrite path, set header, gọi `ReverseProxy.Serve` |

Response lỗi (`400/403/404/429/500`) được render bằng template `templ`
([core/internal/render](../../staging/src/github.com/sentinez/core/internal/render/))
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
    core_ruleset:                # OWASP CRS, mỗi namespace một cấu hình
      status: STATUS_ACTIVE
      mode: ENGINE_MODE_ON       # ENGINE_MODE_OFF | ENGINE_MODE_DETECTION_ONLY | ENGINE_MODE_ON
      version: VERSION_V4_17_0   # VERSION_V4_16_0 (chỉ RCE, SQLi) | VERSION_V4_17_0 (mặc định)
      paranoiaLevel: 1           # 1 (ít false positive nhất) .. 4 (chặt nhất)
      detectionParanoiaLevel: 2  # log tới mức này, chặn tới paranoiaLevel; 0 = bằng nhau
      threshold:                 # điểm anomaly để chặn, 0 = giá trị setup.conf
        inbound: 5
        outbound: 4
        earlyBlocking: true
      categories:                # nhóm rule được bật, rỗng = tất cả
        - CATEGORY_RCE           # 932
        - CATEGORY_XSS           # 941
        - CATEGORY_SQLI          # 942
      request:
        bodyAccess: true         # kiểm tra request body
        bodyLimitBytes: 13107200 # 12.5 MiB, 0 = mặc định
        allowedMethods: [GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS]
        # allowedContentTypes: [application/json]   # rỗng = giá trị setup.conf
      exclusions:                # tinh chỉnh false positive
        - location: /api/search  # bỏ rule SQLi dưới prefix này
          tags: [attack-sqli]
        - targets:               # không có location = cả namespace
            - ruleId: 942100
              variable: ARGS:password
      overrides:                 # trạng thái từng rule
        - id: 920420
          state: RULE_STATE_DISABLED        # chỉ gỡ dưới /webhook
          location: /webhook
        - id: 942430
          state: RULE_STATE_DETECTION_ONLY  # chỉ log, không tính điểm
        - id: 942130
          state: RULE_STATE_ENABLED         # ép bật dù trên paranoiaLevel
    sec_rules:                   # SecRule (dạng Lite)
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
        maxRequests: 500
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

- `security.core_ruleset` là một `coreruleset.v1.CoreRuleset`
  ([setting.proto](../../api/proto/sentinez/types/coreruleset/v1/setting.proto)).
  `version` chọn bộ rule nhúng trong binary (`VERSION_V4_16_0` chỉ có nhóm RCE
  và SQLi; không đặt = `VERSION_V4_17_0`). `paranoiaLevel` và
  `detectionParanoiaLevel` giới hạn 0–4; rule ID trong `exclusions` và
  `overrides` phải nằm trong 900000–999999; `location` phải bắt đầu bằng `/`.
- Giá trị `categories` ánh xạ 1:1 với file CRS: nhóm request
  `COMMON_EXCEPTIONS` (905), `METHOD_ENFORCEMENT` (911), `SCANNER_DETECTION`
  (913), `PROTOCOL_ENFORCEMENT` (920), `PROTOCOL_ATTACK` (921),
  `MULTIPART_ATTACK` (922), `LFI` (930), `RFI` (931), `RCE` (932), `PHP`
  (933), `GENERIC` (934), `XSS` (941), `SQLI` (942), `SESSION_FIXATION`
  (943), `JAVA` (944); nhóm response `DATA_LEAKAGES` (950),
  `DATA_LEAKAGES_SQL` (951), `DATA_LEAKAGES_JAVA` (952), `DATA_LEAKAGES_PHP`
  (953), `DATA_LEAKAGES_IIS` (954), `WEB_SHELLS` (955), `DATA_LEAKAGES_RUBY`
  (956). Tất cả đều có tiền tố `CATEGORY_`.
- `overrides[].state`: `RULE_STATE_DISABLED` có thể giới hạn theo `location`;
  `RULE_STATE_DETECTION_ONLY` và `RULE_STATE_ENABLED` luôn áp dụng cho cả
  namespace.
- Phần "Lite" (`SecRuleLite`, `cdn.RuleLite`, `ExpressionLite`) dùng chuỗi
  cho `source/operator/status/value`. `NormalizeEdgeSetting` chuyển sang bản
  runtime (`SecRule`, `cdn.Rule`, `Expression`) với enum thật, `value` thành
  `structpb.Value` dạng string, và sinh ID `srule_<uuid>` /
  `cond_<nanoid>`.
- Biến hỗ trợ trong `proxySetHeaders`
  ([core/szhttp/variable](../../staging/src/github.com/sentinez/core/szhttp/variable/variable.go)):
  `$host`, `$remote_addr`, `$scheme`, `$request_uri`,
  `$proxy_add_x_forwarded_for`. Chỉ hỗ trợ khi **toàn bộ giá trị** là một biến;
  chuỗi khác chứa `$` gây lỗi khi route. Header key không hợp lệ (RFC 7230) bị
  bỏ qua với cảnh báo.
- Edge luôn set thêm `X-Forwarded-Prefix: <location>`.

## 3.6 Routing

[internal/memory/routes/routes.go](../../internal/memory/routes/routes.go):

- `Store(server)`: sao chép location, bỏ qua (kèm cảnh báo) location không
  có `proxyPass`, lọc header hợp lệ, đặt `proxyRewrite` mặc định, **sắp xếp
  giảm dần theo độ dài `location`** để khớp longest-prefix như nginx. Server
  không còn location nào thì bị gỡ.
- `Match(ctx)`: lấy route theo namespace, tìm location đầu tiên khớp path
  **theo ranh giới segment** (`/api` khớp `/api` và `/api/x` nhưng không khớp
  `/apix`; location kết thúc bằng `/` khớp mọi path bên dưới), đặt path mới =
  `rewrite + (path − location)` mà không lặp `/` ở chỗ nối, set header, trả về
  upstream kế tiếp trong `proxyPass` theo **round robin** (mỗi location một
  bộ đếm atomic; là `balanceStrategy` duy nhất, dùng cả khi không khai báo).

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
  `Upgrade: websocket` chuyển sang `WSReverseProxy`, target
  (`upstream + URI`) được dựng riêng cho từng request.
- `net/http` (`httpx.ReverseProxy`, dùng cho engine `Quic` và `Standard`):
  `httputil.NewSingleHostReverseProxy` với `http.RoundTripper` truyền vào
  (`network.StandardTransporter()`, dial timeout 30s). Writer được bọc để
  context thấy status và body của upstream; `Unwrap` giữ `Flush`/`Hijack`
  cho streaming và upgrade.

## 3.7 Memory store

`memory.MemStore` (singleton) gom các kho theo **namespace** (`server.name`):

| Kho | Kiểu | Nội dung |
|---|---|---|
| `settings.Setting` | `Map[string,*edgepb.Setting]` | Setting đã normalize theo namespace, dựng lại từ setting được đồng bộ (3.8) |
| `routes.Router` | `Map[string,[]*Location]` | Location đã sắp xếp |
| `reverseproxy.ReverseProxy` | `Map[string,szhttp.ReverseProxy]` | Key = upstream server |
| `secrules.SecRule` | `Map[string,[]Entry]` | `Entry{Eval, Rule}` đã compile, lọc active, sort theo priority |
| `ratelimiter.Limiter` | `Map[string,*RateLimiter]` | Một limiter / namespace |
| `rulesets.Rulesets` | `Map[string,coraza.WAF]` | WAF instance / namespace |
| `cdnrules.Rule` | `Map[string,EvalFunc]` + `Map[string,*cdn.CDN]` | Một CDN rule / namespace |

Mọi map dùng `shared/szsync.Map[K,V]` (bọc `sync.Map` có generic).

## 3.8 Cluster (Olric)

[pkg/cluster](../../pkg/cluster/):

- `NewCluster` tạo Olric với config `local`. Nếu config là `Clustered()`
  (có cả `SENZ_MEMBERSHIP_ADDRESS` và `SENZ_DISCOVERY_ADDRESS`) thì đặt `Peers`
  và bind memberlist theo discovery address.
- Olric không có compare-and-swap, nên không được giữ state do leader ghi;
  state đó nằm ở Consul, nơi việc ghi có fencing (xem bên dưới).
- `Cluster.Ready()` được đóng khi Olric nhận kết nối, dù chạy cluster hay
  không, và `Cluster.Node()` là tên của process (`<hostname>-<6 ký tự>`).

### Đồng bộ memory giữa các edge

`cluster.SyncMap[T]` ([syncmap.go](../../pkg/cluster/syncmap.go),
[syncmap_run.go](../../pkg/cluster/syncmap_run.go)) giữ một map giá trị proto
giống nhau trên mọi edge của cluster, mà không đưa Olric vào đường đi của
request. Đo trên một node, `Get` của Olric embedded mất ~500 ns so với ~20 ns
của `Load` trên `sync.Map` (~1,3 µs so với 2 ns khi đọc song song một key
nóng), và key do node khác giữ còn tốn thêm một round trip mạng.

| Thao tác | Việc nó làm |
|---|---|
| `Load`, `Range` | Chỉ đọc `szsync.Map` local: không lock, không bao giờ qua Olric |
| `Store(ctx, key, v)`, `Delete(ctx, key)` | Áp dụng local, `Put` record vào DMap `<name>`, rồi `Publish` nó lên kênh PubSub Olric `sentinez.sync.<name>` |
| `Start(ctx)` | Đợi `Cluster.Ready()`, subscribe, rồi nạp toàn bộ DMap (`Scan` + `Get`) |
| mỗi 30s | Đối soát: lấy các record trong DMap mới hơn bản local, ghi lại các bản local mà DMap thiếu hoặc giữ bản cũ hơn |
| `Close()` | Dừng đồng bộ; giá trị local vẫn giữ |

- **Thứ tự.** Mỗi lần ghi là một record `{version, origin, deleted,
  payload}`. `version` lấy từ hybrid logical clock: nano giây theo đồng hồ
  thật, không bao giờ lùi và luôn vượt mọi version node đã áp dụng, nên lần
  ghi tiếp theo của một node luôn xếp sau những gì nó đã thấy dù đồng hồ
  lệch. Hoà thì so `origin` (tên node). Record chỉ được áp dụng nếu mới hơn:
  người ghi sau cùng thắng, và mọi node hội tụ về cùng một giá trị.
- **Mất message.** PubSub của Olric là at most once: node không đang
  subscribe (restart, mạng chia cắt) sẽ lỡ message. `Start` subscribe trước
  rồi mới nạp nên không mất lần ghi nào ở giữa, phần còn lại do đối soát
  định kỳ vá trong vòng 30s.
- **Ghi đồng thời.** Hai node cùng ghi một key có thể để lại record cũ hơn
  trong DMap (hai `Put` tranh nhau); lần đối soát của node giữ bản mới hơn sẽ
  ghi đè lại.
- **Xoá** là tombstone: giữ 24h trong DMap (`EX`) và ở local, để giá trị cũ
  không quay lại.
- **Ghi trước `Start`** chỉ ở local; lần nạp đầu tiên sẽ chia sẻ chúng.

Không dùng cho: bộ đếm (hai lần tăng cùng lúc sẽ mất một: dùng `Incr` của
Olric), và state cần thứ tự chặt giữa các node (đồng hồ lệch quá giới hạn
của HLC có thể đảo thứ tự hai lần ghi đồng thời): state đó nằm ở Consul có
fencing.

`OnChange(fn(key, prev, next))` báo mỗi lần một key đổi giá trị trên node,
dù do ghi tại đó, nhận về hay đối soát; `prev`/`next` là `nil` khi key chưa
có hoặc không còn giá trị. Các lần gọi tuần tự theo đúng thứ tự áp dụng.

### Setting giữa các edge

Các store hot path (route, reverse proxy, CDN, SecRule, limiter, WAF) giữ
giá trị đã compile, không gửi qua mạng được (`coraza.WAF`, proxy kèm
transport, limiter kèm bộ đếm). Vì vậy các edge đồng bộ **nguồn** là
`edgepb.Setting` thô, và mỗi node tự compile store của mình từ đó
([internal/memory/sync.go](../../internal/memory/sync.go)):

- `MemStore` giữ một `SyncMap` `sentinez.dmz.edge.settings` key theo
  `metadata.resource_id`, id của domain resource. Setting không có id chỉ
  được nạp trên edge của nó, kèm cảnh báo.
- `MemStore.Sync(ctx, server, local...)` khởi động map (đợi Olric, nạp
  setting của cluster), rồi chia sẻ `local`. `SyncMap.Store` bỏ qua giá trị
  bằng giá trị đang giữ (cùng mã hoá tất định), nên restart không làm mọi
  edge dựng lại nó.
- Mỗi lần thay đổi (`OnChange`), node normalize một bản sao
  (`sznormalize.NormalizeEdgeSetting`) rồi dựng lại store của namespace
  (`server.name`) mà request vẫn dùng để định tuyến. Namespace đổi tên hoặc
  resource bị xoá thì được gỡ trước.
- Một namespace thuộc về một resource: resource thứ hai nhận cùng namespace
  sẽ không được nạp (có log), và không gỡ được namespace đó.
- Request vẫn chỉ đọc các `szsync.Map` local (~20 ns); Olric không bao giờ
  nằm trên đường đi của request.

Bộ đếm rate limiter và JA4 fingerprint không được đồng bộ: mỗi edge giới
hạn phần traffic của nó, và một kết nối TLS chỉ đến một edge.

### Bầu leader theo zone

[pkg/cluster/leader](../../pkg/cluster/leader/) bầu ra một edge duy nhất làm
leader cho zone của nó (`SENZ_ZONE`). Không dùng Olric cho việc này: chính
docs của Olric gọi lock của nó là "approximate" (lock nằm trên một partition
owner, mặc định không có replica, không có quorum và không có token tăng
dần), nên không đảm bảo được chỉ có một leader. Việc bầu chọn đi qua session
lock của Consul, được Consul ghi qua Raft log.

| Cấu hình | Hành vi |
|---|---|
| `SENZ_CONSUL_URI` + `SENZ_ZONE` | Tranh lock trên key Consul `sentinez/dmz/edge/leader/<zone>`; zone được lowercase giống tiền tố id của `szrand`, nên `SGN` và `sgn` dùng chung một key |
| `conf.Clustered()` mà thiếu `SENZ_CONSUL_URI` | Khởi động thất bại (`errNoConsul`): cluster không bao giờ chạy mà không có bầu chọn |
| `SENZ_CONSUL_URI` mà thiếu `SENZ_ZONE` | Khởi động thất bại (`errNoZone`) |
| Không có cả hai | Standalone: node làm leader ngay (term 1) và không bao giờ mất |

`conf.Clustered()` ([config.go](../../api/proto/sentinez/types/setting/v1/config.go))
trả về true khi có cả `SENZ_MEMBERSHIP_ADDRESS` và `SENZ_DISCOVERY_ADDRESS`;
`cluster.newConfig` dùng cùng điều kiện này để join peer.

Một term diễn ra như sau:

1. `consul.Client.Acquire` ([api/client/consul/lock.go](../../api/client/consul/lock.go))
   tạo một session (TTL 15s, do node giữ lock gia hạn) và chờ tới khi giữ
   được key. Mỗi lần chờ một lock đang bị giữ tối đa 2s, nên follower nhận
   ra shutdown sớm, trong giới hạn stop timeout 15s của runner.
2. Node thắng đọc `LockIndex` của key bằng consistent read: đó là
   `Term.Index`, tăng mỗi lần lock đổi chủ. Giá trị của lock là node id
   (`<hostname>-<6 ký tự ngẫu nhiên>`), được so sánh để chắc index thuộc về
   node này.
3. `Elector.Run(ctx, task)` công bố term (`Elector.Leader()` trả về nó) và
   chạy `task(ctx, term)`.
4. Term kết thúc khi mất lock (`Lost()` bị đóng, sau 3 lần monitor thử lại),
   khi task trả về lỗi, hoặc khi `ctx` kết thúc. Elector xoá `Leader()`
   **trước** khi cancel `ctx` của task, chờ task dừng, nhả lock rồi tranh
   lại (chờ 5s nếu term thất bại).

Task trả về `nil` thì vẫn giữ quyền leader cho tới khi mất lock;
`Run(ctx, nil)` chỉ giữ quyền leader.

Chuyển leader:

- Dừng bình thường: lock được nhả, node khác lên leader trong vài mili giây.
- Leader chết hoặc mất kết nối Consul: session hết hạn sau TTL, rồi tới
  lock-delay của Consul (15s), nên zone không có leader khoảng 30s.

**Fencing.** Leader cũ vẫn có thể hành động thêm một lúc (GC pause, mạng bị
chia cắt) trước khi thấy `Lost()`. Mọi lần leader ghi state dùng chung phải
gửi kèm `term.Index`, và nơi lưu trữ phải từ chối index nhỏ hơn index lớn
nhất đã thấy. Riêng elector không thể tự đảm bảo điều này;
`consul.Client.PutFenced` ([api/client/consul/fenced.go](../../api/client/consul/fenced.go))
đảm bảo việc đó trên Consul KV:

- Giá trị được lưu dạng JSON `{"token": <term.Index>, "value": ...}`
  (`consul.Fenced[T]`, đọc lại bằng `consul.GetFenced[T]`).
- Mỗi lần ghi đọc key (consistent read), trả `consul.ErrFenced` nếu token
  đang lưu lớn hơn token của người ghi (bằng nhau là cùng một leader), rồi
  ghi bằng CAS trên `ModifyIndex` vừa đọc, nên kiểm tra và ghi là nguyên tử.
  CAS thua một người ghi khác thì thử lại vài lần (`_casRetries`).
- Olric không có fencing: nó không có compare-and-swap (chỉ có `Put`
  `NX`/`XX`, `GetPut`, `Incr` và `Lock` gần đúng), nên kiểm tra token trên đó
  vẫn có thể bị race. State dùng chung do leader ghi nằm ở Consul.

**Bản ghi của zone.** Leader giữ zone của nó dưới
`sentinez/dmz/edge/zone/<zone>/` ([zone.go](../../pkg/cluster/leader/zone.go)),
mọi lần ghi đều qua `PutFenced` với `term.Index`:

- `Elector.PublishStatus()` ([status.go](../../pkg/cluster/leader/status.go))
  ghi `<zone>/status` ngay lập tức, rồi mỗi 10s. Đây là `Task` edge chạy.
- `Elector.Publish(name, value)` ghi `<zone>/<name>` mỗi term một lần, thử
  lại mỗi 10s tới khi ghi được, dành cho các bản ghi zone-wide sau này.
- `leader.All(tasks...)` ([task.go](../../pkg/cluster/leader/task.go)) chạy
  nhiều task trong một term; task đầu tiên lỗi sẽ huỷ các task còn lại và
  kết thúc term.

Bản ghi `status`:

| Trường | Giá trị |
|---|---|
| `zone` | `SENZ_ZONE` viết thường |
| `node`, `term` | Leader và `term.Index` của nó |
| `updated_at` | Thời điểm ghi (UTC); bản ghi cũ hơn ~10s nghĩa là không có leader nào đang ghi |

Ghi bị fence (`ErrFenced`) thì kết thúc term: đã có leader mới hơn ghi bản
ghi, nên node nhả lock rồi tranh cử lại. Lỗi khác (không kết nối được
Consul) được log và thử lại ở tick sau. Edge standalone không có Consul nên
`PublishStatus` trả `nil`, node chỉ giữ quyền leader.

**Wiring.** [cmd/szedge/v1/server/leader.go](../../cmd/szedge/v1/server/leader.go)
chạy elector như một szrunner app riêng (`Inject(leader.New)`,
`Serve(e.Run(ctx, e.PublishStatus()))`). Mỗi app có container fx riêng, nên app
`edgeServer`/`grpcServer` không truy cập được `Elector`; việc cần làm cho cả
zone phải đặt trong `Task`.

**Deploy.** [deploy/docker/cluster/docker-compose.yaml](../../deploy/docker/cluster/docker-compose.yaml)
chạy hai edge trong zone `SGN` cùng một Consul ở dev mode. Production cần 3
hoặc 5 server Consul: nếu Consul không đạt quorum thì không zone nào bầu
được leader.

## 3.9 gRPC EdgeService

`edgepb.EdgeService` chỉ còn `Status`, trả về `OK`
([internal/dmz/edge/api.go](../../internal/dmz/edge/api.go), package `edge`).
Nó được đóng gói trong `edge.Service` (`pkg/apps/dmz/edge/edge_service.go`), và
`main` khởi chạy nó qua app `grpcServer` trên `defaults.EdgeAddress`. RPC
`EvalCoreRuleset` (chạy lại một request với một setting CRS) đã bị bỏ cùng
`szcrs.Evaluate`. `internal/dmz/edge/events/{pub,sub}` hiện là package rỗng.

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
