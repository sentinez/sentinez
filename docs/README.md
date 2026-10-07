# Sentinéz Engineering Documentation

**English** | [Tiếng Việt](vi/README.md)

This documentation describes the Sentinéz source code — a Go security
platform made of an edge proxy with a WAF (OWASP CRS via Coraza), a custom
rule engine (SecRule), rate limiting, a CDN cache, an eBPF/XDP dataplane and
a web console.

It was written by reading the source directly at commit `47e5f876` (branch
`main`). When the code changes, update the matching file (and its Vietnamese
counterpart in [vi/](vi/)).

## Contents

| # | Document | Topic |
|---|---|---|
| 1 | [Architecture overview](01-architecture.md) | Services, modules, overall data flow |
| 2 | [Getting started](02-getting-started.md) | Requirements, build, run, environment variables, flags |
| 3 | [Edge proxy](03-edge.md) | Request pipeline, `proxy.yaml`, memory store, cluster |
| 4 | [Rule engine & WAF](04-rule-engine.md) | SecRule expressions, operators, CDN rules, Coraza/CRS, rule parser |
| 5 | [eBPF dataplane](05-dataplane-ebpf.md) | XDP program, BPF maps, dev netns |
| 6 | [Gateway: API server & realtime](06-gateway.md) | grpc-gateway, middleware, Swagger, WebSocket, service discovery |
| 7 | [Control plane](07-controlplane.md) | IAM, tenant, security, analytic, greeter, centraldata; DB layer; auth |
| 8 | [API & Protobuf](08-api-protobuf.md) | Proto layout, custom options, codegen, REST endpoint table |
| 9 | [Framework & shared libraries](09-core-shared.md) | `core/runner`, `core/http`, `zlog`, `config`, `errorx`, ... |
| 10 | [Web console (UI)](10-ui.md) | pnpm/turbo workspace, console app, API calls |
| 11 | [Deployment, CI & module publishing](11-deploy-ci.md) | Docker, monitoring, GitHub Actions, staging sync |
| 12 | [Known issues](12-known-issues.md) | Bugs, unfinished parts and risks found while reading the code |

## Reading conventions

- File paths are relative to the repository root, e.g.
  [internal/dmz/edge/http/init.go](../internal/dmz/edge/http/init.go).
- Modules under `staging/src/github.com/sentinez/<name>` are referred to by
  their import path: `core`, `shared`, `controlplane`, `contrib/httphz`,
  `bpf`, `tools`.
- "SecRule" is a user-defined rule (the old name `RuleBased` is retired);
  "rulesets"/"WAF" is the OWASP CRS running on Coraza.
