# Changelog

All notable changes to Sentinéz are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow
[Semantic Versioning](https://semver.org/). Until 1.0.0, any minor or patch
release may break the API, the protos, `proxy.yaml` or the database schema.

The release process is in [docs/13-release.md](docs/13-release.md), the test
plan in [docs/14-test-plan.md](docs/14-test-plan.md).

## [0.0.1] - Unreleased

The first release. It ships every service, the staged Go modules and the
web console at one version.

### Edge proxy (`szedge`)

- Caddy-style handler chain: tracer, logging, domain-based namespaces, CDN
  cache, rate limiter, waiting room (empty), static files, SecRules,
  OWASP CRS rulesets, router.
- HTTP/1.1, HTTP/2 and HTTP/3 (QUIC) listeners. Upstreams over HTTP and HTTPS,
  and HTTP/3 to upstreams that advertise it with `Alt-Svc`.
- SecRule engine: expressions over the path, method, IP, headers, query,
  body, JA4 and TLS fields. Actions: `BLOCK`, `CHALLENGE` (proof-of-work
  page with a `pow-pass` cookie) and `REDIRECT`.
- Rate limiters scoped per key, each selected by an ingress expression.
- Lua rule scripts: a secrule or a rate limiter may set `script` in place of
  `expr` (and `action`). A secrule script defines `match(ctx)` returning an
  action constant (`return SZ_BLOCK`) or `{type = SZ_REDIRECT, params = ...}`;
  a limiter script defines `key(ctx)`, the key to count the request under.
  `internal/rulescript` converts an expression to a script and back.
- OWASP CRS 4.16.0 and 4.17.0 on Coraza. Paranoia and detection levels,
  anomaly thresholds, rule exclusions scoped by expressions.
- CDN cache with `X-Cache: HIT`.
- Lua hooks around the chain stages, declared with
  `---@stage http.request|response/<stage>`. The stages are `caching`,
  `secratelimit`, `waitingroom`, `staticassets`, `secrule`, `secruleset` and
  `router`. Each call has a 50 ms limit.
- Clustering through embedded Olric, and zone leader election through Consul.
- JA4 fingerprints, GeoIP-enriched request logs, and upstream latency.
- Traffic events are sent to Kafka (raw protos) or OTLP.
- Settings come from `proxy.yaml` or the Consul key `sentinez/proxy.yaml`,
  plus the tables `szcentraldata` syncs. TLS modes are available, including
  passthrough.

### Dataplane (`szdataplane`)

- XDP/eBPF program with a BPF map API. The TCP conn id is the kernel socket
  cookie.

### Control plane and gateway

- `szapiserver`: REST through grpc-gateway, Swagger, CORS and an audit log
  of the control plane calls that records the fields an update changed.
- `szrealtime`: WebSocket gateway.
- IAM: local accounts, JWT, passkeys (WebAuthn) and a static admin account.
  Calls are scoped to the tenants of the caller.
- Tenant: tenants, members, and resources (domains) with CA certificates
  sealed by `SZ_DATA_KEY`.
- Security: CRUD for SecRules, rate limits (`SecRateLimit`) and rulesets
  (`SecRuleset`), a rule catalog, and import of `proxy.yaml`.
- Analytic: event lists, histograms over resolution windows, top values,
  egress latency and activities.
- `szcollector`: moves the edge events from Kafka into the analytic event
  tables.
- `szcentraldata`: syncs the chosen tables to the edges over mTLS.
- `szgreeter`: internal service.
- Generated row scanners for the database models. Each repository creates
  its own tables; there are no migrations.

### Web console (`ui/`)

- Next.js console with Vietnamese and English (i18n). It has tenant, domain,
  member and settings pages.
- Pages for SecRules, rulesets and rate limits. The rule condition editor is
  built from the condition schema.
- Analytic log explorer: zoomable histograms (hours, then minutes, then
  seconds), top values and an activity page.
- Sign-in with a sample user when the backend is out of reach.
- Public website app (`ui/apps/web`).

### Go modules

These modules are published from `staging/` at 0.0.1: `core`, `pkg`,
`controlplane`, `contrib/pkg`, `contrib/httphz`, `tools`, `scripting`, `bpf`
and `integration`.

### Known issues

See [docs/12-known-issues.md](docs/12-known-issues.md). The ones that matter
for a deployment are:

- pprof is always open on `0.0.0.0:6060`.
- The edge trusts `X-Forwarded-For` and `X-Real-IP` from any client.
- The edge Docker image bakes in the source tree's certificate.
- The schema sync never renames or retypes columns, so a database created
  before `28dd2173` must be recreated.
- A standalone (not clustered) edge gets no control plane changes, and the
  edge does not reload `proxy.yaml`.

[0.0.1]: https://github.com/sentinez/sentinez/releases/tag/v0.0.1
