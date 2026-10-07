# CLAUDE.md

Sentinéz is a Go security platform (WAF / edge proxy with OWASP CRS rule-based filtering, rate limiting, and a web console). Requires Go 1.27+, Node 26+, protobuf/buf, make. The repo is mirrored to GitLab (`ci_sync_repo.yaml`).

## Layout

This is a **multi-module monorepo** using `replace` directives in the root [go.mod](go.mod); there is no go.work.

| Path | Purpose |
|---|---|
| `cmd/` | Service entrypoints: `szapiserver`, `szrealtime`, `szedge/v1`, `szdataplane/v1`, `szgreeter/v1`, `szcentraldata/v1` |
| `internal/` | Root-module private code: `dmz` (edge/dataplane), `funcs` (edge HTTP middleware nodes, chained in `dmz/edge/http/init.go`), `cluster`, `bpf`, `memory`, `defaults` |
| `pkg/` | Root-module shared packages: `apps/{gateway,greeter,dmz}` (service wiring: `gateway/{apiserver,realtime}`, `dmz/{edge,dataplane}`), `network`, `pools`, `protocol`, `queue`, `tracer` |
| `api/` | Separate module (`sentinez/api`): protobuf sources under `proto/sentinez/`: `apps/<domain>/v1` (domain services, REST via grpc-gateway: `analytic`, `centraldata`, `greeter`, `iam`, `security`, `tenant`), `dmz/{edge,dataplane}/v1` (edge `Setting`/`Context`, dataplane), `gateway/{apiserver,realtime}/v1`, `types/` (shared messages: `v1` known/model/options, `setting/v1`, `rule/v1` expressions & actions, `secrule/v1` SecRule engine & events, `coreruleset/v1`, `cdn/v1`, `net/v1` conn, `net/http/v1` request/event); `docs/v1/*.swagger.json` (generated OpenAPI; no centraldata), `buf.gen.yaml`, generated clients, third_party |
| `staging/src/github.com/sentinez/` | Separately-versioned modules (published to their own repos): |
| ↳ `core` | Runtime framework: `runner`, `http`, `grpc`, `limiter`, `modsec` (Coraza), `rules`, `rulesets`, `storage`, `context` |
| ↳ `shared` | Utility libs: `zlog` (zap + OTLP), `config`, `errorx`, `eventq`, `store`, `topic`, `jsonx`, `perms`, `cron`, etc. |
| ↳ `controlplane` | Domain services: `iam`, `tenant`, `security`, `analytic`, `centraldata`, `greeter`, plus `pkg/{crypto,headers,passkey}` |
| ↳ `bpf` | eBPF programs (cilium/ebpf) |
| ↳ `contrib/httphz` | HTTP helpers |
| ↳ `tools` | Code gen / rule parser tooling |
| `ui/` | pnpm + turbo workspace: `apps/{console,web,template}`, `packages/{ui,proto,eslint-config,tsconfig}` |
| `deploy/` | caddy, nginx, docker, monitoring configs, `ruleroot` (CRS rules) |
| `_submodules/` | git submodules: coreruleset, googleapis, grpc-gateway, opentelemetry-proto, protovalidate (`git submodule update --init --recursive`) |
| `hack/` | Dev scripts (lint, proto lint, gobump, timescale up/down, mkcert, rule parsing, gitlab sync) |

Naming: binaries live in `cmd/sz<name>`; `pkg/apps/gateway` = API/realtime gateway, `dmz` = demilitarized zone (edge/dataplane), greeter = internal service. WAF rules are called `SecRule` (proto: API model `apps.security.v1.SecRule`, runtime `types.secrule.v1.SecRule`/`SecRuleLite`, edge wrapper `dmz.edge.v1.SecRule`; REST `/security/secrule(s)`, console route `security/sec-rule`, Go packages `secrule`/`secrules`); the old `RuleBased` name is retired. Domain protos are under `api/proto/sentinez/apps/<domain>/v1`.

## Commands

```sh
make                         # build all default services (outputs to cmd/<svc>/bin/)
make sz.apiserver.run        # build + run API server
make sz.realtime.run
make sz.edge.run             # edge proxy
make sz.dataplane.run        # runs inside the `gateway` netns (sudo)
make sz.greeter.run
make sz.<svc>.build          # build only: apiserver, realtime, edge, dataplane, greeter, centraldata
make sz.<svc>.image.build    # docker image (apiserver, edge, greeter, centraldata)
make compose.up / compose.down
make test.cover              # go test ./... -cover  (root module only)
make lint                    # golangci-lint across root + core/shared/controlplane/httphz, plus buf lint
make lint.core | lint.shared | lint.controlplane | lint.proto
make fmt.proto               # buf format
```

