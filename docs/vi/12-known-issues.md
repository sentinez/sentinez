# 12. Các vấn đề đã biết

Danh sách được tổng hợp khi đọc code tại commit `47e5f876`, rà lại tại
`d64307d3`. Mỗi mục ghi rõ vị trí và hậu quả để dễ tạo issue. Mục đánh dấu
*(cần xác minh)* là suy luận từ code, chưa được chạy thử. Khi sửa xong, hãy
xoá mục tương ứng.

## 12.1 Lỗi logic (ảnh hưởng hành vi)

| # | Vị trí | Vấn đề | Hậu quả |
|---|---|---|---|
| 2 | [controlplane/apps/security/v1/repos/secrule/secrule.go](../../staging/src/github.com/sentinez/controlplane/apps/security/v1/repos/secrule/secrule.go) `Update` | Chỉ cập nhật `name`, `description`, `expr`, `status` | Sửa `priority`/`action` qua API không được lưu |
| 3 | [pkg/cluster/cluster.go](../../pkg/cluster/cluster.go) `newConfig` | Khi có cả membership và discovery address hợp lệ, hàm `return conf` trước khi gán `conf.Started` | `dictReady` không bao giờ được báo → DMap không khởi tạo; sau 1024 lần `Put`, goroutine gọi sẽ bị block |
| 7 | [contrib/httphz/proxy/ws_reverse_proxy.go](../../staging/src/github.com/sentinez/contrib/httphz/proxy/ws_reverse_proxy.go) | `p.target += string(uri)` sửa trường của proxy dùng chung | Target WebSocket dài thêm sau mỗi request, có data race |
| 8 | [contrib/httphz/server.go](../../staging/src/github.com/sentinez/contrib/httphz/server.go) `Handle` | Vòng `fn = s.chains[i](fn)` nằm **trong** handler, gán lại biến capture | Mỗi request bọc thêm middleware một lần nữa (hiện chưa lộ vì edge không gọi `Use`) |
| 9 | [internal/memory/routes/routes.go](../../internal/memory/routes/routes.go) | Luôn dùng `proxyPass[0]`; không kiểm tra rỗng; bỏ qua `balanceStrategy` | Location không có upstream gây panic; không cân bằng tải |
| 11 | [internal/handlers/logging/logging.go](../../internal/handlers/logging/logging.go) | Gọi `bpf.LookupBandwidth` mỗi request, nhưng object eBPF chỉ được nạp trong tiến trình dataplane | Luôn lỗi/0, tốn chi phí và log `Infof` ở mọi request |
| 12 | [internal/dmz/dataplane/bpf/security.go](../../internal/dmz/dataplane/bpf/security.go) `BlockCIDR` | Ghi `ip` bằng `binary.BigEndian.Uint32` trong khi XDP tra bằng `saddr` thô (thứ tự byte mạng trong bộ nhớ) | Trên máy little-endian khoá không khớp, CIDR bị chặn không có tác dụng *(cần xác minh)*. Hàm hiện chưa được gọi |

## 12.2 Bảo mật

| # | Vị trí | Vấn đề |
|---|---|---|
| S1 | [pkg/apps/gateway/apiserver/middleware/middleware.go](../../pkg/apps/gateway/apiserver/middleware/middleware.go) `AllowCORS` | Phản chiếu mọi `Origin` kèm `Allow-Credentials: true` (code có ghi chú không dùng cho production) |
| S2 | Handler controlplane | Phân quyền phải tự gọi trong từng handler; chỉ `ListAccounts`, `ListActivities` có kiểm tra. Tenant, Security, `ListUsers`, `DeleteUser`… không yêu cầu token |
| S3 | IAM `loginAdmin` | So mật khẩu admin dạng plaintext, không constant-time; admin không có giới hạn đăng nhập sai |
| S4 | Middleware `Logging` (apiserver) | Đọc toàn bộ body và log ở mức debug khi lỗi — có thể ghi mật khẩu từ `/iam/login` vào log |
| S5 | [cmd/szedge/v1/server/pprof.go](../../cmd/szedge/v1/server/pprof.go) | pprof luôn mở ở `defaults.PprofAddress` (`0.0.0.0:6060`; compose cũng publish port này) |
| S6 | [pkg/network/httpx/context.go](../../pkg/network/httpx/context.go), WebSocket realtime | `CheckOrigin` luôn `true` |
| S7 | `cmd/szedge/v1/Dockerfile` | Image đóng gói cặp cert/key từ thư mục source |
| S8 | `httpx.Context.RequestIP` | Tin `X-Forwarded-For`/`X-Real-IP` từ client mà không có danh sách proxy tin cậy → giả mạo IP để né rule IP/rate limit (chế độ `Quic` và `Standard`, tức mặc định) |

## 12.3 Chức năng chưa hoàn thiện

- Không có kênh đồng bộ cấu hình control plane → edge: SecRule/tenant trong
  DB không ảnh hưởng edge; edge chỉ đọc `proxy.yaml` lúc khởi động, không
  reload.
- `szcentraldata`: `main()` rỗng. `AnalyticService` không được mount vào
  apiserver. Discovery Consul (`WithDiscorvery`, `VisitToEndpoint`) có code
  nhưng chưa bật.
- `EdgeService.EvaluateRuleset` đánh giá
  `Expression` rỗng (luôn khớp) và bỏ qua `ruleset_id`.
- Waiting room (`ROM`), `shared/szqueue`, `shared/sztopic`, `shared/szstorage/{keyval,
  clickhouse,elastic}`, `eventpub/eventsub` là package rỗng.
- `DataPlaneService` chưa có handler; `wsz.WebSocket.Shutdown` là no-op.
- Tính duy nhất trong IAM (mỗi email một user, mỗi user một account cho mỗi
  provider, username của `LOCAL`) chỉ được service kiểm tra trước khi
  insert; bảng không có unique index (`dbx.Index` chỉ một cột, không
  unique), nên hai lượt đăng ký đồng thời có thể cùng lọt.
- IAM: user đăng ký bằng passkey chưa thể thêm mật khẩu (chưa có RPC "link"
  cho `LOCAL`). Console có `PasskeyLink()` trong `lib/api/iam/passkey.ts`
  nhưng chưa trang nào gọi.
- IAM `UpdateUser` đổi `User.email` mà không cập nhật email trên các account
  của user, và lỗi nếu chưa có account nào mang email đó.
- Session passkey và registration đang chờ nằm trong bộ nhớ tiến trình: mất
  khi restart và không dùng chung giữa các bản sao apiserver.
- `SENZ_TIMESCALE_URI`, `SENZ_CLICKHOUSE_URI` khai báo nhưng chưa dùng.
