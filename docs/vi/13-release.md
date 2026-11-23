# 13. Phát hành

Tài liệu này mô tả cách phát hành một phiên bản Sentinéz. Nó được viết cho
0.0.1, bản phát hành đầu tiên; các bản sau làm theo cùng các bước. Những gì
thay đổi được ghi vào [CHANGELOG.md](../../CHANGELOG.md), còn những gì cần
kiểm thử nằm trong [14-test-plan.md](14-test-plan.md).

## 13.1 Phiên bản

- Theo [SemVer](https://semver.org/). Trước 1.0.0 không cam kết tương thích:
  proto, `proxy.yaml`, REST API và schema có thể đổi ở bất kỳ bản nào.
- Một version chung cho tất cả. Repo, các module staging, image và console
  đều phát hành cùng một số.
- `core.Version`
  ([core/core.go](../../staging/src/github.com/sentinez/core/core.go)) là
  version in ra ở banner lúc khởi động. Hiện nó là một `const` có giá trị
  `v0.0.1-beta`, nên `-ldflags -X` không ghi đè được. Khi release, hãy sửa
  tay giá trị này, hoặc đổi nó sang `var` rồi truyền
  `-X github.com/sentinez/core.Version=...` từ Makefile và Dockerfile.

| Thành phần | Nơi phát hành | Tag / version |
|---|---|---|
| Module root `github.com/sentinez/sentinez` | repo này | `v0.0.1` |
| `github.com/sentinez/sentinez/api` | repo này, `api/` | `api/v0.0.1` |
| `core`, `pkg`, `controlplane`, `tools`, `scripting`, `bpf`, `integration` | repo tách `github.com/sentinez/<name>` | `v0.0.1` |
| `contrib/pkg`, `contrib/httphz` | repo tách `github.com/sentinez/contrib` | `pkg/v0.0.1`, `httphz/v0.0.1` |
| `go-import` | repo tách | không gắn tag (đây là server vanity import, không phải thư viện) |
| Image | `sentinez/{apiserver,edge,greeter,centraldata,collector}` | `v0.0.1` |
| Console, web | `ui/package.json`, `ui/apps/{console,web}/package.json` | `0.0.1` (đã đặt sẵn) |

## 13.2 Trước khi gắn tag: go.mod của các module staging

Trong repo này, các module dùng nhau qua `replace` với đường dẫn tương đối.
Go bỏ qua `replace` của dependency, nên khi một module staging được tải
riêng, các dòng `require` của nó được resolve đúng như đang ghi:

| Module | Các require không resolve được ngoài repo |
|---|---|
| `core` | `contrib/pkg v0.0.0-00010101000000-…`, `pkg v0.0.0-00010101000000-…`, `sentinez/api v0.0.0` |
| `pkg` | `sentinez/api v0.0.0` |
| `controlplane` | pseudo-version của `core` và `pkg` trỏ vào commit mà repo tách có thể không có, `sentinez/api v0.0.0`, `contrib/pkg v0.0.0-00010101000000-…` |
| `contrib/httphz`, `tools` | giống `core` |

`contrib/pkg`, `scripting`, `bpf` và `integration` không phụ thuộc module
Sentinéz nào nên dùng được ngay.

Trước khi gắn tag, đặt các require này về `v0.0.1`. Giữ nguyên `replace`,
nên build local không thay đổi. Sau đó chạy `go mod tidy` trong từng module
và ở root. Gắn tag theo thứ tự phụ thuộc, để mọi version mà một module
require đều đã tồn tại khi module đó được tải:

1. `api/v0.0.1` (repo này), `pkg/v0.0.1` (repo `contrib`), `scripting`,
   `bpf`, `integration`
2. `pkg`
3. `core`
4. `httphz/v0.0.1` (repo `contrib`), `tools`, `controlplane`
5. root `v0.0.1`

## 13.3 Checklist

1. **Đóng băng.** Merge những gì sẽ vào bản release. Tạo nhánh
   `release/0.0.1` từ `main`; CI lint và test cũng chạy trên `release/*`.
2. **Cổng tự động.** Chạy `make release.check` và đảm bảo nó pass. Target
   này kiểm tra gofmt, chạy golangci-lint và `go test` trong mọi module của
   [hack/gomodules.sh](../../hack/gomodules.sh), chạy `buf lint`, rồi build
   mọi service. Sau đó chạy `pnpm lint && pnpm typecheck && pnpm build` trong
   `ui/`.
3. **Test plan.** Chạy [14-test-plan.md](14-test-plan.md) trên một môi trường
   mới và ghi kết quả vào bảng của nó.
4. **Vấn đề đã biết.** Rà lại [12-known-issues.md](12-known-issues.md). Với
   mỗi mục bảo mật (S4 đến S8), hoặc sửa, hoặc ghi vào release notes. Không
   publish image edge khi nó vẫn còn đóng gói sẵn certificate của source tree
   (S7).
5. **Version.** Đặt `core.Version` thành `v0.0.1` và sửa go.mod như mục 13.2.
   Trong [CHANGELOG.md](../../CHANGELOG.md), thay `Unreleased` bằng ngày phát
   hành.
6. **Merge** nhánh release vào `main` (qua PR). `ci_sync_repo` sẽ đẩy từng
   module staging lên nhánh `main` của repo tách tương ứng.
7. **Gắn tag** theo thứ tự ở 13.2. Job sync chỉ đẩy nhánh, không đẩy tag. Gắn
   tag cho từng repo tách trên commit mà job sync đã đẩy, ví dụ:

   ```sh
   git clone git@github.com:sentinez/core.git && cd core
   git tag -a v0.0.1 -m "v0.0.1" origin/main && git push origin v0.0.1
   ```

   Ở repo này, gắn tag lên merge commit:

   ```sh
   git tag -a api/v0.0.1 -m "api v0.0.1"
   git tag -a v0.0.1 -m "v0.0.1"
   git push origin api/v0.0.1 v0.0.1
   bash hack/gitlab_sync.sh     # mirror GitLab, kèm tag
   ```

8. **Image.** `sz.<svc>.image.build` luôn gắn tag `$(TAG):latest` và không có
   tag version, nên phải tự thêm tag version:

   ```sh
   make sz.edge.image.build
   docker tag sentinez/edge:latest sentinez/edge:v0.0.1
   docker push sentinez/edge:v0.0.1 && docker push sentinez/edge:latest
   ```

   Lặp lại cho apiserver, greeter, centraldata và collector.
9. **GitHub release.** Tạo release trên tag `v0.0.1`. Nội dung release notes
   là mục 0.0.1 của changelog.

## 13.4 Sau khi phát hành

- Kiểm tra từng module qua proxy:
  `GOPROXY=https://proxy.golang.org go list -m github.com/sentinez/<name>@v0.0.1`.
- Trong một module trống, chạy `go get github.com/sentinez/core@v0.0.1` rồi
  `go build ./...`. Bước này bắt được require nào còn sót ở 13.2.
- Pull các image `v0.0.1`. Khởi động chúng bằng
  [deploy/docker/docker-compose.yaml](../../deploy/docker/docker-compose.yaml)
  sau khi đổi các tag `:latest` thành `:v0.0.1`, rồi chạy lại các smoke test
  ở 14.3 (`INS`, `EDG`).
- Thêm mục `## [Unreleased]` mới vào changelog, và đặt `core.Version` thành
  pre-release kế tiếp (`v0.0.2-beta`).
