# 2. Getting started

## 2.1 Requirements

- Go 1.27+ (`go.mod` sets `go 1.27.1`)
- Node 26 + pnpm (for `ui/` and for generating TypeScript from protos)
- `protoc` / `buf`, plus the plugins pinned in
  [tools/tools.go](../staging/src/github.com/sentinez/tools/tools.go)
  (`protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`,
  `protoc-gen-openapiv2`, `protoc-gen-go-senz`, ...)
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
bash hack/timescale-up.sh  # start only the sentinez.postgres container
```

### API server

```sh
cp cmd/szapiserver/.env.example cmd/szapiserver/.env   # then edit the values
make sz.apiserver.run
# REST:    http://localhost:8080/...
# Swagger: http://localhost:8080/swagger/
```

On startup the repositories create their own tables (`CREATE TABLE IF NOT
EXISTS` + `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`); there are no separate
migrations.

### Edge proxy

```sh
./hack/mkcert-local.sh     # creates a *.sentinez.vn cert (see note below)
# /etc/hosts: 127.0.0.1 example.sentinez.vn sentinez.vn
make sz.edge.run
```

`make sz.edge.run` passes `--cert_file`, `--cert_key_file`,
`--proxy_config=./cmd/szedge/v1/proxy.yaml` and `--env_file`. The edge needs
`SENZ_HOSTNAME` to match the root domain: a request to
`badcheese.<SENZ_HOSTNAME>` uses the config with `server.name: badcheese`.
The edge also opens pprof on `:6060`.

> `hack/mkcert-local.sh` runs `mv *.pem ./cmd/edge/v1/` (an old path). Move
> the files to `cmd/szedge/v1/` or pass the cert flags yourself.

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

The console reads `NEXT_PUBLIC_BASE_PATH` (default `http://localhost:8080`).

## 2.3 Environment variables

Variable names are the values of the `Senz` enum in
[types/setting/v1/config.proto](../api/proto/sentinez/types/setting/v1/config.proto).
They are loaded by `shared/config.LoadEnv` (using `godotenv`) and read with
`conf.Get(settingpb.Senz_...)` or `config.GetEnv(...)`.

| Variable | Used by | Meaning |
|---|---|---|
| `SENZ_HOSTNAME` | edge, IAM | Edge root domain (`sentinez.vn`); WebAuthn RP ID |
| `SENZ_ADDRESS` | all | Listen address. Edge defaults to `0.0.0.0:443` and listens on both TCP and UDP (HTTP/3) at that address; apiserver defaults to `:9000` when empty |
| `SENZ_SECRET_KEY` | IAM | HMAC key for signing JWTs, **must be base64** |
| `SENZ_CLIENT_ORIGIN` | IAM | Allowed WebAuthn origin (console URL) |
| `SENZ_POSTGRES_URI` | control plane | PostgreSQL DSN (pgxpool) |
| `SENZ_TIMESCALE_URI`, `SENZ_CLICKHOUSE_URI` | — | Declared but unused |
| `SENZ_CONSUL_URI` | gateway, gRPC | Consul for service discovery |
| `SENZ_MEMBERSHIP_ADDRESS` | edge | Olric peer to join |
| `SENZ_DISCOVERY_ADDRESS` | edge | `host:port` memberlist bind (default `0.0.0.0:3322`) |
| `SENZ_ADMIN_USERNAME` / `SENZ_ADMIN_PASSWORD` | IAM | Static admin account |
| `SENZ_OTLP_ENDPOINT` | all | OTLP/gRPC collector (default `localhost:4317`) |

Some services (edge, dataplane, realtime) load only a specific list of
variables; the others load the whole `Senz` enum.

## 2.4 Command-line flags

Defined in
[types/setting/v1/flags.proto](../api/proto/sentinez/types/setting/v1/flags.proto),
parsed with `pflag` in `shared/flagx` and `pkg/apps/**/flags`. Values are
validated with protovalidate.

| Flag | Default | Constraint | Service |
|---|---|---|---|
| `--mode`, `-m` | `dev` | `dev\|prod\|sandbox` | all |
| `--log_level` | `debug` | `debug\|info\|warn\|error` | all |
| `--env_file` | per service | — | all (`""` = don't read a file) |
| `--proxy_config` | `./proxy.yaml` | — | edge |
| `--cert_file` | `./_wildcard.sentinez.vn+1.pem` | — | edge |
| `--cert_key_file` | `./_wildcard.sentinez.vn+1-key.pem` | — | edge |

`--mode` also sets the DB table prefix (`dev_sentinez_iam_users`, ...) and
enables fx logging when it is `dev`.

## 2.5 Tests & lint

```sh
make test.cover            # go test ./... -cover (root module ONLY)
cd staging/src/github.com/sentinez/core && go test ./...    # each module separately
make lint                  # golangci-lint for root, core, shared, controlplane, httphz, tools + buf lint
make fmt.proto             # buf format -w
```

Existing tests cover: `core/limiter`, `core/rules`, `core/modsec/ruleparser`,
`controlplane/iam` (handler, service with mocks), `controlplane/pkg/crypto`,
`shared/{config,errorx,perms,protobuf,zlog}`, `pkg/apps/dmz/edge`,
`pkg/apps/gateway/apiserver`.

Lint rules (`.golangci.yaml`): `lll` 80 columns (tab = 4), `funlen` 35 lines,
`errcheck`, `govet`, `misspell`, `revive`, `staticcheck`, `gofmt`,
`goimports`. The full style guide is in
[.agent/skills/go-style-guide/SKILL.md](../.agent/skills/go-style-guide/SKILL.md).

## 2.6 Code generation

| Task | Command |
|---|---|
| Go + gateway + validate + senz + OpenAPI | `cd api/proto/sentinez/<domain>/v1 && ./generate.sh` |
| TypeScript for the UI | `cd api && buf generate` (`ts-proto` plugin → `ui/packages/proto/ts-proto`) |
| Error page templates | `cd staging/.../core/common/render && templ generate` |
| CRS rules → Go | `bash hack/ruleparser.sh` (needs `ruleparser-sentinez` on PATH) |
| eBPF → Go | `cd staging/.../bpf/sentinez && go generate .` |
| Copy third-party protos | `bash hack/copy_proto3pt.sh` |
| Copy CRS `.data` files | `bash hack/copy_rule_data.sh` |
| Mocks | `mockery` |
