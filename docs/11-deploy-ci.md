# 11. Deployment, CI & module publishing

## 11.1 Docker images

Each `cmd/<svc>` has its own Dockerfile; build from the repo root:

```sh
make sz.apiserver.image.build    # default TAG sentinez/apiserver
make sz.edge.image.build         # sentinez/edge
make sz.greeter.image.build      # sentinez/greeter
make sz.centraldata.image.build  # sentinez/centraldata
```

Common pattern: a `golang:1.27.1-alpine` stage (CGO off, `go mod tidy`,
`go build -ldflags="-s -w"`) → an `alpine` stage with `ca-certificates`,
`tzdata`.

| Image | Run command |
|---|---|
| apiserver | `docker-entrypoint.sh`: `--log_level info --mode prod --env_file ""` (variables come from the container environment) |
| edge | `docker-entrypoint.sh`: adds `--proxy_config ./proxy.yaml`, `--engine quic`, certs in `/etc/senz/ssl/{certs,private}`. The image **bakes in** `proxy.yaml` and the cert/key pair from `cmd/szedge/v1` |
| greeter, centraldata | Run `/bin/main` |

## 11.2 Docker Compose

[deploy/docker/docker-compose.yaml](../deploy/docker/docker-compose.yaml)
(`make compose.up` / `compose.down`):

| Service | Image | Ports | Notes |
|---|---|---|---|
| `sentinez.postgres` | `postgres:16` | 5432 | user/pass `root/root`, DB `sentinez`, volume `_volume/postgres` |
| `sentinez.consul` | `hashicorp/consul` | 8500, 8600/udp | `agent -dev` |
| `sentinez.gateway.apiserver` | `sentinez/apiserver` | 8080 | `env_file: cmd/szapiserver/.env`, overrides `SENZ_POSTGRES_URI` |
| `sentinez.dmz.edge` | `sentinez/edge` | 7443/tcp, 7443/udp (HTTP/3), 6060 (pprof) | `env_file: cmd/szedge/v1/.env` |

Redis, TimescaleDB and NATS are defined but commented out. There are also
`deploy/docker/{cluster,mesh}/docker-compose.yaml` for multi-node scenarios.

## 11.3 Other configs in `deploy/`

| Directory | Contents |
|---|---|
| `caddy/` | `Caddyfile`: `localhost { reverse_proxy localhost:7777 }` + compose |
| `nginx/` | `nginx.conf` + `conf/` |
| `monitoring/` | Prometheus (`prometheus/prometheus.yml`), Grafana, dashboard `sentinez-edge-cluster-grafana.json` |
| `configs/{dev,prod,sandbox}` | Per-environment configs |
| `ruleroot/` | CRS: `setup.conf`, `default.conf`, `audit.conf`, `block.conf`, `modsecurity.conf`, `REQUEST-901/949`, directories `v4-16-0`, `v4-17-0` (`*.conf` + `*.data`) |

## 11.4 Publishing staging modules

Modules in `staging/src/github.com/sentinez/<name>` are synced to
`github.com/sentinez/<name>`:

- Targets are listed in
  [staging/publishing/rules.yaml](../staging/publishing/rules.yaml)
  (`core`, `shared`, `tools`, `contrib`, `controlplane`, `bpf`), together
  with the branches (`staging`, `main`).
- The workflow [ci_sync_repo.yaml](../.github/workflows/ci_sync_repo.yaml)
  is **generated** by `bash staging/publishing/gen_sync_workflow.sh` — edit
  `rules.yaml` and rerun the script; do not edit the workflow by hand.
- Each job runs [hack/ci/action_sync.sh](../hack/ci/action_sync.sh)
  `<name>`: `git subtree split --prefix=staging/src/github.com/sentinez/<name>`,
  merges the remote history with the `ours` strategy, then
  `git push --force` to the branch of the same name. Requires the
  `SENZ_GITHUB_TOKEN` secret.

Because the root module points into `staging/` with `replace`, changes in
staging take effect at the root immediately; after changing a staging
module's dependencies, run `go mod tidy` in that module **and** at the root.

## 11.5 GitHub Actions

| Workflow | Trigger | Job |
|---|---|---|
| `ci_lint.yaml` | push `main`, `release/*`; PRs | `golangci-lint v2.13.2` via `hack/golint.sh` (`go mod tidy && golangci-lint run` at the root); `buf lint` via `hack/protolint.sh` |
| `ci_test.yaml` | push `main`, `release/*`; PRs | `go mod download && go test ./...` (root module only) |
| `ci_sync_repo.yaml` | push `staging`, `main` | Syncs staging modules |

Dependabot updates `gomod` (root) and `npm` (`ui/apps/console`) daily.

Note: CI lint/test only cover the root module; staging modules must be
linted locally with `make lint`.

## 11.6 GitLab mirror

```sh
bash hack/gitremote.sh    # add the gitlab remote
bash hack/gitlab_sync.sh  # push --all and --tags to gitlab
```

## 11.7 Benchmarking & profiling

- `hack/benchmark/greeter_say.sh`: `bombardier -c 100 -n 10000` against the
  edge.
- `hack/benchmark/gobench.sh`: `go test -bench . -benchmem` with CPU/memory
  profiles.
- The edge always exposes `net/http/pprof` on `:6060`:
  `go tool pprof http://localhost:6060/debug/pprof/profile`.
