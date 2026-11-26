# 2. Bắt đầu phát triển

## 2.1 Yêu cầu

- Go 1.27+ (`go.mod` đặt `go 1.27.1`)
- Node 26 + pnpm (cho `ui/` và sinh TypeScript từ proto)
- `protoc` / `buf`, các plugin trong
  [tools/tools.go](../../staging/src/github.com/sentinez/tools/tools.go)
  (`protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`,
  `protoc-gen-openapiv2`, `protoc-gen-go-sz`, ...)
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
bằng cấu hình có `ingress.resource.domain` trùng host của nó (bỏ port): request
tới `badcheese.is.s6z.io.vn` dùng `domain: badcheese.is.s6z.io.vn`. Edge còn mở pprof ở `:6060`.

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

Tên biến là các giá trị của enum `Sz` trong
[types/setting/v1/config.proto](../../api/proto/sentinez/types/setting/v1/config.proto).
Đọc qua `pkg/szconfig.LoadEnv` (dùng `godotenv`), truy cập bằng
`conf.Get(settingpb.Sz_...)` hoặc `config.GetEnv(...)`.

| Biến | Dùng ở | Ý nghĩa |
|---|---|---|
| `SZ_HOSTNAME` | IAM | RP ID cho WebAuthn (`sentinez.vn`) |
| `SZ_ADDRESS` | tất cả | Địa chỉ listen; nếu rỗng mỗi service dùng hằng trong [internal/defaults](../../internal/defaults/defaults.go): edge `0.0.0.0:443` (TCP và UDP/HTTP/3), apiserver `0.0.0.0:9000`, realtime `0.0.0.0:7778`, dataplane và greeter `0.0.0.0:9001` |
| `SZ_SECRET_KEY` | IAM | Khoá HMAC ký JWT, **phải là base64** |
| `SZ_DATA_KEY` | apiserver | Khoá AES-256 (32 byte ngẫu nhiên, **base64**, `openssl rand -base64 32`) dùng để mã hoá PEM và private key của CA certificate trong database. Bắt buộc: apiserver không khởi động nếu thiếu. Mất hoặc đổi khoá thì các certificate đã lưu không đọc lại được |
| `SZ_CLIENT_ORIGIN` | IAM, apiserver | Origin của console: origin WebAuthn được phép và là origin duy nhất CORS cho phép. Nhiều giá trị: cách nhau bằng dấu phẩy |
| `SZ_POSTGRES_URI` | controlplane | DSN PostgreSQL (pgxpool) |
| `SZ_TIMESCALE_URI`, `SZ_CLICKHOUSE_URI` | — | Khai báo nhưng chưa dùng |
| `SZ_CONSUL_URI` | gateway, gRPC, edge | Consul cho service discovery; ở edge dùng để bầu leader theo zone ([3.8](03-edge.md#bầu-leader-theo-zone)), bắt buộc khi edge chạy cụm |
| `SZ_MEMBERSHIP_ADDRESS` | edge | Peer Olric để join cluster; đi cùng `SZ_DISCOVERY_ADDRESS` thì edge được tính là chạy cụm |
| `SZ_DISCOVERY_ADDRESS` | edge | `host:port` bind memberlist (mặc định `0.0.0.0:3322`) |
| `SZ_ADMIN_USERNAME` / `SZ_ADMIN_PASSWORD` | IAM | Tài khoản admin tĩnh |
| `SZ_OTLP_INTERNAL_ENDPOINT` | tất cả | Collector OTLP/gRPC cho log hệ thống; không đặt thì không push |
| `SZ_OTLP_EXTERNAL_ENDPOINT` | edge | Collector OTLP/gRPC cho log traffic của edge (node `logging` trong func chains); không đặt thì không push |
| `SZ_KAFKA_BROKERS` | edge | Kafka broker (phân tách bằng dấu phẩy) mà edge đẩy thẳng event traffic tới, dạng proto thô (`pkg/szkafka`); được ưu tiên hơn `SZ_OTLP_EXTERNAL_ENDPOINT` |
| `SZ_ZONE` | edge | Zone nơi node được deploy (vd `ap-southeast-1a`); làm tiền tố cho mọi id sinh từ `szrand` `NewULID`/`NewNanoId` (id request và event của handler chain: `<zone>-req_<ulid>`, `<zone>-srevt_<ulid>`, `<zone>-rsevt_<ulid>`; không đặt: `req_<ulid>`). Primary key (`NewUUID`) không bao giờ gắn zone. Cũng là zone mà edge tranh quyền leader; bắt buộc khi có `SZ_CONSUL_URI` |
| `SZ_SERVICE_KEY` | centraldata, `szservicetoken` | Khoá HMAC (base64) ký và kiểm tra token service. Chỉ nơi kiểm tra và ký token giữ, edge không bao giờ giữ; đừng đặt vào shared env |
| `SZ_SERVICE_TOKEN` | edge | Token service edge trình cho centraldata (`szservicetoken -name <CN>`); có giá trị thì edge leader theo dõi centraldata |
| `SZ_TLS_CA`, `SZ_TLS_CERT`, `SZ_TLS_KEY` | centraldata, edge | File PEM của CA nội bộ, certificate và key của service, cho mTLS edge → centraldata ([7.6](07-controlplane.md#thiết-lập-mtls-nội-bộ-edge--centraldata)). Đủ cả ba hoặc không có cái nào: để trống là plaintext (chỉ dùng khi phát triển) |
| `SZ_OLRIC_ADDRESSES` | centraldata | Địa chỉ Olric (cách nhau bằng dấu phẩy) của cluster edge mà centraldata publish thay đổi tới |
| `SZ_SYNC_TABLES` | centraldata | Các bảng (cách nhau bằng dấu phẩy) centraldata đồng bộ tới edge; để trống thì dùng bộ mặc định |
| `SZ_SYNC_KEY` | centraldata, edge | Khoá HMAC (base64, tối thiểu 32 byte) ký các thay đổi centraldata publish; edge kiểm tra bằng cùng giá trị |

Một số service (edge, dataplane, realtime) chỉ nạp danh sách biến cụ thể; các
service khác nạp toàn bộ enum `Sz`.

### Biến của từng service

Mỗi service đọc gì, từ environment (`cmd/<service>/.env`, hoặc `--env-file`)
hoặc từ Consul. Các biến này không có flag; flag của service chỉ chỉ ra file
env và, với edge và apiserver, proxy config và certificate. `SZ_CONSUL_URI`
luôn được đọc.

| Service | Đọc |
|---|---|
| apiserver (`szapiserver`) | `SZ_ADDRESS`, `SZ_CONSUL_URI`, `SZ_POSTGRES_URI`, `SZ_HOSTNAME`, `SZ_CLIENT_ORIGIN`, `SZ_SECRET_KEY`, `SZ_DATA_KEY`, `SZ_ADMIN_USERNAME`, `SZ_ADMIN_PASSWORD`, `SZ_OTLP_INTERNAL_ENDPOINT` |
| realtime (`szrealtime`) | `SZ_ADDRESS`, `SZ_HOSTNAME`, `SZ_SECRET_KEY`, `SZ_OTLP_INTERNAL_ENDPOINT` |
| edge (`szedge`) | `SZ_ADDRESS`, `SZ_CONSUL_URI`, `SZ_ZONE`, `SZ_MEMBERSHIP_ADDRESS`, `SZ_DISCOVERY_ADDRESS`, `SZ_SECRET_KEY`, `SZ_DATA_KEY`, `SZ_KAFKA_BROKERS`, `SZ_OTLP_INTERNAL_ENDPOINT`, `SZ_OTLP_EXTERNAL_ENDPOINT`, `SZ_SERVICE_TOKEN`, `SZ_SYNC_KEY`, `SZ_TLS_CA`, `SZ_TLS_CERT`, `SZ_TLS_KEY` |
| dataplane (`szdataplane`) | `SZ_ADDRESS`, `SZ_HOSTNAME`, `SZ_SECRET_KEY`, `SZ_OTLP_INTERNAL_ENDPOINT` |
| centraldata (`szcentraldata`) | `SZ_ADDRESS`, `SZ_CONSUL_URI`, `SZ_POSTGRES_URI`, `SZ_SECRET_KEY`, `SZ_SERVICE_KEY`, `SZ_SYNC_KEY`, `SZ_SYNC_TABLES`, `SZ_OLRIC_ADDRESSES`, `SZ_TLS_CA`, `SZ_TLS_CERT`, `SZ_TLS_KEY` |
| collector (`szcollector`) | `SZ_ADDRESS`, `SZ_CONSUL_URI`, `SZ_KAFKA_BROKERS`, `SZ_POSTGRES_URI` |
| greeter (`szgreeter`) | `SZ_ADDRESS`, `SZ_CONSUL_URI` |

Bắt buộc, nếu không service không khởi động: `SZ_DATA_KEY` (apiserver), và
`SZ_ZONE` cùng `SZ_CONSUL_URI` (edge). Centraldata không có
`SZ_OLRIC_ADDRESSES` thì không publish gì; có thì `SZ_SYNC_KEY` là bắt buộc.

Không đặt vào shared env mà để trong `.env` của đúng service đó:
`SZ_SERVICE_KEY` (centraldata), `SZ_SERVICE_TOKEN` và các file `SZ_TLS_*`
(mỗi service có certificate riêng).

### Config từ Consul

Khi có `SZ_CONSUL_URI`, mọi service đọc thêm biến của mình từ Consul KV, một
lần lúc khởi động (`pkg/appconfig`; service chỉ khai báo flags và danh sách key
cần đọc trong `config/config.go`).

| Key trên Consul | Dành cho |
|---|---|
| `sentinez/env/SZ_KEY` | mọi service |
| `sentinez/env/<service_key>/SZ_KEY` | một service, ví dụ `sentinez/env/sentinez.dmz.edge.v1/SZ_KAFKA_BROKERS` (`service_key` nằm trong proto của service, `x_meta`) |

- **Thứ tự ưu tiên:** ENV (gồm cả `.env`) thắng, rồi tới key của service, rồi
  tới key chung. Giá trị rỗng coi như chưa đặt.
- **Chỉ ENV mới đặt được** `SZ_CONSUL_URI`, `SZ_ADDRESS`,
  `SZ_MEMBERSHIP_ADDRESS`, `SZ_DISCOVERY_ADDRESS` và `SZ_ZONE`: đó là cách
  tới Consul, hoặc giá trị khác nhau theo từng instance. Key như vậy trên Consul
  bị bỏ qua kèm cảnh báo, key lạ cũng vậy.
- **Khi lỗi:** có `SZ_CONSUL_URI` mà không tới được Consul thì service không
  khởi động. Không có URI thì mọi thứ như cũ.
- **Bí mật:** `SZ_SECRET_KEY`, `SZ_DATA_KEY`, `SZ_ADMIN_PASSWORD` và
  `SZ_POSTGRES_URI` có thể để trên Consul, nhưng KV của Consul không mã hoá:
  hãy bảo vệ bằng ACL của Consul. Đổi giá trị thì phải restart service.
- **Edge:** đọc `proxy.yaml` từ key `sentinez/proxy.yaml` trên Consul (chính key
  apiserver import, `make consul.proxy.put`); không có `SZ_CONSUL_URI`, hoặc
  thiếu key, thì đọc file của `--proxy_config`.

Phần ENV dùng chung cho mọi service nằm ở
[deploy/consul/shared.env.example](../../deploy/consul/shared.env.example); phần
khác nhau giữa các instance (`SZ_ADDRESS`, `SZ_CONSUL_URI`, và với edge là
`SZ_ZONE`, `SZ_MEMBERSHIP_ADDRESS`, `SZ_DISCOVERY_ADDRESS`) vẫn để trong
`cmd/<service>/.env` (xem `.env.example` của service).

Các key của `shared.env`:

| Key | Đọc bởi | Ý nghĩa |
|---|---|---|
| `SZ_HOSTNAME` | apiserver, realtime, dataplane, edge | Domain của console, relying party của WebAuthn (`sentinez.vn`) |
| `SZ_CLIENT_ORIGIN` | apiserver | Origin của console: origin WebAuthn được phép và là origin duy nhất CORS cho qua |
| `SZ_SECRET_KEY` | apiserver, realtime, dataplane, edge, centraldata | Khoá HMAC ký JWT, base64 |
| `SZ_DATA_KEY` | apiserver, edge | Khoá AES-256 mã hoá CA certificate trong database, base64 (`openssl rand -base64 32`). Hai service phải cùng giá trị; mất khoá thì các certificate đã lưu không đọc lại được |
| `SZ_SYNC_KEY` | centraldata, edge | Khoá HMAC (base64) ký các thay đổi centraldata publish; edge bỏ những gì không ký bằng khoá này. Chỉ hai service này dùng chung |
| `SZ_ADMIN_USERNAME`, `SZ_ADMIN_PASSWORD` | apiserver | Tài khoản admin tĩnh |
| `SZ_POSTGRES_URI` | apiserver, centraldata, collector | DSN PostgreSQL |
| `SZ_TIMESCALE_URI`, `SZ_CLICKHOUSE_URI` | không ai | Có khai báo nhưng chưa dùng |
| `SZ_KAFKA_BROKERS` | edge, collector | Kafka broker mà event traffic của edge được đẩy vào, và collector đọc từ đó |
| `SZ_OTLP_INTERNAL_ENDPOINT` | tất cả | OTLP/gRPC collector của log hệ thống |
| `SZ_OTLP_EXTERNAL_ENDPOINT` | edge | OTLP/gRPC collector của log traffic edge |

`SZ_SERVICE_KEY` cố ý không có trong file: chỉ centraldata (và người vận hành
ký token) giữ, nên đặt trong `.env` của centraldata. Mỗi service chỉ đọc các
key trong danh sách của nó ([Biến của từng service](#biến-của-từng-service));
các key khác bị bỏ qua. Đặt key vào `sentinez/env/<service_key>/` để cho một
service giá trị khác.

```sh
cp deploy/consul/shared.env.example deploy/consul/shared.env   # git bỏ qua file này; chỉnh giá trị
make consul.env.put                                            # một lần, cho mọi service
make consul.env.put ENV_FILE=cmd/szedge/v1/.env SERVICE=sentinez.dmz.edge.v1  # một service
```

`hack/consul-env-put.sh` đẩy các dòng `SZ_*` của file lên Consul (không in
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
| `--cert_file` | `""` | — | edge, apiserver (có certificate thì apiserver phục vụ HTTPS và HTTP/3; không có thì edge chỉ phục vụ certificate của centraldata) |
| `--cert_key_file` | `""` | — | edge, apiserver |
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
| Go + gateway + validate + sz + OpenAPI | `cd api/proto/sentinez/<domain>/v1 && ./generate.sh` |
| TypeScript cho UI | `cd api && buf generate` (plugin `ts-proto` → `ui/packages/proto/ts-proto`) |
| Template trang lỗi | `cd staging/.../core/internal/render && templ generate` |
| Rule CRS → Go | `bash hack/ruleparser.sh` (cần `szruleparser` trong PATH) |
| eBPF → Go | `cd staging/.../bpf/sentinez && go generate .` |
| Copy proto bên thứ ba | `bash hack/copy_proto3pt.sh` |
| Copy file `.data` của CRS | `bash hack/copy_rule_data.sh` |
| Mock | `mockery` |
