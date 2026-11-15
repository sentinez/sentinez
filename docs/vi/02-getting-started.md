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
sudo tee -a /etc/hosts < deploy/hosts  # một lần: các domain mà config dùng
bash hack/timescale-up.sh  # chỉ bật sentinez.timescale (TimescaleDB) và sentinez.consul
```

### API server

```sh
cp cmd/szapiserver/.env.example cmd/szapiserver/.env   # chỉnh lại giá trị
make sz.apiserver.run
# REST:    http://sentinez.gateway.apiserver:8080/...
# Swagger: http://sentinez.gateway.apiserver:8080/swagger/
```

Khi khởi động, các repository tự tạo bảng (`CREATE TABLE IF NOT EXISTS` +
`ALTER TABLE ... ADD COLUMN IF NOT EXISTS`), không cần migration riêng.

### Edge proxy

```sh
./hack/mkcert-local.sh     # tạo cert *.sentinez.vn (xem lưu ý dưới)
# /etc/hosts: the names of deploy/hosts (sentinez.vn, example.sentinez.vn)
make sz.edge.run
```

`make sz.edge.run` truyền `--cert_file`, `--cert_key_file`,
`--proxy_config=./cmd/szedge/v1/proxy.yaml`, `--env_file`, `--engine=quic`.
Request được phục vụ
bằng cấu hình có `server.resourceDomain` trùng host của nó (bỏ port): request
tới `badcheese.is.s6z.io.vn` dùng `resourceDomain: badcheese.is.s6z.io.vn`. Edge còn mở pprof ở `:6060`.

> `hack/mkcert-local.sh` đặt cert vào `cmd/szedge/v1/`; truyền chúng qua
> `--cert_file`/`--cert_key_file`.

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

Console gọi `/api`, server Next.js chuyển tiếp tới `API_BASE_PATH`
(chỉ có ở server, mặc định `http://sentinez.gateway.apiserver:8080`).

## 2.3 Biến môi trường

Tên biến là các giá trị của enum `Senz` trong
[types/setting/v1/config.proto](../../api/proto/sentinez/types/setting/v1/config.proto).
Đọc qua `pkg/szconfig.LoadEnv` (dùng `godotenv`), truy cập bằng
`conf.Get(settingpb.Senz_...)` hoặc `config.GetEnv(...)`.

