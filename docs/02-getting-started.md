# 2. Getting started

## 2.1 Requirements

- Go 1.27+ (`go.mod` sets `go 1.27.1`)
- Node 26 + pnpm (for `ui/` and for generating TypeScript from protos)
- `protoc` / `buf`, plus the plugins pinned in
  [tools/tools.go](../staging/src/github.com/sentinez/tools/tools.go)
  (`protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`,
  `protoc-gen-openapiv2`, `protoc-gen-go-sz`, ...)
- `make`, Docker (PostgreSQL, Consul)
- For the dataplane: Linux with root access, clang/llvm to compile eBPF
- For running the edge locally: `mkcert` to create certificates

```sh
git clone https://github.com/sentinez/sentinez.git \
  $GOPATH/src/github.com/sentinez/sentinez
git submodule update --init --recursive
```

The proto generation scripts (`api/proto/**/generate.sh`) require the repo to
live under `$GOPATH/src/github.com/sentinez/sentinez`.

## 2.2 Build & run

```sh
make                       # build everything: edge, dataplane, apiserver, realtime, greeter, centraldata
make sz.<svc>.build        # apiserver | realtime | edge | dataplane | greeter | centraldata
make sz.<svc>.run          # build + run with --env_file=cmd/<svc>/.env
make sz.<svc>.image.build  # docker image: apiserver | edge | greeter | centraldata
```

Binaries go to `cmd/<svc>/bin/` (or `cmd/<svc>/v1/bin/`). Rename the binary
with `SENTINEZ_OUT=...`, change the image tag with `TAG=...`.

### Supporting infrastructure

```sh
make compose.up            # postgres:16 (5432), consul (8500), apiserver, edge
sudo tee -a /etc/hosts < deploy/hosts  # once: the domain names the configs use
bash hack/timescale-up.sh  # start only sentinez.timescale (TimescaleDB) and sentinez.consul
```

### API server

```sh
cp cmd/szapiserver/.env.example cmd/szapiserver/.env   # then edit the values
make sz.apiserver.run
# REST:    http://sentinez.gateway.apiserver:8080/...
# Swagger: http://sentinez.gateway.apiserver:8080/swagger/
```

On startup the repositories create their own tables (`CREATE TABLE IF NOT
EXISTS` + `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`); there are no separate
migrations.

### Edge proxy

```sh
./hack/mkcert-local.sh     # creates a *.sentinez.vn cert (see note below)
# /etc/hosts: the names of deploy/hosts (sentinez.vn, example.sentinez.vn)
make sz.edge.run
```

`make sz.edge.run` passes `--cert_file`, `--cert_key_file`,
`--proxy_config=./cmd/szedge/v1/proxy.yaml`, `--env_file` and
`--engine=quic`. A request is served with the config whose
`ingress.resource.domain` is its host (port stripped): a request to
`badcheese.is.s6z.io.vn` uses `domain: badcheese.is.s6z.io.vn`.
The edge also opens pprof on `:6060`.

> `hack/mkcert-local.sh` puts the certs in `cmd/szedge/v1/`; pass them with
> `--cert_file`/`--cert_key_file`.

### Dataplane (eBPF)

```sh
sudo make -f hack/net/dev/Makefile setup   # create client/gateway netns, veth pairs, NAT
make sz.dataplane.run                      # sudo ip netns exec gateway ...
```

See [05-dataplane-ebpf.md](05-dataplane-ebpf.md).

### UI

```sh
cd ui && pnpm install && pnpm dev   # turbo dev for console/web
```

The console calls `/api`, which its Next.js server forwards to
`API_BASE_PATH` (server-only, default `http://sentinez.gateway.apiserver:8080`).

## 2.3 Environment variables

Variable names are the values of the `Sz` enum in
[types/setting/v1/config.proto](../api/proto/sentinez/types/setting/v1/config.proto).
They are loaded by `pkg/szconfig.LoadEnv` (using `godotenv`) and read with
`conf.Get(settingpb.Sz_...)` or `config.GetEnv(...)`.

