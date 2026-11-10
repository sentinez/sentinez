# 12. Các vấn đề đã biết

Danh sách được tổng hợp khi đọc code tại commit `47e5f876`, rà lại tại
`d64307d3`. Mỗi mục ghi rõ vị trí và hậu quả để dễ tạo issue. Mục đánh dấu
*(cần xác minh)* là suy luận từ code, chưa được chạy thử. Khi sửa xong, hãy
xoá mục tương ứng.

## 12.1 Lỗi logic (ảnh hưởng hành vi)

| # | Vị trí | Vấn đề |
|---|---|---|
| L1 | [controlplane/pkg/schema](../../staging/src/github.com/sentinez/controlplane/pkg/schema/) từ `28dd2173` *(cần xác minh)* | Cột timestamp của các bảng iam, tenant, security và activities đã đổi tên (`created_at` → `metadata_created_at` / `owner_created_at`) và `security.sec_rules.action` đổi từ `JSONB` sang `TEXT`, nhưng sync schema không bao giờ đổi tên/đổi kiểu cột: với database tạo từ trước, scanner được sinh sẽ select cột không tồn tại. Xoá bảng (dev) hoặc migrate tay |

## 12.2 Bảo mật

| # | Vị trí | Vấn đề |
|---|---|---|
| S1 | [pkg/apps/gateway/apiserver/middleware/middleware.go](../../pkg/apps/gateway/apiserver/middleware/middleware.go) `AllowCORS` | Phản chiếu mọi `Origin` kèm `Allow-Credentials: true` (code có ghi chú không dùng cho production) |
| S2 | Handler controlplane | Phân quyền phải tự gọi trong từng handler (`headers.Allow`); chỉ `ListAccounts`, Tenant và Analytic có kiểm tra. Security, `ListUsers`, `DeleteUser`… không yêu cầu token |
| S3 | IAM `loginAdmin` | So mật khẩu admin dạng plaintext, không constant-time; admin không có giới hạn đăng nhập sai |
| S4 | Middleware `Logging` (apiserver) | Đọc toàn bộ body và log ở mức debug khi lỗi — có thể ghi mật khẩu từ `/iam/login` vào log |
| S5 | [cmd/szedge/v1/server/pprof.go](../../cmd/szedge/v1/server/pprof.go) | pprof luôn mở ở `defaults.PprofAddress` (`0.0.0.0:6060`; compose cũng publish port này) |
| S6 | [pkg/network/httpx/context.go](../../pkg/network/httpx/context.go), WebSocket realtime | `CheckOrigin` luôn `true` |
| S7 | `cmd/szedge/v1/Dockerfile` | Image đóng gói cặp cert/key từ thư mục source |
| S8 | `httpx.Context.RequestIP` | Tin `X-Forwarded-For`/`X-Real-IP` từ client mà không có danh sách proxy tin cậy → giả mạo IP để né rule IP/rate limit (chế độ `Quic` và `Standard`, tức mặc định) |
| S9 | Đọc event của Analytic | Chưa lọc theo tenant: không truyền `resource_ids` thì mọi token hợp lệ đều đọc được event của mọi tenant |

## 12.3 Chức năng chưa hoàn thiện

- Không có kênh đồng bộ cấu hình control plane → edge: SecRule/tenant trong
  DB không ảnh hưởng edge; edge chỉ đọc `proxy.yaml` lúc khởi động, không
  reload.
- Setting không khai báo `metadata.tenantId` / `server.resourceId` dùng
  chung id `tnt_default` / `rsrc_default`, nên event của chúng bị gộp
  chung trong các bảng event; setting dùng tenant id mặc định không được
  sync giữa các edge.
- Chọn rõ `HISTOGRAM_RESOLUTION_SECOND` mà không có `from` sẽ đọc cả 7 ngày
  của view giây.
- Database cũ vẫn giữ index đơn `resource_id` dư thừa: index phức hợp
  `(resource_id, <time>)` đã được thêm cho mọi bảng event, nhưng sync schema
  không bao giờ xoá index.
- [deploy/docker/docker-compose.yaml](../../deploy/docker/docker-compose.yaml)
  tạo các topic event với 3 partition và tắt tự tạo topic, nên
  `DefaultTopicPartitions` (6) của producer không bao giờ có tác dụng trong
  môi trường đó.
- `szcentraldata`: `main()` rỗng. Discovery Consul (`WithDiscorvery`,
  `VisitToEndpoint`) có code nhưng chưa bật.
- Leader theo zone mới chỉ ghi zone status
  (`sentinez/dmz/edge/zone/<zone>/status`); `Elector.Publish` và
  `leader.All` đã sẵn cho các bản ghi zone-wide khác nhưng chưa có bản ghi
  nào. Cũng chưa có gì đọc status.
- Gỡ một namespace (resource bị xoá, `server.resourceDomain` đổi) vẫn giữ các
  reverse proxy của nó: chúng được key theo upstream và dùng chung giữa các
  namespace, không có đếm tham chiếu.
- Waiting room (`ROM`), `pkg/szqueue`, `pkg/sztopic`, `pkg/szstorage/{keyval,
  clickhouse,elastic}`, `eventpub/eventsub` là package rỗng.
- `DataPlaneService` chưa có handler; `wsz.WebSocket.Shutdown` là no-op.
- Tính duy nhất trong IAM (mỗi email một user, mỗi user một account cho mỗi
  provider, username của `LOCAL`) chỉ được service kiểm tra trước khi
  insert; bảng không có unique index (chưa dùng unique index của schema ở
  đây), nên hai lượt đăng ký đồng thời có thể cùng lọt.
- IAM: user đăng ký bằng passkey chưa thể thêm mật khẩu (chưa có RPC "link"
  cho `LOCAL`). Console có `PasskeyLink()` trong `lib/api/iam/passkey.ts`
  nhưng chưa trang nào gọi.
- IAM `UpdateUser` đổi `User.email` mà không cập nhật email trên các account
  của user, và lỗi nếu chưa có account nào mang email đó.
- Session passkey và registration đang chờ nằm trong bộ nhớ tiến trình: mất
  khi restart và không dùng chung giữa các bản sao apiserver.
- `SENZ_TIMESCALE_URI`, `SENZ_CLICKHOUSE_URI` khai báo nhưng chưa dùng.
