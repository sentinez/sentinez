# 2. Bắt đầu phát triển

## 2.1 Yêu cầu

- Go 1.27+ (`go.mod` đặt `go 1.27.1`)
- Node 26 + pnpm (cho `ui/` và sinh TypeScript từ proto)
- `protoc` / `buf`, các plugin trong
  [tools/tools.go](../../staging/src/github.com/sentinez/tools/tools.go)
  (`protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`,
  `protoc-gen-openapiv2`, `protoc-gen-go-senz`, ...)
- `make`, Docker (PostgreSQL, Consul)
- Với dataplane: Linux có quyền root, clang/llvm để biên dịch eBPF
- Với edge chạy local: `mkcert` để tạo chứng chỉ

```sh
git clone https://github.com/sentinez/sentinez.git \
  $GOPATH/src/github.com/sentinez/sentinez
git submodule update --init --recursive
```

Script sinh code proto (`api/proto/**/generate.sh`) yêu cầu repo nằm dưới
`$GOPATH/src/github.com/sentinez/sentinez`.

## 2.2 Build & chạy

```sh
make                       # build tất cả: edge, dataplane, apiserver, realtime, greeter, centraldata
make sz.<svc>.build        # apiserver | realtime | edge | dataplane | greeter | centraldata
make sz.<svc>.run          # build + chạy với --env_file=cmd/<svc>/.env
make sz.<svc>.image.build  # docker image: apiserver | edge | greeter | centraldata
```

Binary được đặt ở `cmd/<svc>/bin/` (hoặc `cmd/<svc>/v1/bin/`). Đổi tên binary
bằng `SENTINEZ_OUT=...`, đổi tag image bằng `TAG=...`.

### Hạ tầng phụ trợ

```sh
make compose.up            # postgres:16 (5432), consul (8500), apiserver, edge
bash hack/timescale-up.sh  # chỉ bật container sentinez.postgres
```

### API server

```sh
cp cmd/szapiserver/.env.example cmd/szapiserver/.env   # chỉnh lại giá trị
make sz.apiserver.run
# REST:    http://localhost:8080/...
# Swagger: http://localhost:8080/swagger/
```

Khi khởi động, các repository tự tạo bảng (`CREATE TABLE IF NOT EXISTS` +
`ALTER TABLE ... ADD COLUMN IF NOT EXISTS`), không cần migration riêng.

### Edge proxy

```sh
./hack/mkcert-local.sh     # tạo cert *.sentinez.vn (xem lưu ý dưới)
# /etc/hosts: 127.0.0.1 example.sentinez.vn sentinez.vn
make sz.edge.run
```

`make sz.edge.run` truyền `--cert_file`, `--cert_key_file`,
`--proxy_config=./cmd/szedge/v1/proxy.yaml`, `--env_file`, `--engine=quic`.
Edge yêu cầu
`SENZ_HOSTNAME` khớp domain gốc: request tới `badcheese.<SENZ_HOSTNAME>` sẽ
dùng cấu hình `server.name: badcheese`. Edge còn mở pprof ở `:6060`.

> `hack/mkcert-local.sh` đang `mv *.pem ./cmd/edge/v1/` (đường dẫn cũ). Hãy
> chuyển file sang `cmd/szedge/v1/` hoặc truyền flag cert thủ công.

### Dataplane (eBPF)

```sh
sudo make -f hack/net/dev/Makefile setup   # tạo netns client/gateway, veth, NAT
make sz.dataplane.run                      # sudo ip netns exec gateway ...
```

Xem [05-dataplane-ebpf.md](05-dataplane-ebpf.md).

### UI

```sh
cd ui && pnpm install && pnpm dev   # turbo dev cho console/web
```

Console đọc `NEXT_PUBLIC_BASE_PATH` (mặc định `http://localhost:8080`).

## 2.3 Biến môi trường

Tên biến là các giá trị của enum `Senz` trong
[types/setting/v1/config.proto](../../api/proto/sentinez/types/setting/v1/config.proto).
Đọc qua `shared/config.LoadEnv` (dùng `godotenv`), truy cập bằng
`conf.Get(settingpb.Senz_...)` hoặc `config.GetEnv(...)`.

