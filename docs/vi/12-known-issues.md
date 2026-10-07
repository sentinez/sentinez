# 12. Các vấn đề đã biết

Danh sách được tổng hợp khi đọc code tại commit `47e5f876`. Mỗi mục ghi rõ vị
trí và hậu quả để dễ tạo issue. Mục đánh dấu *(cần xác minh)* là suy luận từ
code, chưa được chạy thử. Khi sửa xong, hãy xoá mục tương ứng.

## 12.1 Lỗi logic (ảnh hưởng hành vi)

| # | Vị trí | Vấn đề | Hậu quả |
|---|---|---|---|
| 1 | [internal/memory/memory.go](../../internal/memory/memory.go) `LoadRateLimiter` + [core/common/edge_normalize.go](../../staging/src/github.com/sentinez/core/common/edge_normalize.go) | Limiter chỉ được nạp khi `IngressRuntime.Status == ACTIVE`, nhưng `NormalizeEdgeSetting` chỉ điền `IngressRuntime` cho `secRules`, không cho `limiters`. Ngoài ra gặp limiter không active thì `return nil` (bỏ cả các limiter sau) | Rate limit cấu hình trong `proxy.yaml` **không bao giờ có hiệu lực** |
| 2 | [controlplane/security/v1/repos/secrule/secrule.go](../../staging/src/github.com/sentinez/controlplane/security/v1/repos/secrule/secrule.go) `Update` | Chỉ cập nhật `name`, `description`, `expr`, `status` | Sửa `priority`/`action` qua API không được lưu |
| 3 | [internal/cluster/cluster.go](../../internal/cluster/cluster.go) `newConfig` | Khi có cả membership và discovery address hợp lệ, hàm `return conf` trước khi gán `conf.Started` | `dictReady` không bao giờ được báo → DMap không khởi tạo; sau 1024 lần `Put`, goroutine gọi sẽ bị block |
| 4 | [cmd/szedge/v1/main.go](../../cmd/szedge/v1/main.go) + [core/runner/hook.go](../../staging/src/github.com/sentinez/core/runner/hook.go) | `engine.Hertz` gắn `OnConnect` trong một hook OnStart, còn `server.Start` ở hook OnStart khác; mỗi hook chạy trong goroutine riêng | Race: `Start` có thể đọc `options` trước khi `SetOptions` chạy → `opt.OnConnect` nil, panic khi có kết nối mới *(cần xác minh)* |
| 5 | [core/runner/hook.go](../../staging/src/github.com/sentinez/core/runner/hook.go) `OnStart` | Lỗi trả về từ `start` chỉ được log nếu là `http.ErrServerClosed`; lỗi khác bị nuốt | Service lỗi khi listen (port bận, cert sai, ...) vẫn "chạy" mà không có log |
| 6 | [core/runner/runner.go](../../staging/src/github.com/sentinez/core/runner/runner.go) `_OTLP` | `Insecure: secure` — giá trị bị đảo | Endpoint `https://...` bị dùng không TLS; endpoint thường lại dùng TLS |
| 7 | [contrib/httphz/proxy/ws_reverse_proxy.go](../../staging/src/github.com/sentinez/contrib/httphz/proxy/ws_reverse_proxy.go) | `p.target += string(uri)` sửa trường của proxy dùng chung | Target WebSocket dài thêm sau mỗi request, có data race |
| 8 | [contrib/httphz/server.go](../../staging/src/github.com/sentinez/contrib/httphz/server.go) `Handle` | Vòng `fn = s.chains[i](fn)` nằm **trong** handler, gán lại biến capture | Mỗi request bọc thêm middleware một lần nữa (hiện chưa lộ vì edge không gọi `Use`) |
| 9 | [internal/memory/routes/routes.go](../../internal/memory/routes/routes.go) | Luôn dùng `proxyPass[0]`; không kiểm tra rỗng; bỏ qua `balanceStrategy`; ghép path không chuẩn hoá | Location không có upstream gây panic; không cân bằng tải; có thể sinh `//path` |
| 10 | [internal/memory/memory.go](../../internal/memory/memory.go) `LoadRulesets` | Bỏ qua nội dung `rulesets[]`; luôn CRS v4.16.0 + flag RCE; kiểm tra status bị comment | Không cấu hình được phiên bản/nhóm rule; SQLi (942) không bao giờ bật |
| 11 | [internal/funcs/logging/logging.go](../../internal/funcs/logging/logging.go) | Gọi `bpf.LookupBandwidth` mỗi request, nhưng object eBPF chỉ được nạp trong tiến trình dataplane | Luôn lỗi/0, tốn chi phí và log `Infof` ở mọi request |
| 12 | [internal/bpf/security.go](../../internal/bpf/security.go) `BlockCIDR` | Ghi `ip` bằng `binary.BigEndian.Uint32` trong khi XDP tra bằng `saddr` thô (thứ tự byte mạng trong bộ nhớ) | Trên máy little-endian khoá không khớp, CIDR bị chặn không có tác dụng *(cần xác minh)*. Hàm hiện chưa được gọi |
| 13 | [shared/eventq/eventq.go](../../staging/src/github.com/sentinez/shared/eventq/eventq.go) `Close`, `CloseAll` | Ép kiểu `chan string` / `chan any` thay vì `chan T` | Panic với `T` khác (chưa có nơi dùng) |

