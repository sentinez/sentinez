# 14. Kế hoạch kiểm thử release

Đây là các bước kiểm tra một bản release phải vượt qua trước khi gắn tag
([13-release.md](13-release.md), bước 3). Mục 14.2 liệt kê các cổng tự động,
mục 14.3 liệt kê các test thủ công trên hệ thống đang chạy. Ghi kết quả mỗi
lần chạy vào 14.5.

## 14.1 Môi trường

Dùng một bản checkout mới trên Linux. Cài đặt theo
[02-getting-started.md](02-getting-started.md):

- `make compose.up`: Postgres/TimescaleDB, Consul, Kafka, ClickHouse và OTLP
  collector.
- `sudo tee -a /etc/hosts < deploy/hosts`.
- Certificate cho edge: `./hack/mkcert-local.sh`.
- Một database **mới**. Schema sync không migrate database cũ
  ([12-known-issues.md](12-known-issues.md), L1).
- Các service, mỗi cái chạy bằng `make sz.<svc>.run`: apiserver, realtime,
  greeter, collector, centraldata và edge. Dataplane chỉ cần cho nhóm `DPL`.
- Console: `cd ui && pnpm install && pnpm dev` (http://localhost:3000).
- Một upstream ở `sentinez.upstream:80` trả lại nguyên request, kể cả header.
  Ví dụ: `docker run --rm -p 80:8080 mendhak/http-https-echo`.

Trong các test edge, `$EDGE` là domain của setting đang test, gọi qua địa chỉ
listen của edge. Các test dùng file
[cmd/szedge/v1/proxy.yaml](../../cmd/szedge/v1/proxy.yaml) có sẵn trong repo,
với domain `badcheese.is.s6z.io.vn`. Muốn chạy không cần internet thì trỏ
domain của file này vào một upstream local.

## 14.2 Cổng tự động

| Id | Lệnh | Pass khi |
|---|---|---|
| A-01 | `make release.check` | Exit 0. Target chạy gofmt, golangci-lint trong mọi module, `buf lint`, `go test` trong mọi module, rồi build mọi service |
| A-02 | `cd ui && pnpm lint && pnpm typecheck && pnpm build` | Cả ba exit 0 (lint chạy với `--max-warnings 0`) |
| A-03 | `go test -race ./internal/handlers/... ./pkg/apps/dmz/...` (ở root) | Không có race nào được báo |
| A-04 | `make bench` | Không benchmark nào của đường xử lý request chậm hơn quá 10% so với bảng trong README |
| A-05 | `cd api && buf breaking --against '.git#branch=main'` | Không phải cổng chặn cho 0.0.1 (chưa có bản release trước). Ghi lại kết quả để dùng cho các bản sau |

Các test sau đã bảo vệ những ràng buộc này:

- Tên stage Lua khớp với chain của edge:
  `internal/handlers/script_stages_test.go`.
- File `proxy.yaml` có sẵn import được vào control plane:
  `pkg/apps/gateway/apiserver/proxyimport_test.go`.
- Setting của edge được chuẩn hoá: `pkg/sznormalize`.

## 14.3 Test thủ công

Mỗi test ghi các bước làm và kết quả mong đợi. Chạy theo thứ tự, vì các test
sau dùng lại trạng thái mà các test trước tạo ra.

### INS: cài đặt và khởi động

| Id | Các bước | Kết quả mong đợi |
|---|---|---|
| INS-01 | `make` | Mọi binary được build ra `cmd/<svc>/bin` hoặc `cmd/<svc>/v1/bin` |
| INS-02 | Chạy apiserver khi không có `SZ_DATA_KEY` | Không khởi động được, log ghi rõ key bị thiếu |
| INS-03 | Chạy apiserver với `SZ_CONSUL_URI` trỏ tới địa chỉ không có ai lắng nghe | Không khởi động được |
| INS-04 | Chạy mọi service | Banner của mỗi service in version release (`core.Version`). Các bảng được tạo trong database mới |
| INS-05 | Mở http://sentinez.gateway.apiserver:8080/swagger/ | Swagger UI liệt kê các service IAM, tenant, security và analytic |
| INS-06 | `make consul.proxy.put`, rồi khởi động lại apiserver và edge | Apiserver import file vào tenant `admin`. Edge đọc setting từ Consul |

### EDG: định tuyến của edge

| Id | Các bước | Kết quả mong đợi |
|---|---|---|
| EDG-01 | `curl -k https://$EDGE/local/anything` | Upstream nhận `/anything` (do `proxyRewrite`) kèm `X-Forwarded-*`, `X-Real-IP` và `Host` |
| EDG-02 | Gửi request với một `Host` mà không setting nào có | Trang `403` từ `domain` |
| EDG-03 | `curl -k --http3 https://$EDGE/local/` | Được phục vụ qua HTTP/3. Log của request có JA4 fingerprint |
| EDG-04 | Làm cho mọi upstream của một location không truy cập được | Trang `502` |
| EDG-05 | Xem log của một request | Dòng log có request id, `tenant_id`, `resource_id`, và các trường GeoIP nếu là IP public |

### SEC: SecRule, limiter và CRS

| Id | Các bước | Kết quả mong đợi |
|---|---|---|
| SEC-01 | `curl -k https://$EDGE/block` (GET) | `403`. Rule "(Path AND Method) OR IP" khớp |
| SEC-02 | `curl -k -X POST https://$EDGE/block` | Rule đó không chặn |
| SEC-03 | `curl -k -H 'X-Forwarded-For: 1.1.1.1' https://$EDGE/` | `403` từ nhánh IP. Test này cũng cho thấy vấn đề đã biết S8: client tự chọn được IP của mình |
| SEC-04 | Mở `https://$EDGE/login` bằng trình duyệt | Trang proof-of-work. Sau khi tick checkbox, trang tải xong và cookie `pow-pass` được đặt. Các GET tiếp theo trong vòng một giờ không bị challenge nữa |
| SEC-05 | Gửi 101 request trong dưới 5 giây từ cùng một IP: `for i in $(seq 101); do curl -sk -o /dev/null -w '%{http_code}\n' https://$EDGE/local/; done \| sort \| uniq -c` | Request thứ 101 nhận `429`. IP bị giới hạn trong `timeout` (5 giây), sau đó được phục vụ lại |
| SEC-06 | Trong lúc SEC-05 đang chạy, gửi request từ một IP thứ hai | IP thứ hai không bị giới hạn (giới hạn tính theo từng key) |
| SEC-07 | `curl -k "https://$EDGE/local/?q=<script>alert(1)</script>"` khi CRS đặt `mode: ENGINE_MODE_ON` | Trang `403`. Một event `LOG_TYPE_SECRULESET` được ghi log |
| SEC-08 | Lặp lại SEC-07 với `mode: ENGINE_MODE_DETECTION_ONLY` | Request tới được upstream, event vẫn được ghi log |
| SEC-09 | Trong `proxy.yaml`, loại trừ rule mà SEC-07 dính cho `/local`; khởi động lại edge | Request của SEC-07 tới được upstream |
| SEC-10 | Chạy SEC-07 với `version: CRS_VERSION_V4_16_0` | Kết quả giống SEC-07 |

### CDN: cache

| Id | Các bước | Kết quả mong đợi |
|---|---|---|
| CDN-01 | GET cùng một URL cache được hai lần | Response thứ hai có `X-Cache: HIT`, upstream chỉ nhận một request |
| CDN-02 | GET `/login` hai lần | Không có `X-Cache: HIT`. Challenge ở SEC-04 vẫn áp dụng |
| CDN-03 | Cùng URL nhưng khác `Accept-Encoding` | Miss (cache key khác) |

### LUA: hook scripting

| Id | Các bước | Kết quả mong đợi |
|---|---|---|
| LUA-01 | Dùng `script_example` có sẵn; gửi request tới `/local/` | Upstream nhận header `X-Edge: sentinez`. Khi bật debug log, dòng `tag` có trong log của request |
| LUA-02 | Hook `http.request/secrule` bằng một hàm gọi `ctx:respond(418, "teapot")`, rồi gửi request tới `/block` | Nhận `418 teapot`. Stage SecRule bị bỏ qua nên không có `403` |
| LUA-03 | Hook `http.response/router` bằng một hàm gọi `ctx:set_response_header("X-Seen", "1")` | Response có `X-Seen: 1` |
| LUA-04 | Một hàm lặp vô hạn | Mỗi request vẫn được phục vụ, script chạy tối đa khoảng 50 ms. Lỗi của script được ghi log và chain chạy tiếp |
| LUA-05 | `---@stage http.request/sec_rate_limit` (tên cũ) | Edge ghi log `edge: load '<domain>': … @stage: unknown stage "sec_rate_limit", one of caching, secratelimit, …` và không nạp script nào của namespace, các script đã nạp từ trước được giữ lại. Phần còn lại của namespace vẫn được nạp |
| LUA-06 | Gọi `os.execute`, `io.open` hoặc `require` trong script | Không dùng được (bị sandbox chặn) |
| LUA-07 | `script` của secrule: `if ctx:method() == "GET" and (ctx:path():sub(1, 2) == "/a" or ctx:path():sub(1, 2) == "/b") then return SZ_BLOCK end`; gửi `GET /a`, `GET /c`, `POST /a` | Chỉ `GET /a` bị chặn. Event `secrule: matched` đặt tên rule bằng điều kiện Lua, giữ nguyên ngoặc |
| LUA-08 | `script` của secrule dùng `re.match(ctx:path(), "^/wp-")`, `ctx:ip_in("10.0.0.0/8")`, `ctx:has_header("X-Debug")` hoặc `ctx:tls()` | Nạp được và khớp như `expr` tương ứng (`MATCHES`, CIDR của IP, có header, TLS) |
| LUA-09 | `script` của secrule có biến local, vòng lặp, hoặc pattern `re.match` sai | Setting bị từ chối với lỗi bọc `not declarative`, chain cũ được giữ |
| LUA-10 | `expr` của secrule có hai nhóm `or`, nhóm thứ hai trong `andCondition` (rule "Challenge GET/HEAD on /a or /b" có sẵn) | `GET /a/x` bị challenge; `GET /c` và `POST /a/x` thì không |

### CLS: cluster

| Id | Các bước | Kết quả mong đợi |
|---|---|---|
| CLS-01 | Hai edge có `SZ_MEMBERSHIP_ADDRESS`, `SZ_DISCOVERY_ADDRESS`, `SZ_CONSUL_URI` và cùng `SZ_ZONE` | Hai edge join với nhau. Một edge làm leader: key `sentinez/dmz/edge/zone/<zone>/status` được ghi |
| CLS-02 | Dừng edge leader | Edge còn lại tiếp quản key status |
| CLS-03 | Một setting có `tenantId` khác mặc định | Setting được đồng bộ giữa các edge. Setting dùng `tnt_default` thì không được đồng bộ (đúng thiết kế) |

### API: control plane qua REST và console

| Id | Các bước | Kết quả mong đợi |
|---|---|---|
| API-01 | `PUT /iam/login` bằng tài khoản `SZ_ADMIN_USERNAME` | Nhận được JWT |
| API-02 | `POST /iam/account` (tạo user local mới), rồi đăng nhập | Cả hai thành công. Đăng ký lần hai với cùng email bị từ chối |
| API-03 | Đăng ký passkey trong console ở http://localhost:3000, đăng xuất, rồi đăng nhập bằng passkey | Cả hai thành công |
| API-04 | User A tạo một tenant và một resource (domain). User B, không phải thành viên, liệt kê resource | B không thấy resource của A (các lời gọi được giới hạn theo tenant của người gọi) |
| API-05 | Thêm B làm thành viên tenant của A | B thấy resource đó |
| API-06 | CRUD SecRule (`/security/secrule(s)`) và trang console `security/sec-rule` | Tạo, xem, sửa, liệt kê và xoá đều chạy đúng. Trình soạn điều kiện tạo ra expression hợp lệ |
| API-07 | CRUD rate limit và ruleset (`/security/sec{ratelimit,ruleset}(s)`) cùng các trang console của chúng | Như API-06 |
| API-08 | Upload một CA certificate cho một resource | Được lưu dạng đã seal, không đọc được PEM trong bảng. Đọc lại qua API thì được |
| API-09 | Sửa một resource, rồi mở trang activity | Có một mục audit ghi người gọi và các trường đã thay đổi |
| API-10 | Gọi một endpoint PORTAL/ADMIN không kèm token, rồi kèm một token đã hết hạn | `401` |

### EVT: event và analytics

| Id | Các bước | Kết quả mong đợi |
|---|---|---|
| EVT-01 | Edge chạy với `SZ_KAFKA_BROKERS`; chạy lại SEC-01, SEC-05 và SEC-07, kèm một ít traffic bình thường | Message xuất hiện trên các topic `sentinez.event.http.*` (Kafka UI :8088) |
| EVT-02 | Collector đang chạy | Các dòng có trong `analytic.*.events`. Mỗi id có tiền tố là `SZ_ZONE` |
| EVT-03 | Các trang console `analytic/logs/sec-rule`, `sec-ruleset` và `sec-rate-limit` | Liệt kê các event của EVT-01. Top values và filter hoạt động |
| EVT-04 | Zoom histogram từ giờ xuống phút xuống giây | Các bucket khớp với số event ở mỗi độ phân giải |
| EVT-05 | Biểu đồ egress latency | Hiển thị latency upstream của EDG-01 |
| EVT-06 | Edge chỉ có `SZ_OTLP_EXTERNAL_ENDPOINT` | Traffic log tới được OTLP collector |

### CDT: centraldata

| Id | Các bước | Kết quả mong đợi |
|---|---|---|
| CDT-01 | Chạy centraldata qua mTLS và trỏ edge vào nó; tạo một SecRule cho domain trong control plane | Edge nạp setting từ các bảng được đồng bộ (xem các giới hạn ở 12.3) |
| CDT-02 | Một client không có certificate hợp lệ | Kết nối mTLS bị từ chối |

### UI: console

| Id | Các bước | Kết quả mong đợi |
|---|---|---|
| UI-01 | Mở mọi trang trong sidebar (`tests/e2e/screenshots.mjs` đi qua tất cả) | Không trang nào lỗi. Spinner của tab che mỗi lần chuyển tab |
| UI-02 | Chuyển qua lại giữa tiếng Việt và tiếng Anh | Mọi chuỗi hiển thị đều được dịch, không còn key thô |
| UI-03 | Dừng apiserver, rồi mở console | Đăng nhập bằng sample user vẫn được (proxy trả 502) |
| UI-04 | Màn hình mobile (375 px) | Sidebar thu gọn, không trang nào bị cuộn ngang |
| UI-05 | `ui/apps/web` (`pnpm build && pnpm start`) | Trang public hiển thị đúng, các trang docs hoạt động |

### DPL: dataplane (Linux, quyền root)

| Id | Các bước | Kết quả mong đợi |
|---|---|---|
| DPL-01 | `sudo make -f hack/net/dev/Makefile setup && make sz.dataplane.run` | Chương trình XDP được gắn vào bên trong netns `gateway` |
| DPL-02 | Gửi traffic từ netns `client` qua edge | Edge dùng socket cookie làm TCP conn id |

## 14.4 Tiêu chí hoàn thành

- Mọi cổng A đều pass.
- Mọi test thủ công đều pass, trừ các mục được đánh dấu *đã biết* bên dưới.
  Một lỗi trùng với vấn đề đã biết ([12-known-issues.md](12-known-issues.md))
  chỉ được chấp nhận nếu changelog có ghi lại.
- Không bỏ sót mục mới nào của 12.2 (bảo mật) khỏi release notes.

Các kết quả *đã biết* của 0.0.1:

- SEC-03 giả mạo được IP (S8).
- Cổng `:6060` trả về pprof (S5).
- Edge chạy đơn lẻ không nhận thay đổi từ control plane cho tới khi restart
  (12.3, #393).

## 14.5 Kết quả

Chép bảng này vào PR release. Đánh dấu mỗi dòng là pass, fail (kèm link
issue) hoặc n/a.

| Nhóm | Người test | Ngày | Commit | Kết quả | Ghi chú |
|---|---|---|---|---|---|
| A (cổng tự động) | | | | | |
| INS | | | | | |
| EDG | | | | | |
| SEC | | | | | |
| CDN | | | | | |
| LUA | | | | | |
| CLS | | | | | |
| API | | | | | |
| EVT | | | | | |
| CDT | | | | | |
| UI | | | | | |
| DPL | | | | | |
