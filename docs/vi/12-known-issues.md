# 12. Các vấn đề đã biết

Danh sách được tổng hợp khi đọc code tại commit `47e5f876`, rà lại tại
`d64307d3` và `83b43464`. Mỗi mục ghi rõ vị trí và hậu quả để dễ tạo issue. Mục đánh dấu
*(cần xác minh)* là suy luận từ code, chưa được chạy thử. Khi sửa xong, hãy
xoá mục tương ứng.

## 12.1 Lỗi logic (ảnh hưởng hành vi)

| # | Vị trí | Vấn đề |
|---|---|---|
| L1 ([#395](https://github.com/sentinez/sentinez/issues/395)) | [controlplane/pkg/schema](../../staging/src/github.com/sentinez/controlplane/pkg/schema/) từ `28dd2173` *(cần xác minh)* | Cột timestamp của các bảng iam, tenant, security và activities đã đổi tên (`created_at` → `metadata_created_at` / `owner_created_at`) và `security.secrules.action` đổi từ `JSONB` sang `TEXT`, nhưng sync schema không bao giờ đổi tên/đổi kiểu cột: với database tạo từ trước, scanner được sinh sẽ select cột không tồn tại. Xoá bảng (dev) hoặc migrate tay |

## 12.2 Bảo mật

| # | Vị trí | Vấn đề |
|---|---|---|
| S4 ([#388](https://github.com/sentinez/sentinez/issues/388)) | Middleware `Logging` (apiserver) | Đọc toàn bộ body và log ở mức debug khi lỗi — có thể ghi mật khẩu từ `/iam/login` vào log. Body còn được đọc mà không giới hạn kích thước, trước cả bước xác thực (có thể làm cạn bộ nhớ); `middleware/injection.go` cũng vậy |
| S5 ([#389](https://github.com/sentinez/sentinez/issues/389)) | [cmd/szedge/v1/server/pprof.go](../../cmd/szedge/v1/server/pprof.go) | pprof luôn mở ở `defaults.PprofAddress` (`0.0.0.0:6060`; compose cũng publish port này) |
| S6 ([#390](https://github.com/sentinez/sentinez/issues/390)) | [pkg/network/httpx/context.go](../../pkg/network/httpx/context.go), WebSocket realtime | `CheckOrigin` luôn `true` |
| S7 ([#391](https://github.com/sentinez/sentinez/issues/391)) | `cmd/szedge/v1/Dockerfile` | Image đóng gói cặp cert/key từ thư mục source |
| S8 ([#392](https://github.com/sentinez/sentinez/issues/392)) | `httpx.Context.RequestIP` | Tin `X-Forwarded-For`/`X-Real-IP` từ client mà không có danh sách proxy tin cậy → giả mạo IP để né rule IP/rate limit (chế độ `Quic` và `Standard`, tức mặc định) |

## 12.3 Chức năng chưa hoàn thiện

- [#393](https://github.com/sentinez/sentinez/issues/393) Edge chạy đơn lẻ không nhận được thay đổi từ control plane: thay
  đổi chỉ tới edge qua `szcentraldata`, do zone leader của edge chạy cluster
  theo dõi (cần Consul, `SZ_SERVICE_TOKEN`, `SZ_SYNC_KEY`). Ngoài trường hợp
  đó, edge chỉ đọc `proxy.yaml` một lần lúc khởi động và không reload.
- [#394](https://github.com/sentinez/sentinez/issues/394) Setting không khai báo `metadata.tenantId` /
  `ingress.resource.id` dùng chung id `tnt_default` / `rsrc_default`, nên
  event của chúng bị gộp chung trong các bảng event; setting dùng tenant id
  mặc định không được sync giữa các edge.
- [#395](https://github.com/sentinez/sentinez/issues/395) Database cũ vẫn giữ index đơn `resource_id` dư thừa: index phức hợp
  `(resource_id, <time>)` đã được thêm cho mọi bảng event, nhưng sync schema
  không bao giờ xoá index (cùng nguyên nhân với L1: không có migration).
- [#396](https://github.com/sentinez/sentinez/issues/396) [deploy/docker/docker-compose.yaml](../../deploy/docker/docker-compose.yaml)
  tạo các topic event với 3 partition và tắt tự tạo topic, nên
  `DefaultTopicPartitions` (6) của producer không bao giờ có tác dụng trong
  môi trường đó.
- [#397](https://github.com/sentinez/sentinez/issues/397) Đồng bộ `szcentraldata`: edge nạp các bảng thành `edgepb.Setting`
  dựng lại từ row, chỉ giữ loại action của rule (không giữ `params`) và không
  có gì về CDN: CDN của file được giữ. CA certificate tới được edge (đã seal)
  nhưng edge chưa dùng tới. Edge chiếm namespace của setting từ file khi các
  bảng có cùng domain.
- [#398](https://github.com/sentinez/sentinez/issues/398) Zone leader ghi zone status
  (`sentinez/dmz/edge/zone/<zone>/status`) và theo dõi centraldata; chưa có
  gì đọc status.
- [#399](https://github.com/sentinez/sentinez/issues/399) Gỡ một namespace (resource bị xoá, `ingress.resource.domain` đổi)
  vẫn giữ các reverse proxy thường (`plainProxy`) của nó: chúng được key theo
  upstream và dùng chung giữa các namespace, không có đếm tham chiếu.
- [#400](https://github.com/sentinez/sentinez/issues/400) Waiting room (`waitingroom`) chỉ chuyển tiếp request, `pkg/szqueue` rỗng.
- [#401](https://github.com/sentinez/sentinez/issues/401) `pkg/sztopic`, `pkg/szstorage/{cache/keyval,dbx/clickhouse,
  search/elastic}` là package rỗng; `SZ_TIMESCALE_URI`, `SZ_CLICKHOUSE_URI`
  khai báo nhưng chưa dùng.
- [#402](https://github.com/sentinez/sentinez/issues/402) `DataPlaneService` chưa có handler.
- [#403](https://github.com/sentinez/sentinez/issues/403) `wsz.WebSocket.Shutdown` là no-op (`httpx/std.Shutdown` chỉ là
  placeholder).
- [#404](https://github.com/sentinez/sentinez/issues/404) Tính duy nhất trong IAM (mỗi email một user, mỗi user một account
  cho mỗi provider, username của `LOCAL`) chỉ được service kiểm tra trước khi
  insert; bảng không có unique index (chưa dùng unique index của schema ở
  đây), nên hai lượt đăng ký đồng thời có thể cùng lọt.
- [#405](https://github.com/sentinez/sentinez/issues/405) IAM: user đăng ký bằng passkey chưa thể thêm mật khẩu (chưa có RPC
  "link" cho `LOCAL`). Console có `PasskeyLink()` trong
  `lib/api/iam/passkey.ts` nhưng chưa trang nào gọi.
- [#406](https://github.com/sentinez/sentinez/issues/406) IAM `UpdateUser` đổi `User.email` mà không cập nhật email trên các
  account của user, và lỗi nếu chưa có account nào mang email đó.
- [#407](https://github.com/sentinez/sentinez/issues/407) Session passkey và registration đang chờ nằm trong bộ nhớ tiến
  trình: mất khi restart và không dùng chung giữa các bản sao apiserver.

## 12.4 Phát hành và công cụ

Các mục này được phát hiện khi chuẩn bị 0.0.1 ([13-release.md](13-release.md)).
R1 chặn việc release.

| # | Vị trí | Vấn đề |
|---|---|---|
| R1 ([#383](https://github.com/sentinez/sentinez/issues/383)) | `go.mod` của `core`, `pkg`, `controlplane`, `contrib/httphz`, `tools` | Các module require `sentinez/api v0.0.0`, pseudo-version rỗng `v0.0.0-00010101000000-…` của `contrib/pkg` và `pkg`, hoặc pseudo-version của commit trong monorepo. Go bỏ qua `replace` của dependency, nên `go get` các module này từ ngoài repo sẽ lỗi. Trước khi gắn tag, phải đổi các require sang version release (13.2) |
| R3 ([#385](https://github.com/sentinez/sentinez/issues/385)) | Các `api/proto/**/generate.sh` có gọi `protoc-go-inject-tag` | Bước inject làm tag dài ra nhưng không căn lại comment, nên các file `*.pb.go` sinh ra fail `gofmt` (`make fmt.go.check`) sau mỗi lần generate. CI không chạy `fmt.go.check` và cũng không build, nên lỗi này không bị phát hiện |
| R4 ([#386](https://github.com/sentinez/sentinez/issues/386)) | `cmd/sz*/Makefile`, `sz.<svc>.image.build` | Image chỉ được gắn tag `$(TAG):latest`; tag version phải tự thêm bằng `docker tag` |
| R5 ([#387](https://github.com/sentinez/sentinez/issues/387)) | [hack/gomodules.sh](../../hack/gomodules.sh) | `bpf` và `integration` được publish nhưng không có trong danh sách, nên lint, `go mod tidy`, test, CI và `make release.check` đều bỏ qua chúng |