## 12.2 Bảo mật

| # | Vị trí | Vấn đề |
|---|---|---|
| S1 | [pkg/apps/gateway/apiserver/middleware/middleware.go](../../pkg/apps/gateway/apiserver/middleware/middleware.go) `AllowCORS` | Phản chiếu mọi `Origin` kèm `Allow-Credentials: true` (code có ghi chú không dùng cho production) |
| S2 | Handler controlplane | Phân quyền phải tự gọi trong từng handler; chỉ `ListAccounts`, `ListActivities` có kiểm tra. Tenant, Security, `ListUsers`, `DeleteUser`… không yêu cầu token |
| S3 | IAM `loginAdmin` | So mật khẩu admin dạng plaintext, không constant-time; admin không có giới hạn đăng nhập sai |
| S4 | Middleware `Logging` (apiserver) | Đọc toàn bộ body và log ở mức debug khi lỗi — có thể ghi mật khẩu từ `/iam/login` vào log |
| S5 | [cmd/szedge/v1/main.go](../../cmd/szedge/v1/main.go) `init` | pprof luôn mở ở `:6060` (compose cũng publish port này) |
| S6 | [pkg/network/httpx/std/context.go](../../pkg/network/httpx/std/context.go), WebSocket realtime | `CheckOrigin` luôn `true` |
| S7 | `cmd/szedge/v1/Dockerfile` | Image đóng gói cặp cert/key từ thư mục source |
| S8 | `stdhttpx.Context.RequestIP` | Tin `X-Forwarded-For`/`X-Real-IP` từ client mà không có danh sách proxy tin cậy → giả mạo IP để né rule IP/rate limit (chế độ `Standard`) |

## 12.3 Chức năng chưa hoàn thiện

- Không có kênh đồng bộ cấu hình control plane → edge: SecRule/tenant trong
  DB không ảnh hưởng edge; edge chỉ đọc `proxy.yaml` lúc khởi động, không
  reload.
- `szcentraldata`: `main()` rỗng. `AnalyticService` không được mount vào
  apiserver. Discovery Consul (`WithDiscorvery`, `VisitToEndpoint`) có code
  nhưng chưa bật.
- `EdgeService` gRPC không được khởi chạy; `EvaluateRuleset` đánh giá
  `Expression` rỗng (luôn khớp) và bỏ qua `ruleset_id`.
- SecRule action ngoài `BLOCK` chưa thực thi; toán tử `GT/GTE/LT/LTE` chưa cài.
- Waiting room (`ROM`), `pkg/queue`, `shared/topic`, `core/storage/{keyval,
  clickhouse,kafka,elastic}`, `eventpub/eventsub` là package rỗng.
- Mỗi namespace chỉ giữ một CDN rule; một limiter.
- `DataPlaneService` chưa có handler; `wsz.WebSocket.Shutdown` và
  `stdhttpx.Shutdown` là no-op.
- `SENZ_TIMESCALE_URI`, `SENZ_CLICKHOUSE_URI` khai báo nhưng chưa dùng.

## 12.4 Vấn đề nhỏ / vệ sinh code

- Tiền tố ID không khớp giữa code và ràng buộc proto: repo sinh
  `senz.security.sec_rules.<uuid>` nhưng proto yêu cầu `senz.security.secrules.`;
  user `senz.iam.users.` nhưng `GetUserRequest/UpdateUserRequest/
  DeleteUserRequest` yêu cầu `senz.users.`. Hiện không lộ vì gRPC server chưa
  bật validate request.
- `corehttp.Forbidden/NotFound/...` dùng một `sync.Mutex` toàn cục khi render,
  tuần tự hoá mọi response lỗi.
- `preflightHandler` set `Access-Control-Max-Age` sau `WriteHeader` (không có
  tác dụng); `extendHeader` của grpc-gateway set header `Server` sau khi
  handler đã ghi response.
- `GenerateRulesets` ghi `WAF.conf.lock` vào thư mục làm việc mỗi lần tạo WAF.
- Đường dẫn mặc định cũ: `--env_file` của greeter (`./cmd/mesh/greeter/v1/.env`)
  và realtime (`./cmd/realtime/.env`); `hack/mkcert-local.sh` chuyển cert vào
  `./cmd/edge/v1/`.
- `tenant` repo `scan` bỏ qua `rows.Err()`; `UpdateResource` không cập nhật
  `resource_setting`.
- `LoadRateLimiter` dùng `zlog.Fatalf` khi parse duration lỗi (dừng cả edge).
- CI chỉ lint/test module root; module staging chỉ được lint qua `make lint`
  local.