| Biến | Dùng ở | Ý nghĩa |
|---|---|---|
| `SENZ_HOSTNAME` | edge, IAM | Domain gốc của edge (`sentinez.vn`); RP ID cho WebAuthn |
| `SENZ_ADDRESS` | tất cả | Địa chỉ listen. Edge mặc định `0.0.0.0:443` và lắng nghe cả TCP lẫn UDP (HTTP/3) trên địa chỉ đó, apiserver mặc định `:9000` nếu rỗng |
| `SENZ_SECRET_KEY` | IAM | Khoá HMAC ký JWT, **phải là base64** |
| `SENZ_CLIENT_ORIGIN` | IAM | Origin được phép cho WebAuthn (URL console) |
| `SENZ_POSTGRES_URI` | controlplane | DSN PostgreSQL (pgxpool) |
| `SENZ_TIMESCALE_URI`, `SENZ_CLICKHOUSE_URI` | — | Khai báo nhưng chưa dùng |
| `SENZ_CONSUL_URI` | gateway, gRPC | Consul cho service discovery |
| `SENZ_MEMBERSHIP_ADDRESS` | edge | Peer Olric để join cluster |
| `SENZ_DISCOVERY_ADDRESS` | edge | `host:port` bind memberlist (mặc định `0.0.0.0:3322`) |
| `SENZ_ADMIN_USERNAME` / `SENZ_ADMIN_PASSWORD` | IAM | Tài khoản admin tĩnh |
| `SENZ_OTLP_ENDPOINT` | tất cả | Collector OTLP/gRPC (mặc định `localhost:4317`) |

Một số service (edge, dataplane, realtime) chỉ nạp danh sách biến cụ thể; các
service khác nạp toàn bộ enum `Senz`.

## 2.4 Command-line flags

Định nghĩa trong
[types/setting/v1/flags.proto](../../api/proto/sentinez/types/setting/v1/flags.proto),
parse bằng `pflag` trong `shared/flagx` và `pkg/apps/**/flags`. Giá trị được
validate bằng protovalidate.

| Flag | Mặc định | Ràng buộc | Service |
|---|---|---|---|
| `--mode`, `-m` | `dev` | `dev\|prod\|sandbox` | tất cả |
| `--log_level` | `debug` | `debug\|info\|warn\|error` | tất cả |
| `--env_file` | tuỳ service | — | tất cả (`""` = không đọc file) |
| `--proxy_config` | `./proxy.yaml` | — | edge |
| `--cert_file` | `./_wildcard.sentinez.vn+1.pem` | — | edge |
| `--cert_key_file` | `./_wildcard.sentinez.vn+1-key.pem` | — | edge |
| `--engine` | `quic` | `std\|hertz\|quic` | edge (HTTP server, xem [03-edge.md](03-edge.md#310-chọn-engine)) |

`--mode` còn quyết định tiền tố tên bảng DB (`dev_sentinez_iam_users`, ...)
và bật log của fx khi bằng `dev`.

## 2.5 Kiểm thử & lint

```sh
make test.cover            # go test ./... -cover (CHỈ module root)
cd staging/src/github.com/sentinez/core && go test ./...    # từng module riêng
make lint                  # golangci-lint cho root, core, shared, controlplane, httphz, tools + buf lint
make fmt.proto             # buf format -w
```

Test hiện có tập trung ở: `core/limiter`, `core/rules`, `core/modsec/ruleparser`,
`controlplane/iam` (handler, service với mock), `controlplane/pkg/crypto`,
`shared/{config,errorx,perms,protobuf,zlog}`, `pkg/apps/dmz/edge`,
`pkg/apps/gateway/apiserver`.

Quy tắc lint (`.golangci.yaml`): `lll` 80 cột (tab = 4), `funlen` 35 dòng,
`errcheck`, `govet`, `misspell`, `revive`, `staticcheck`, `gofmt`,
`goimports`. Style guide đầy đủ ở
[.agent/skills/go-style-guide/SKILL.md](../../.agent/skills/go-style-guide/SKILL.md).

## 2.6 Sinh code

| Việc | Lệnh |
|---|---|
| Go + gateway + validate + senz + OpenAPI | `cd api/proto/sentinez/<domain>/v1 && ./generate.sh` |
| TypeScript cho UI | `cd api && buf generate` (plugin `ts-proto` → `ui/packages/proto/ts-proto`) |
| Template trang lỗi | `cd staging/.../core/common/render && templ generate` |
| Rule CRS → Go | `bash hack/ruleparser.sh` (cần `ruleparser-sentinez` trong PATH) |
| eBPF → Go | `cd staging/.../bpf/sentinez && go generate .` |
| Copy proto bên thứ ba | `bash hack/copy_proto3pt.sh` |
| Copy file `.data` của CRS | `bash hack/copy_rule_data.sh` |
| Mock | `mockery` |