Two forms, one per kind of variable:

- **`SZ_*_URI`, where a service connects to:**
  `scheme://[user:password@]host:port[,host:port...][/path][?query]`,
  parsed by `pkg/szuri`. The scheme picks TLS: `https`, `kafkas`, `olrics`.
  The client certificate of an mTLS is the `SZ_TLS_*` files. A value
  without a scheme (a bare `host:port`) is refused at startup.
- **`SZ_*_ADDRESS`, where a service listens:** `[host]:port`. A server
  serves TLS from the `SZ_TLS_*` files, not from its address.

| Form | Variables | TLS |
|---|---|---|
| URI | `SZ_POSTGRES_URI`, `SZ_TIMESCALE_URI` | `?sslmode=verify-full&sslrootcert=...&sslcert=...&sslkey=...` (pgx) |
| URI | `SZ_CONSUL_URI` | `https://`, with the `CONSUL_CACERT`, `CONSUL_CLIENT_CERT`, `CONSUL_CLIENT_KEY` env of the Consul client |
| URI | `SZ_OTLP_INTERNAL_URI`, `SZ_OTLP_EXTERNAL_URI` | `https://` (server TLS only) |
| URI | `SZ_KAFKA_URI` | `kafkas://`: mTLS with `SZ_TLS_*`, else TLS on the system roots |
| URI | `SZ_OLRIC_URI` | `olrics://`, to a TLS proxy in front of each node: Olric serves no TLS |
| Address | `SZ_ADDRESS`, `SZ_DISCOVERY_ADDRESS`, `SZ_MEMBERSHIP_ADDRESS`, `SZ_OLRIC_ADDRESS` | `SZ_TLS_*` (centraldata); none for memberlist and Olric |

