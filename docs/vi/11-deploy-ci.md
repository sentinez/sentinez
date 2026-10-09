# 11. Triển khai, CI & phát hành module

## 11.1 Docker image

Dockerfile nằm trong từng `cmd/<svc>`; build từ gốc repo:

```sh
make sz.apiserver.image.build    # TAG mặc định sentinez/apiserver
make sz.edge.image.build         # sentinez/edge
make sz.greeter.image.build      # sentinez/greeter
make sz.centraldata.image.build  # sentinez/centraldata
```

Mẫu chung: stage `golang:1.27.1-alpine` (CGO tắt, `go mod tidy`, `go build
-ldflags="-s -w"`) → stage `alpine` có `ca-certificates`, `tzdata`.

| Image | Lệnh chạy |
|---|---|
| apiserver | `docker-entrypoint.sh`: `--log_level info --mode prod --env_file ""` (biến môi trường lấy từ container) |
| edge | `docker-entrypoint.sh`: thêm `--proxy_config ./proxy.yaml`, `--engine quic`, cert ở `/etc/senz/ssl/{certs,private}`. Image **đóng gói sẵn** `proxy.yaml` và cặp cert/key trong `cmd/szedge/v1` |
| greeter, centraldata | Chạy `/bin/main` |

## 11.2 Docker Compose

[deploy/docker/docker-compose.yaml](../../deploy/docker/docker-compose.yaml)
(`make compose.up` / `compose.down`):

| Service | Image | Port | Ghi chú |
|---|---|---|---|
| `sentinez.postgres` | `postgres:16` | 5432 | user/pass `root/root`, DB `sentinez`, volume `_volume/postgres` |
| `sentinez.consul` | `hashicorp/consul` | 8500, 8600/udp | `agent -dev` |
| `sentinez.gateway.apiserver` | `sentinez/apiserver` | 8080 | `env_file: cmd/szapiserver/.env`, ghi đè `SENZ_POSTGRES_URI` |
| `sentinez.dmz.edge` | `sentinez/edge` | 7443/tcp, 7443/udp (HTTP/3), 6060 (pprof) | `env_file: cmd/szedge/v1/.env` |

Redis, TimescaleDB, NATS có sẵn nhưng đang comment. Ngoài ra còn
`deploy/docker/{cluster,mesh}/docker-compose.yaml` cho các kịch bản nhiều node.

## 11.3 Các cấu hình khác trong `deploy/`

| Thư mục | Nội dung |
|---|---|
| `caddy/` | `Caddyfile`: `localhost { reverse_proxy localhost:7777 }` + compose |
| `nginx/` | `nginx.conf` + `conf/` |
| `monitoring/` | Prometheus (`prometheus/prometheus.yml`), Grafana, dashboard `sentinez-edge-cluster-grafana.json` |
| `configs/{dev,prod,sandbox}` | Cấu hình theo môi trường |
| `ruleroot/` | CRS: `setup.conf`, `default.conf`, `audit.conf`, `block.conf`, `modsecurity.conf`, `REQUEST-901/949`, thư mục `v4-16-0`, `v4-17-0` (`*.conf` + `*.data`) |

## 11.4 Phát hành module staging

Các module trong `staging/src/github.com/sentinez/<name>` được đồng bộ sang
`github.com/sentinez/<name>`:

- Danh sách đích:
  [staging/publishing/rules.yaml](../../staging/publishing/rules.yaml)
  (`core`, `shared`, `tools`, `contrib`, `controlplane`, `bpf`) và nhánh
  (`staging`, `main`).
- Workflow [ci_sync_repo.yaml](../../.github/workflows/ci_sync_repo.yaml) được
  **sinh ra** bởi `bash staging/publishing/gen_sync_workflow.sh` — sửa
  `rules.yaml` rồi chạy lại script, không sửa tay workflow.
- Mỗi job chạy [hack/ci/action_sync.sh](../../hack/ci/action_sync.sh) `<name>`:
  `git subtree split --prefix=staging/src/github.com/sentinez/<name>`, merge
  lịch sử remote bằng chiến lược `ours`, rồi `git push --force` lên nhánh cùng
  tên. Cần secret `SENZ_GITHUB_TOKEN`.

Vì module root dùng `replace` trỏ vào `staging/`, mọi thay đổi trong staging
có hiệu lực ngay ở root; sau khi đổi dependency của một module staging hãy
chạy `go mod tidy` trong module đó **và** ở root.

## 11.5 GitHub Actions

| Workflow | Kích hoạt | Việc làm |
|---|---|---|
| `ci_lint.yaml` | push `main`, `release/*`; PR | `golangci-lint v2.13.2` qua `hack/golint.sh` (`go mod tidy && golangci-lint run` ở root); `buf lint` qua `hack/protolint.sh` |
| `ci_test.yaml` | push `main`, `release/*`; PR | `go mod download && go test ./...` (chỉ module root) |
| `ci_sync_repo.yaml` | push `staging`, `main` | Đồng bộ module staging |

Dependabot cập nhật `gomod` (root) và `npm` (`ui/apps/console`) hằng ngày.

Lưu ý: CI lint/test chỉ chạy ở module root; lint các module staging phải chạy
local bằng `make lint`.

## 11.6 Mirror GitLab

```sh
bash hack/gitremote.sh    # thêm remote gitlab
bash hack/gitlab_sync.sh  # push --all và --tags lên gitlab
```

## 11.7 Benchmark & profiling

- `hack/benchmark/greeter_say.sh`: `bombardier -c 100 -n 10000` vào edge.
- `hack/benchmark/gobench.sh`: `go test -bench . -benchmem` kèm CPU/mem
  profile.
- Edge luôn mở `net/http/pprof` ở `:6060`:
  `go tool pprof http://localhost:6060/debug/pprof/profile`.
