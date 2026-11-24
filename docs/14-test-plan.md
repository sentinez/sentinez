# 14. Release test plan

These are the checks a release must pass before it is tagged
([13-release.md](13-release.md), step 3). 14.2 lists the automated gates,
and 14.3 lists the manual tests on a running system. Record each run in 14.5.

## 14.1 Environment

Use a fresh checkout on Linux. Follow
[02-getting-started.md](02-getting-started.md) to set it up:

- `make compose.up`: Postgres/TimescaleDB, Consul, Kafka, ClickHouse and the
  OTLP collector.
- `sudo tee -a /etc/hosts < deploy/hosts`.
- Certificates for the edge: `./hack/mkcert-local.sh`.
- A **new** database. The schema sync does not migrate an older one
  ([12-known-issues.md](12-known-issues.md), L1).
- The services, each started with `make sz.<svc>.run`: apiserver, realtime,
  greeter, collector, centraldata and edge. The dataplane is only needed for
  `DPL`.
- The console: `cd ui && pnpm install && pnpm dev` (http://localhost:3000).
- An upstream on `sentinez.upstream:80` that echoes the request, headers
  included. For example:
  `docker run --rm -p 80:8080 mendhak/http-https-echo`.

In the edge tests, `$EDGE` is the domain of the setting under test, reached
on the edge's listen address. The tests use the shipped
[cmd/szedge/v1/proxy.yaml](../cmd/szedge/v1/proxy.yaml), whose domain is
`badcheese.is.s6z.io.vn`. To run without the internet, point the domain of
that file at a local upstream.

## 14.2 Automated gates

| Id | Command | Pass when |
|---|---|---|
| A-01 | `make release.check` | It exits 0. It runs gofmt, golangci-lint in every module, `buf lint`, `go test` in every module, and builds every service |
| A-02 | `cd ui && pnpm lint && pnpm typecheck && pnpm build` | All three exit 0 (lint runs with `--max-warnings 0`) |
| A-03 | `go test -race ./internal/handlers/... ./pkg/apps/dmz/...` (root) | No race is reported |
| A-04 | `make bench` | No benchmark of the request path regresses by more than 10% against the README table |
| A-05 | `cd api && buf breaking --against '.git#branch=main'` | Not a gate for 0.0.1 (there is no earlier release). Record the result for later releases |

The tests already guard these contracts:

- The Lua stage names map onto the edge chain:
  `internal/handlers/script_stages_test.go`.
- The shipped `proxy.yaml` imports into the control plane:
  `pkg/apps/gateway/apiserver/proxyimport_test.go`.
- The edge setting is normalized: `pkg/sznormalize`.

## 14.3 Manual tests

Each test lists its steps and the result it expects. Run them in order,
because later tests reuse the state that earlier ones create.

### INS: install and start

| Id | Steps | Expected |
|---|---|---|
| INS-01 | `make` | Every binary is built under `cmd/<svc>/bin` or `cmd/<svc>/v1/bin` |
| INS-02 | Start apiserver without `SZ_DATA_KEY` | It does not start, and the log names the missing key |
| INS-03 | Start apiserver with `SZ_CONSUL_URI` set to an address nobody listens on | It does not start |
| INS-04 | Start every service | Each banner shows the release version (`core.Version`). The tables are created in the new database |
| INS-05 | Open http://sentinez.gateway.apiserver:8080/swagger/ | The Swagger UI lists the IAM, tenant, security and analytic services |
| INS-06 | `make consul.proxy.put`, then restart apiserver and edge | The apiserver imports the file under the tenant `admin`. The edge reads its setting from Consul |

### EDG: edge routing

| Id | Steps | Expected |
|---|---|---|
| EDG-01 | `curl -k https://$EDGE/local/anything` | The upstream receives `/anything` (from `proxyRewrite`) along with `X-Forwarded-*`, `X-Real-IP` and `Host` |
| EDG-02 | Request with a `Host` that no setting has | `403` page from `domain` |
| EDG-03 | `curl -k --http3 https://$EDGE/local/` | Served over HTTP/3. A JA4 fingerprint is in the log of the request |
| EDG-04 | Make every upstream of a location unreachable | `502` page |
| EDG-05 | Request with an id in the log | The log line has the request id, `tenant_id` and `resource_id`, and GeoIP fields for a public IP |

### SEC: SecRule, limiter and CRS

| Id | Steps | Expected |
|---|---|---|
| SEC-01 | `curl -k https://$EDGE/block` (a GET) | `403`. The rule "(Path AND Method) OR IP" matches |
| SEC-02 | `curl -k -X POST https://$EDGE/block` | Not blocked by that rule |
| SEC-03 | `curl -k -H 'X-Forwarded-For: 1.1.1.1' https://$EDGE/` | `403` from the IP branch. This also shows known issue S8: the client chooses its own IP |
| SEC-04 | Open `https://$EDGE/login` in a browser | Proof-of-work page. After the checkbox, the page loads and a `pow-pass` cookie is set. The next GET in the hour skips the challenge |
| SEC-05 | Send 101 requests in under 5 s from one IP: `for i in $(seq 101); do curl -sk -o /dev/null -w '%{http_code}\n' https://$EDGE/local/; done \| sort \| uniq -c` | The 101st request gets `429`. The IP is limited for `timeout` (5 s), then served again |
| SEC-06 | Keep SEC-05 going from a second IP | The second IP is not limited (the limit is per key) |
| SEC-07 | `curl -k "https://$EDGE/local/?q=<script>alert(1)</script>"` with the CRS `mode: ENGINE_MODE_ON` | `403` page. A `LOG_TYPE_SECRULESET` event is logged |
| SEC-08 | SEC-07 with `mode: ENGINE_MODE_DETECTION_ONLY` | Reaches the upstream, and the event is still logged |
| SEC-09 | Exclude the rule that SEC-07 hit for `/local`, in `proxy.yaml`; restart the edge | SEC-07 reaches the upstream |
| SEC-10 | Run SEC-07 with `version: CRS_VERSION_V4_16_0` | Same result as SEC-07 |

### CDN: cache

| Id | Steps | Expected |
|---|---|---|
| CDN-01 | GET the same cacheable URL twice | The second response has `X-Cache: HIT`, and the upstream sees one request |
| CDN-02 | GET `/login` twice | No `X-Cache: HIT`. The challenge of SEC-04 still applies |
| CDN-03 | The same URL with another `Accept-Encoding` | A miss (a separate cache key) |

### LUA: scripting hooks

| Id | Steps | Expected |
|---|---|---|
| LUA-01 | The shipped `script_example`; request `/local/` | The upstream receives `X-Edge: sentinez`. With debug logs on, the `tag` line is in the log of the request |
| LUA-02 | Hook `http.request/secrule` with a function that calls `ctx:respond(418, "teapot")`, then request `/block` | `418 teapot`. The SecRule stage is skipped, so there is no `403` |
| LUA-03 | Hook `http.response/router` with a function that calls `ctx:set_response_header("X-Seen", "1")` | The response has `X-Seen: 1` |
| LUA-04 | A function that loops forever | Each request is served after at most about 50 ms of script time. The script error is logged and the chain goes on |
| LUA-05 | `---@stage http.request/sec_rate_limit` (the old name) | The edge logs `edge: load '<domain>': … @stage: unknown stage "sec_rate_limit", one of caching, secratelimit, …` and loads none of the namespace's scripts, keeping the ones it held before. The rest of the namespace still loads |
| LUA-06 | `os.execute`, `io.open` or `require` in a script | Not available (sandbox) |

### CLS: cluster

| Id | Steps | Expected |
|---|---|---|
| CLS-01 | Two edges with `SZ_MEMBERSHIP_ADDRESS`, `SZ_DISCOVERY_ADDRESS`, `SZ_CONSUL_URI` and the same `SZ_ZONE` | They join. One is the leader: `sentinez/dmz/edge/zone/<zone>/status` is written |
| CLS-02 | Stop the leader | The other edge takes over the status key |
| CLS-03 | A setting with a non-default `tenantId` | It is synced between the edges. A setting with `tnt_default` is not (by design) |

### API: control plane through REST and the console

| Id | Steps | Expected |
|---|---|---|
| API-01 | `PUT /iam/login` as `SZ_ADMIN_USERNAME` | A JWT |
| API-02 | `POST /iam/account` (new local user), then log in | Both succeed. A second sign-up with the same email is refused |
| API-03 | Register a passkey in the console at http://localhost:3000, log out, log in with it | Both succeed |
| API-04 | Create a tenant and a resource (domain) as user A. List resources as user B, who is not a member | B does not see A's resource (calls are scoped to the caller's tenants) |
| API-05 | Add B as a member of A's tenant | B now sees the resource |
| API-06 | SecRule CRUD (`/security/secrule(s)`) and the console page `security/sec-rule` | Create, get, update, list and delete all work. The condition editor builds valid expressions |
| API-07 | Rate limit and ruleset CRUD (`/security/sec{ratelimit,ruleset}(s)`) and their console pages | As API-06 |
| API-08 | Upload a CA certificate to a resource | Stored sealed: the PEM is not readable in the table. It reads back through the API |
| API-09 | Update a resource, then open the activity page | An audit entry shows the caller and the changed fields |
| API-10 | Call a PORTAL/ADMIN endpoint without a token, then with an expired one | `401` |