Because modules are separate, `go test ./...` at the root does **not** cover `staging/` modules — run tests inside each module directory, e.g. `cd staging/src/github.com/sentinez/shared && go test ./zlog/...`. After changing deps in a staged module, run `go mod tidy` in that module and the root (replace directives keep them in sync).

Other tooling: `mockery` for mocks, `templ generate` for templ templates, `buf` for protobuf (`cd api && buf generate`).

## Go style guide

Full rules live in [.agent/skills/go-style-guide/SKILL.md](.agent/skills/go-style-guide/SKILL.md) (based on the Uber Go Style Guide); read it before writing or reviewing Go. Scaffolding skills are alongside it in `.agent/skills/` (`create-service`, `create-handler`, `create-module`, `create-repo`, `create-factory`). Key points:

**Enforced by `.golangci.yaml`** (this wins over the skill's softer 99-char limit): `lll` 80 columns (tab = 4), `funlen` 35 lines per function, plus errcheck, govet, misspell, revive (context-as-argument, error-naming, error-strings, indent-error-flow, superfluous-else, unused-parameter, receiver-naming, etc.), staticcheck; formatted with gofmt + goimports.

**Interfaces / types**
- Never use pointers to interfaces; assert compliance with `var _ I = (*T)(nil)`.
- Don't embed mutexes (use a named `mu` field); avoid embedding types in public structs.
- Start enums at `iota + 1` unless zero is meaningful.
- Copy slices/maps at API boundaries.
- Use `time.Time`/`time.Duration`; if impossible, put the unit in the name (`IntervalMillis`).
- Always add field tags on marshaled structs.

**Errors**
- Static no-match: `errors.New`; dynamic: `fmt.Errorf`; matchable: exported `ErrXxx` var or `XxxError` type (unexported: `errXxx`).
- Wrap with `%w` only if callers should match; keep context terse (`"new store: %w"`, not `"failed to ..."`).
- Handle an error once: log or return, not both.
- Use comma-ok type assertions. Never panic in production code (only `Must` at init).

**Control flow / process**
- Return early, no needless `else`, narrow scope with `if err := ...; err != nil`.
- `os.Exit`/`log.Fatal` only in `main()`; prefer a `run() error` pattern.
- Avoid `init()` and mutable globals (use dependency injection); prefix unexported globals with `_`.
- No fire-and-forget goroutines: each needs a stop mechanism and a wait (`WaitGroup`/done channel), exposed via `Close`/`Stop`/`Shutdown`. Channels are size 0 or 1.
- Use `defer` for cleanup.

**Style**
- Imports in two groups: stdlib, then everything else. Group related decls only.
- Packages: lowercase, short, singular, no `util`/`common`. Don't shadow builtins.
- Order: types/consts/vars, `NewXxx`, exported methods, then helpers; roughly call order.
- `var x T` for zero values, `&T{Field: ...}` (named fields) over `new(T)`; return `nil` not `[]T{}`; check `len(s) == 0`.
- No naked bool params (`f(true /* isLocal */)`); use raw string literals; `f`-suffix for printf-style funcs.
- Hot paths only: `strconv` over `fmt`, preallocate `make(map/slice, n)`, hoist `[]byte("...")`.

**Tests**
- Table-driven with subtests: slice `tests`, case `tt`, fields `give*`/`want*`, assertions via testify `require`/`assert`. No branching inside the loop; split into separate tests instead.
- Constructors with 3+ optional params use functional options (an `Option` interface with `apply(*options)`).

## Conventions

- Every Go file starts with the Apache 2.0 license header (`// Copyright 2025 Duc-Hung Ho.`).
- Lint is golangci-lint per module; CI runs `hack/golint.sh` and `hack/protolint.sh`.
- Services are bootstrapped through `core/runner` (generic `Context[T]` over a settings proto), which also initialises logging/OTLP.
- Logging goes through `shared/zlog` (zap). OTLP log export lives in `shared/zlog/otlp.go` (`SetupOTLP`, `OTLPConfig`), wired from `core/runner/runner.go` on branch `feat/OTLP`.
- Commit style: conventional commits (`chore:`, `refactor:`, `feat:`), PR number suffix on main.
- CI (`.github/workflows`): `ci_lint`, `ci_test` (`go test ./...`), `ci_sync_repo`.