| Biến | Dùng ở | Ý nghĩa |
|---|---|---|
| `SENZ_HOSTNAME` | IAM | RP ID cho WebAuthn (`sentinez.vn`) |
| `SENZ_ADDRESS` | tất cả | Địa chỉ listen; nếu rỗng mỗi service dùng hằng trong [internal/defaults](../../internal/defaults/defaults.go): edge `0.0.0.0:443` (TCP và UDP/HTTP/3), apiserver `0.0.0.0:9000`, realtime `0.0.0.0:7778`, dataplane và greeter `0.0.0.0:9001` |
| `SENZ_SECRET_KEY` | IAM | Khoá HMAC ký JWT, **phải là base64** |
| `SENZ_DATA_KEY` | apiserver | Khoá AES-256 (32 byte ngẫu nhiên, **base64**, `openssl rand -base64 32`) dùng để mã hoá PEM và private key của CA certificate trong database. Bắt buộc: apiserver không khởi động nếu thiếu. Mất hoặc đổi khoá thì các certificate đã lưu không đọc lại được |
| `SENZ_CLIENT_ORIGIN` | IAM, apiserver | Origin của console: origin WebAuthn được phép và là origin duy nhất CORS cho phép. Nhiều giá trị: cách nhau bằng dấu phẩy |
| `SENZ_POSTGRES_URI` | controlplane | DSN PostgreSQL (pgxpool) |
| `SENZ_TIMESCALE_URI`, `SENZ_CLICKHOUSE_URI` | — | Khai báo nhưng chưa dùng |
| `SENZ_CONSUL_URI` | gateway, gRPC, edge | Consul cho service discovery; ở edge dùng để bầu leader theo zone ([3.8](03-edge.md#bầu-leader-theo-zone)), bắt buộc khi edge chạy cụm |
| `SENZ_MEMBERSHIP_ADDRESS` | edge | Peer Olric để join cluster; đi cùng `SENZ_DISCOVERY_ADDRESS` thì edge được tính là chạy cụm |
| `SENZ_DISCOVERY_ADDRESS` | edge | `host:port` bind memberlist (mặc định `0.0.0.0:3322`) |
| `SENZ_ADMIN_USERNAME` / `SENZ_ADMIN_PASSWORD` | IAM | Tài khoản admin tĩnh |
| `SENZ_OTLP_INTERNAL_ENDPOINT` | tất cả | Collector OTLP/gRPC cho log hệ thống; không đặt thì không push |
| `SENZ_OTLP_EXTERNAL_ENDPOINT` | edge | Collector OTLP/gRPC cho log traffic của edge (node `LOG` trong func chains); không đặt thì không push |
| `SENZ_KAFKA_BROKERS` | edge | Kafka broker (phân tách bằng dấu phẩy) mà edge đẩy thẳng event traffic tới, dạng proto thô (`pkg/szkafka`); được ưu tiên hơn `SENZ_OTLP_EXTERNAL_ENDPOINT` |
| `SENZ_ZONE` | edge | Zone nơi node được deploy (vd `ap-southeast-1a`); làm tiền tố cho mọi id sinh từ `szrand` `NewULID`/`NewNanoId` (id request và event của handler chain: `<zone>-req_<ulid>`, `<zone>-srevt_<ulid>`, `<zone>-crevt_<ulid>`; không đặt: `req_<ulid>`). Primary key (`NewUUID`) không bao giờ gắn zone. Cũng là zone mà edge tranh quyền leader; bắt buộc khi có `SENZ_CONSUL_URI` |

Một số service (edge, dataplane, realtime) chỉ nạp danh sách biến cụ thể; các
service khác nạp toàn bộ enum `Senz`.

### Config từ Consul

Khi có `SENZ_CONSUL_URI`, mọi service đọc thêm biến của mình từ Consul KV, một
lần lúc khởi động (`pkg/appconfig`; service chỉ khai báo flags và danh sách key
cần đọc trong `config/config.go`).

| Key trên Consul | Dành cho |
|---|---|
| `sentinez/env/SENZ_KEY` | mọi service |
| `sentinez/env/<service_key>/SENZ_KEY` | một service, ví dụ `sentinez/env/sentinez.dmz.edge.v1/SENZ_KAFKA_BROKERS` (`service_key` nằm trong proto của service, `x_meta`) |

- **Thứ tự ưu tiên:** ENV (gồm cả `.env`) thắng, rồi tới key của service, rồi
  tới key chung. Giá trị rỗng coi như chưa đặt.
- **Chỉ ENV mới đặt được** `SENZ_CONSUL_URI`, `SENZ_ADDRESS`,
  `SENZ_MEMBERSHIP_ADDRESS`, `SENZ_DISCOVERY_ADDRESS` và `SENZ_ZONE`: đó là cách
  tới Consul, hoặc giá trị khác nhau theo từng instance. Key như vậy trên Consul
  bị bỏ qua kèm cảnh báo, key lạ cũng vậy.
- **Khi lỗi:** có `SENZ_CONSUL_URI` mà không tới được Consul thì service không
  khởi động. Không có URI thì mọi thứ như cũ.
- **Bí mật:** `SENZ_SECRET_KEY`, `SENZ_DATA_KEY`, `SENZ_ADMIN_PASSWORD` và
  `SENZ_POSTGRES_URI` có thể để trên Consul, nhưng KV của Consul không mã hoá:
  hãy bảo vệ bằng ACL của Consul. Đổi giá trị thì phải restart service.
- **Edge:** đọc `proxy.yaml` từ key `sentinez/proxy.yaml` trên Consul (chính key
  apiserver import, `make consul.proxy.put`); không có `SENZ_CONSUL_URI`, hoặc
  thiếu key, thì đọc file của `--proxy_config`.

Phần ENV dùng chung cho mọi service nằm ở
[deploy/consul/shared.env.example](../../deploy/consul/shared.env.example); phần
khác nhau giữa các instance (`SENZ_ADDRESS`, `SENZ_CONSUL_URI`, và với edge là
`SENZ_ZONE`, `SENZ_MEMBERSHIP_ADDRESS`, `SENZ_DISCOVERY_ADDRESS`) vẫn để trong
`cmd/<service>/.env` (xem `.env.example` của service).

```sh
cp deploy/consul/shared.env.example deploy/consul/shared.env   # git bỏ qua file này; chỉnh giá trị
make consul.env.put                                            # một lần, cho mọi service
make consul.env.put ENV_FILE=cmd/szedge/v1/.env SERVICE=sentinez.dmz.edge.v1  # một service
```

`hack/consul-env-put.sh` đẩy các dòng `SENZ_*` của file lên Consul (không in
giá trị) và bỏ qua các key chỉ ENV mới được đặt. Sau đó xoá các key đã chuyển
khỏi file `.env`.

## 2.4 Command-line flags

Định nghĩa trong
[types/setting/v1/flags.proto](../../api/proto/sentinez/types/setting/v1/flags.proto),
parse bằng `pflag` trong `pkg/szflag` và `pkg/apps/**/flags`. Giá trị được
validate bằng protovalidate.

| Flag | Mặc định | Ràng buộc | Service |
|---|---|---|---|
| `--mode`, `-m` | `dev` | `dev\|prod\|sandbox` | tất cả |
| `--log_level` | `debug` | `debug\|info\|warn\|error` | tất cả |
| `--env_file` | tuỳ service | — | tất cả (`""` = không đọc file) |
| `--proxy_config` | `./proxy.yaml` | — | edge |
| `--cert_file` | `./_wildcard.sentinez.vn+1.pem` (edge), `""` (apiserver) | — | edge, apiserver (có certificate thì apiserver phục vụ HTTPS và HTTP/3) |
| `--cert_key_file` | `./_wildcard.sentinez.vn+1-key.pem` (edge), `""` (apiserver) | — | edge, apiserver |
| `--engine` | `quic` | `std\|hertz\|quic` | edge (HTTP server, xem [03-edge.md](03-edge.md#310-chọn-engine)) |

`--mode` còn quyết định tiền tố tên bảng DB (`dev_sentinez_iam_users`, ...)
và bật log của fx khi bằng `dev`.

## 2.5 Kiểm thử & lint

```sh
make test.cover            # go test ./... -cover (CHỈ module root)
cd staging/src/github.com/sentinez/core && go test ./...    # từng module riêng
make lint                  # golangci-lint cho root, core, pkg, controlplane, httphz, tools + buf lint
make fmt.proto             # buf format -w
```

Test hiện có tập trung ở: `core/szlimiter`, `core/szrule`, `core/szmodsec/ruleparser`,
`controlplane/iam` (handler, service với mock), `controlplane/pkg/crypto`,
`pkg/{szconfig,szerror,szperms,szprotobuf,szlog}`, `pkg/apps/dmz/edge`,
`pkg/apps/gateway/apiserver`.

Quy tắc lint (`.golangci.yaml`): `lll` 80 cột (tab = 4), `funlen` 35 dòng,
`errcheck`, `govet`, `misspell`, `revive`, `staticcheck`, `gofmt`,
`goimports`. Style guide đầy đủ ở
[hack/.agent/skills/go-style-guide/SKILL.md](../../hack/.agent/skills/go-style-guide/SKILL.md).

## 2.6 Sinh code

| Việc | Lệnh |
|---|---|
| Go + gateway + validate + senz + OpenAPI | `cd api/proto/sentinez/<domain>/v1 && ./generate.sh` |
| TypeScript cho UI | `cd api && buf generate` (plugin `ts-proto` → `ui/packages/proto/ts-proto`) |
| Template trang lỗi | `cd staging/.../core/internal/render && templ generate` |
| Rule CRS → Go | `bash hack/ruleparser.sh` (cần `szruleparser` trong PATH) |
| eBPF → Go | `cd staging/.../bpf/sentinez && go generate .` |
| Copy proto bên thứ ba | `bash hack/copy_proto3pt.sh` |
| Copy file `.data` của CRS | `bash hack/copy_rule_data.sh` |
| Mock | `mockery` |