### EVT: events and analytics

| Id | Steps | Expected |
|---|---|---|
| EVT-01 | Edge with `SZ_KAFKA_BROKERS`; replay SEC-01, SEC-05 and SEC-07, plus some normal traffic | Messages arrive on the `sentinez.event.http.*` topics (Kafka UI :8088) |
| EVT-02 | Collector running | The rows are in `analytic.*.events`. Each id is prefixed with `SZ_ZONE` |
| EVT-03 | Console `analytic/logs/sec-rule`, `sec-ruleset` and `sec-rate-limit` | The events of EVT-01 are listed. Top values and filters work |
| EVT-04 | Histogram zoom from hours to minutes to seconds | The buckets match the event counts at each resolution |
| EVT-05 | Egress latency chart | Shows the upstream latency of EDG-01 |
| EVT-06 | Edge with only `SZ_OTLP_EXTERNAL_ENDPOINT` | The traffic logs reach the OTLP collector |

### CDT: centraldata

| Id | Steps | Expected |
|---|---|---|
| CDT-01 | centraldata over mTLS with the edge pointed at it; create a SecRule for the domain in the control plane | The edge loads the setting from the synced tables (see the limits in 12.3) |
| CDT-02 | A client without a valid certificate | The mTLS connection is refused |

### UI: console