| Variable | Used by | Meaning |
|---|---|---|
| `SZ_HOSTNAME` | IAM | WebAuthn RP ID (`sentinez.vn`) |
| `SZ_ADDRESS` | all | Listen address; when empty each service falls back to its constant in [internal/defaults](../internal/defaults/defaults.go): edge `0.0.0.0:443` (TCP and UDP/HTTP/3), apiserver `0.0.0.0:9000`, realtime `0.0.0.0:7778`, dataplane and greeter `0.0.0.0:9001` |
| `SZ_SECRET_KEY` | IAM | HMAC key for signing JWTs, **must be base64** |
| `SZ_DATA_KEY` | apiserver | AES-256 key (32 random bytes, **base64**, `openssl rand -base64 32`) that seals the PEM and the private key of the CA certificates in the database. Required: the apiserver does not start without it. Losing or changing it makes the stored certificates unreadable |
| `SZ_CLIENT_ORIGIN` | IAM, apiserver | Origin of the console: the allowed WebAuthn origin and the only one allowed by CORS. Several: separate them with commas |
| `SZ_POSTGRES_URI` | control plane | PostgreSQL URI (pgxpool), `postgres://user:password@host:5432/db?sslmode=...` |
| `SZ_TIMESCALE_URI`, `SZ_CLICKHOUSE_URI` | — | Declared but unused |
| `SZ_CONSUL_URI` | gateway, gRPC, edge | Consul agent, `http(s)://host:port` or `unix:///path`, for service discovery; on the edge, the zone leader election ([3.8](03-edge.md#zone-leader-election)), required when the edge is clustered |
| `SZ_MEMBERSHIP_ADDRESS` | edge | `host:port` memberlist peer to join; together with `SZ_DISCOVERY_ADDRESS` it makes the edge clustered |
| `SZ_DISCOVERY_ADDRESS` | edge | `host:port` memberlist bind; an unspecified host (`0.0.0.0`, `::`, empty) takes a private IP of the host, like `SZ_OLRIC_ADDRESS`; invalid, the port is `3322` |
| `SZ_ADMIN_USERNAME` / `SZ_ADMIN_PASSWORD` | IAM | Static admin account |
| `SZ_OTLP_INTERNAL_URI` | all | OTLP/gRPC collector of the system logs, `http://host:4317` (`https://` for TLS); unset: not exported |
| `SZ_OTLP_EXTERNAL_URI` | edge, apiserver | OTLP/gRPC collector of the edge traffic logs (func chain `logging` node) and of the audit log, `http(s)://host:4317`; unset: not exported |
| `SZ_KAFKA_URI` | edge, apiserver, collector | Kafka brokers the edge traffic events (and the audit log of the apiserver) are produced to directly, as raw protos (`pkg/szkafka`), and the collector reads from: `kafka://host:9092[,host:9092...]`, or `kafkas://` for TLS (mTLS with `SZ_TLS_*`, else the system roots); takes precedence over `SZ_OTLP_EXTERNAL_URI` |
| `SZ_ZONE` | edge | Zone the node is deployed in (eg `ap-southeast-1a`); prefixes every id generated by `szrand` `NewULID`/`NewNanoId` (request and handler-chain event ids: `<zone>-req_<ulid>`, `<zone>-srevt_<ulid>`, `<zone>-rsevt_<ulid>`; unset: `req_<ulid>`). Primary keys (`NewUUID`) are never zoned. Also the zone whose leader the edge campaigns for; required with `SZ_CONSUL_URI` |
| `SZ_SERVICE_KEY` | centraldata, `szservicetoken` | HMAC key (base64) that signs and verifies the service tokens. Held where tokens are verified and signed, never by the edge; never put it in the shared env |
| `SZ_SERVICE_TOKEN` | edge | Service token the edge presents to centraldata (`szservicetoken -name <CN>`); set, the leader edge follows centraldata |
| `SZ_TLS_CA`, `SZ_TLS_CERT`, `SZ_TLS_KEY` | centraldata, edge, apiserver, collector | PEM files of the internal CA and of the certificate and key of the service, for the mTLS edge → centraldata ([7.6](07-controlplane.md#setting-up-the-internal-mtls-edge--centraldata)), and the client certificate of `kafkas://` and `olrics://`. All three or none: unset is plaintext (development only) |
| `SZ_OLRIC_URI` | centraldata | Olric nodes of the edge cluster centraldata publishes the changes to, `olric://host:3320[,host:3320...]`; `olrics://` goes through a TLS proxy in front of each node (Olric serves no TLS), with the `SZ_TLS_*` certificate when set |
| `SZ_OLRIC_ADDRESS` | edge | `host:port` Olric bind, the port centraldata publishes to. Olric names the node after the IP of the host, which the other nodes dial. Unset: `127.0.0.1:3320` for a lone edge, `<private IP>:3320` for a clustered one. An unspecified host (`0.0.0.0`, `::`, empty) takes a private IP (RFC 6890) of the host too; with none the edge does not start, where Olric would have bound a public IP. A public IP set explicitly is bound, with a warning |
| `SZ_OLRIC_PASSWORD` | edge, centraldata | Password of the Olric port of the edges: Olric refuses a connection that does not `AUTH` with it, and gives it to its own client, so the nodes of a cluster and the subscriber of each edge authenticate too. Centraldata publishes with it. Same value on the edges and centraldata; sent in clear, so the port still belongs on a private network. Unset: no password, with a warning when the port is not the loopback |
| `SZ_GOSSIP_KEY` | edge | AES key (base64 of 16, 24 or 32 bytes, `openssl rand -base64 32`) that encrypts and authenticates the memberlist gossip of a cluster of edges: a node without it cannot join. Same value on the edges of a cluster; unset: the gossip is in clear, with a warning |
| `SZ_SYNC_TABLES` | centraldata | Tables (comma-separated) centraldata syncs to the edges; the default set when empty |
| `SZ_SYNC_KEY` | centraldata, edge | HMAC key (base64, 32 bytes at least) signing the changes centraldata publishes; the edge verifies them with the same value |

Some services (edge, dataplane, realtime) load only a specific list of
variables; the others load the whole `Sz` enum.

### Variables of each service

What each service reads, from its environment (`cmd/<service>/.env`, or
`--env-file`) or from Consul. There are no flags for these variables; the
flags of a service only name the env file and, for the edge and the
apiserver, the proxy config and certificates. `SZ_CONSUL_URI` is always read.

| Service | Reads |
|---|---|
| apiserver (`szapiserver`) | `SZ_ADDRESS`, `SZ_CONSUL_URI`, `SZ_POSTGRES_URI`, `SZ_HOSTNAME`, `SZ_CLIENT_ORIGIN`, `SZ_SECRET_KEY`, `SZ_DATA_KEY`, `SZ_ADMIN_USERNAME`, `SZ_ADMIN_PASSWORD`, `SZ_OTLP_INTERNAL_URI`, `SZ_KAFKA_URI` and `SZ_OTLP_EXTERNAL_URI` (audit log), `SZ_TLS_*` (`kafkas://`) |
| realtime (`szrealtime`) | `SZ_ADDRESS`, `SZ_HOSTNAME`, `SZ_SECRET_KEY`, `SZ_OTLP_INTERNAL_URI` |
| edge (`szedge`) | `SZ_ADDRESS`, `SZ_CONSUL_URI`, `SZ_ZONE`, `SZ_MEMBERSHIP_ADDRESS`, `SZ_DISCOVERY_ADDRESS`, `SZ_OLRIC_ADDRESS`, `SZ_OLRIC_PASSWORD`, `SZ_GOSSIP_KEY`, `SZ_SECRET_KEY`, `SZ_DATA_KEY`, `SZ_KAFKA_URI`, `SZ_OTLP_INTERNAL_URI`, `SZ_OTLP_EXTERNAL_URI`, `SZ_SERVICE_TOKEN`, `SZ_SYNC_KEY`, `SZ_TLS_CA`, `SZ_TLS_CERT`, `SZ_TLS_KEY` |
| dataplane (`szdataplane`) | `SZ_ADDRESS`, `SZ_HOSTNAME`, `SZ_SECRET_KEY`, `SZ_OTLP_INTERNAL_URI` |
| centraldata (`szcentraldata`) | `SZ_ADDRESS`, `SZ_CONSUL_URI`, `SZ_POSTGRES_URI`, `SZ_SECRET_KEY`, `SZ_SERVICE_KEY`, `SZ_SYNC_KEY`, `SZ_SYNC_TABLES`, `SZ_OLRIC_URI`, `SZ_OLRIC_PASSWORD`, `SZ_TLS_CA`, `SZ_TLS_CERT`, `SZ_TLS_KEY` |
| collector (`szcollector`) | `SZ_ADDRESS`, `SZ_CONSUL_URI`, `SZ_KAFKA_URI`, `SZ_POSTGRES_URI`, `SZ_TLS_CA`, `SZ_TLS_CERT`, `SZ_TLS_KEY` (`kafkas://`) |
| greeter (`szgreeter`) | `SZ_ADDRESS`, `SZ_CONSUL_URI` |

Required, or the service does not start: `SZ_DATA_KEY` (apiserver), and
`SZ_ZONE` with `SZ_CONSUL_URI` (edge), and a well-formed URI in every
`SZ_*_URI` that is set. Centraldata without `SZ_OLRIC_URI` publishes
nothing; with it, `SZ_SYNC_KEY` is required.

Keep out of the shared env, in the `.env` of the one service: `SZ_SERVICE_KEY`
(centraldata), `SZ_SERVICE_TOKEN` and the `SZ_TLS_*` files (each service has
its own certificate).

### Config from Consul

With `SZ_CONSUL_URI` set, every service also reads its variables from the
Consul KV, once, when it starts (`pkg/appconfig`; a service only declares its
flags and the keys it reads in its `config/config.go`).

| Consul key | For |
|---|---|
| `sentinez/env/SZ_KEY` | every service |
| `sentinez/env/<service_key>/SZ_KEY` | one service, e.g. `sentinez/env/sentinez.dmz.edge.v1/SZ_KAFKA_URI` (`service_key` is in the service proto, `x_meta`) |

- **Precedence:** the environment (`.env` included) wins, then the key of the
  service, then the shared key. An empty value counts as not set.
- **Only the environment** sets `SZ_CONSUL_URI`, `SZ_ADDRESS`,
  `SZ_MEMBERSHIP_ADDRESS`, `SZ_DISCOVERY_ADDRESS`, `SZ_OLRIC_ADDRESS` and
  `SZ_ZONE`: they are
  how Consul is reached, or differ from one instance to the next. Such a key in
  Consul is ignored with a warning, as is an unknown key.
- **Failure:** with `SZ_CONSUL_URI` set and Consul unreachable, the service
  does not start. Without the URI nothing changes.
- **Secrets:** `SZ_SECRET_KEY`, `SZ_DATA_KEY`, `SZ_ADMIN_PASSWORD` and
  `SZ_POSTGRES_URI` can live in Consul too, but its KV is not encrypted:
  protect them with the ACLs of Consul. Changing a value takes a restart.
- **Edge:** it reads `proxy.yaml` from the Consul key `sentinez/proxy.yaml`
  (the one the apiserver imports, `make consul.proxy.put`); without
  `SZ_CONSUL_URI`, or when the key is missing, it reads the file of
  `--proxy_config`.

The env shared by all the services is in
[deploy/consul/shared.env.example](../deploy/consul/shared.env.example); what
differs from one instance to the next (`SZ_ADDRESS`, `SZ_CONSUL_URI`, and
for the edge `SZ_ZONE`, `SZ_MEMBERSHIP_ADDRESS`, `SZ_DISCOVERY_ADDRESS`,
`SZ_OLRIC_ADDRESS`)
stays in `cmd/<service>/.env` (see its `.env.example`).

The keys of `shared.env`:

| Key | Read by | Meaning |
|---|---|---|
| `SZ_HOSTNAME` | apiserver, realtime, dataplane, edge | Domain of the console, the WebAuthn relying party (`sentinez.vn`) |
| `SZ_CLIENT_ORIGIN` | apiserver | Origin of the console: the allowed WebAuthn origin and the only one CORS allows |
| `SZ_SECRET_KEY` | apiserver, realtime, dataplane, edge, centraldata | HMAC key signing the JWTs, base64 |
| `SZ_DATA_KEY` | apiserver, edge | AES-256 key sealing the CA certificates in the database, base64 (`openssl rand -base64 32`). Both must hold the same value; losing it makes the stored certificates unreadable |
| `SZ_SYNC_KEY` | centraldata, edge | HMAC key (base64) signing the changes centraldata publishes; the edges drop what is not signed with it. Shared by these two only |
| `SZ_OLRIC_PASSWORD` | centraldata, edge | Password of the Olric port of the edges |
| `SZ_GOSSIP_KEY` | edge | AES key of the memberlist gossip of the edges, base64 |
| `SZ_ADMIN_USERNAME`, `SZ_ADMIN_PASSWORD` | apiserver | Static admin account |
| `SZ_POSTGRES_URI` | apiserver, centraldata, collector | PostgreSQL URI |
| `SZ_TIMESCALE_URI`, `SZ_CLICKHOUSE_URI` | none | Declared, unused |
| `SZ_KAFKA_URI` | edge, apiserver, collector | Kafka the edge traffic events and the audit log go to, and the collector reads from, `kafka://` or `kafkas://` |
| `SZ_OTLP_INTERNAL_URI` | all | OTLP/gRPC collector of the system logs, `http://` or `https://` |
| `SZ_OTLP_EXTERNAL_URI` | edge | OTLP/gRPC collector of the edge traffic logs, `http://` or `https://` |

`SZ_SERVICE_KEY` is not in the file on purpose: only centraldata (and the
operator who signs tokens) holds it, so it goes in the `.env` of centraldata.
A service reads only the keys of its own list ([Variables of each
service](#variables-of-each-service)); the others are ignored. Put a key in
`sentinez/env/<service_key>/` to give one service a different value.

```sh
cp deploy/consul/shared.env.example deploy/consul/shared.env   # ignored by git; edit it
make consul.env.put                                            # once, for every service
make consul.env.put ENV_FILE=cmd/szedge/v1/.env SERVICE=sentinez.dmz.edge.v1  # one service
```

`hack/consul-env-put.sh` puts the `SZ_*` lines of the file (never printing
the values) and skips the keys that only the environment may set. Then remove
the moved keys from the `.env` files.

## 2.4 Command-line flags

Defined in
[types/setting/v1/flags.proto](../api/proto/sentinez/types/setting/v1/flags.proto),
parsed with `pflag` in `pkg/szflag` and `pkg/apps/**/flags`. Values are
validated with protovalidate.

| Flag | Default | Constraint | Service |
|---|---|---|---|
| `--mode`, `-m` | `dev` | `dev\|prod\|sandbox` | all |
| `--log_level` | `debug` | `debug\|info\|warn\|error` | all |
| `--env_file` | per service | — | all (`""` = don't read a file) |
| `--proxy_config` | `./proxy.yaml` | — | edge |
| `--cert_file` | `""` | — | edge, apiserver (with a certificate the apiserver serves HTTPS and HTTP/3; without one the edge serves only the certificates of centraldata) |
| `--cert_key_file` | `""` | — | edge, apiserver |
| `--engine` | `quic` | `std\|hertz\|quic` | edge (HTTP server, see [03-edge.md](03-edge.md#310-choosing-the-engine)) |

`--mode` also sets the DB table prefix (`dev_sentinez_iam_users`, ...) and
enables fx logging when it is `dev`.

## 2.5 Tests & lint

```sh
make test.cover            # go test ./... -cover (root module ONLY)
cd staging/src/github.com/sentinez/core && go test ./...    # each module separately
make lint                  # golangci-lint for root, core, pkg, controlplane, httphz, tools + buf lint
make fmt.proto             # buf format -w
```

Existing tests cover: `core/szlimiter`, `core/szrule`, `core/szmodsec/ruleparser`,
`controlplane/iam` (handler, service with mocks),
`pkg/{szconfig,szerror,szperms,szprotobuf,szlog,sztoken,szseal}`, `pkg/apps/dmz/edge`,
`pkg/apps/gateway/apiserver`.

Lint rules (`.golangci.yaml`): `lll` 80 columns (tab = 4), `funlen` 35 lines,
`errcheck`, `govet`, `misspell`, `revive`, `staticcheck`, `gofmt`,
`goimports`. The full style guide is in
[hack/.agent/skills/go-style-guide/SKILL.md](../hack/.agent/skills/go-style-guide/SKILL.md).

## 2.6 Code generation

| Task | Command |
|---|---|
| Go + gateway + validate + sz + OpenAPI | `cd api/proto/sentinez/<domain>/v1 && ./generate.sh` |
| TypeScript for the UI | `cd api && buf generate` (`ts-proto` plugin → `ui/packages/proto/ts-proto`) |
| Error page templates | `cd staging/.../core/internal/render && templ generate` |
| CRS rules → Go | `bash hack/ruleparser.sh` (needs `szruleparser` on PATH) |
| eBPF → Go | `cd staging/.../bpf/sentinez && go generate .` |
| Copy third-party protos | `bash hack/copy_proto3pt.sh` |
| Copy CRS `.data` files | `bash hack/copy_rule_data.sh` |
| Mocks | `mockery` |