| Id | Steps | Expected |
|---|---|---|
| UI-01 | Visit every sidebar page (`e2e/screenshots.mjs` visits them all) | No page errors. The tab spinner covers each switch |
| UI-02 | Switch between Vietnamese and English | Every visible string is translated, with no raw keys |
| UI-03 | Stop apiserver, then open the console | Sign-in with the sample user works (the proxy answers 502) |
| UI-04 | Mobile width (375 px) | The sidebar collapses, and no page scrolls sideways |
| UI-05 | `ui/apps/web` (`pnpm build && pnpm start`) | The public site renders, and the docs pages work |

### DPL: dataplane (Linux, root)

| Id | Steps | Expected |
|---|---|---|
| DPL-01 | `sudo make -f hack/net/dev/Makefile setup && make sz.dataplane.run` | The XDP program attaches inside the `gateway` netns |
| DPL-02 | Traffic from the `client` netns through the edge | The edge sees the socket cookie as the TCP conn id |

## 14.4 Exit criteria

- Every A gate passes.
- Every manual test passes, apart from those marked *known* below. A
  failure that is a known issue ([12-known-issues.md](12-known-issues.md))
  is accepted only if the changelog lists it.
- No new entry in 12.2 (security) is left out of the release notes.

Expected *known* results for 0.0.1:

- SEC-03 succeeds in spoofing the IP (S8).
- `:6060` answers pprof (S5).
- A standalone edge does not get control plane changes until it restarts
  (12.3, #393).

## 14.5 Results

Copy this table into the release PR. Mark each row pass, fail (with a
link to an issue) or n/a.

| Group | Tester | Date | Commit | Result | Notes |
|---|---|---|---|---|---|
| A (gates) | | | | | |
| INS | | | | | |
| EDG | | | | | |
| SEC | | | | | |
| CDN | | | | | |
| LUA | | | | | |
| CLS | | | | | |
| API | | | | | |
| EVT | | | | | |
| CDT | | | | | |
| UI | | | | | |
| DPL | | | | | |
