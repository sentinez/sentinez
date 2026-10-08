# 3. Edge proxy (`szedge`)

The edge carries the main load: it accepts HTTPS from the Internet, applies
the protection layers, then reverse-proxies to the tenant's origin.

Relevant source:

| Path | Role |
|---|---|
| [cmd/szedge/v1/main.go](../cmd/szedge/v1/main.go) | Entry point, enables pprof on `:6060` |
| [pkg/apps/dmz/edge/](../pkg/apps/dmz/edge/) | `Server` (Start/Shutdown), config, flags, YAML loading |
| [internal/dmz/edge/engine/](../internal/dmz/edge/engine/engine.go) | HTTP engine choice: `Quic` (default, HTTP/3 + HTTP/1.1/2), `Hertz` or `Standard` (`net/http`) |
| [internal/funcs/](../internal/funcs/) | Middleware chain nodes |
| [internal/dmz/edge/transport/](../internal/dmz/edge/transport/) | TLS (JA4) and OnConnect hooks |
| [internal/memory/](../internal/memory/) | Per-namespace runtime state (routes, proxies, rules, limiters, WAF, CDN) |
| [internal/cluster/](../internal/cluster/) | Embedded Olric + generic `DMap[T]` |
| [contrib/httphz](../staging/src/github.com/sentinez/contrib/httphz/) | Hertz server + `corehttp.Context` + reverse proxy |
| [pkg/network/](../pkg/network/) | TCP (`StdListener`/`StdConn`, with IDs) and QUIC (`QuicListener`) listeners, `net/http` transport |
| [pkg/network/httpx/](../pkg/network/httpx/) | Shared `net/http` `Context` + `ReverseProxy`; servers in `std/` (`stdhttpx`) and `quic/` (`quichttpx`) |

## 3.1 Startup

```
main: runner.Main(NewApp(reverseProxy), NewApp(api))
 ├─ reverseProxy
 │   ├─ engine.Quic(c)         → Inject(quichttpx.NewServer); OnStart: SetOptions(WithOnStdConnect(OnStandardConnect))
 │   ├─ Inject(config.Config, edgeyaml.LoadSetting, edge.New)
 │   └─ Serve(server.Start)
 └─ api
     ├─ Inject(config.Config, edge.NewService)
     └─ Serve(service.Start)   // gRPC on defaults.EdgeAddress

edge.New(conf, setting, server)
 ├─ corecmn.NormalizeEdgeSetting(setting)   // *Lite (YAML) → runtime protos
 └─ memory.NewMemStore(setting)             // singleton; stores Setting in a map + DMap

Server.Start()
 ├─ initialize()
 │   ├─ mem.Start(conf)        → cluster.Start: olric.New + db.Start (goroutine)
 │   ├─ mem.LoadServer(core)   → for each Setting: routes, reverse proxies, CDN, SecRule, limiter, WAF
 │   └─ core.Handle(http.Init(conf, mem).Handle)
 ├─ newTLSConfig(cert, key)  → {GetConfigForClient: transport.TLSConfig, MinVersion: TLS1.3, Certificates}
 ├─ network.StdListen(SENZ_ADDRESS | 0.0.0.0:443)   // TCP
 ├─ network.QuicListen(same addr, WithTLSConfig)    // UDP
 └─ core.ListenAndServe(addr,
        WithCertificate(cert, key), WithTLSConfig(tlsConf),
        WithStdListener(tcp), WithQuicListener(udp), + options (OnStdConnect))
```

The certificate is loaded into the `tls.Config` **before** the listeners are
created: `QuicListen` clones the config (via `http3.ConfigureTLSConfig`, which
also adds the `h3` ALPN), so certificates added later would never be served.
QUIC has no plaintext mode, so the QUIC engine fails with
`quic: tls certificate required` when no certificate is configured.

### QUIC engine (default)

[pkg/network/httpx/quic/server.go](../pkg/network/httpx/quic/server.go)
(`quichttpx.Server`) serves on the same address:

- **HTTP/3 over UDP** with `quic-go/http3` (`ServeListener`, 120s idle
  timeout).
- **HTTP/1.1 and HTTP/2 over TCP** with `net/http` (`ServeTLS`, 120s idle,
  10s read-header timeout). Every TCP response gets an `Alt-Svc` header
  (`http3.Server.SetQUICHeaders`), so clients switch to HTTP/3 on their next
  request — clients always reach TCP first, so both listeners are required.

Both servers run in an `errgroup`; when either exits, the other is closed.
`Shutdown` shuts both down. Requests on both protocols go through the same
`http.ServeMux` and the same middleware chain, using the shared
`httpx.Context`. QUIC is only terminated at the edge: `AcceptReverse` uses
`network.StandardTransporter()`, so upstreams are reached over TCP
(HTTP/1.1 or HTTP/2). `network.QuicTransporter` (an `http3.Transport`)
exists but is not used.

### Hertz engine

The Hertz server ([httphz/server.go](../staging/src/github.com/sentinez/contrib/httphz/server.go))
enables ALPN, H2C, HTTP/2 (`AddProtocol("h2", ...)`) and body streaming, and
uses a `net/std` transport forked from Hertz. Every route goes to
`NoRoute(handler)` — a single handler for all paths. It only uses the TCP
listener (`opt.StdListener`).

## 3.2 Fingerprints & connection context

1. `network.StdListener.Accept` wraps the `net.Conn` as
   `*network.StdConn{Id}` (8 random bytes, hex).
2. `transport.TLSConfig(chi)` (the `GetConfigForClient` callback) computes
   `ja4plus.JA4(chi)` and calls `ja4.Set(conn.Id, fp)` — an in-memory cache
   with a 24h TTL.
3. `OnHertzConnect` / `OnStandardConnect` (registered with
   `WithOnStdConnect`) unwrap `TLSConn → tls.Conn → network.StdConn` and put `netpb.Transport{ConnId, ServerName(SNI)}` into the
   `context`.
4. `httphz.NewContext` reads the transport from the context and sets
   `Request.Fingerprint = ja4.Get(connId)`.
5. `network.StdConn.Close` removes the fingerprint from the cache.

This only applies to TCP connections. For HTTP/3, `chi.Conn` is not a
`*network.StdConn` and the QUIC engine sets no `OnQuicConnect` hook, so
HTTP/3 requests carry no JA4 fingerprint and no `netpb.Transport` (see
[12-known-issues.md](12-known-issues.md)).

## 3.3 `corehttp.Context`

An interface that abstracts request/response
([core/http/context.go](../staging/src/github.com/sentinez/core/http/context.go)):
`RequestContext` (read/write headers, query, path, body, IP, JA4, ...) +
`ResponseWriter` + `X()`, which returns `edgepb.ContextExtra` (holding the
`namespace`).

There are three implementations:

| Implementation | File | Used for |
|---|---|---|
| `httphz.Context` | `contrib/httphz/request_context.go` | Edge (Hertz) |
| `httpx.Context` | `pkg/network/httpx/context.go` | Edge in `Quic` (HTTP/3 and TCP) and `Standard` modes, realtime |
| `corehttpreq.RequestContext` | `core/http/request/context_value.go` | Evaluating rules against an `httppb.Request` (gRPC `EvaluateRuleset`) |

Contexts come from a `sync.Pool` and are returned in `Close()`.

`RequestIP()`: the Hertz version uses Hertz's `ClientIP()`; the `net/http`
version (`httpx.Context`) prefers `X-Forwarded-For` (first entry) → `X-Real-IP` →
`RemoteAddr`.

`httpx.Context.ResponseBody()` returns what the reverse proxy actually
received from the upstream: `httpx.ReverseProxy` wraps the `ResponseWriter`
in a `recordWriter` that records the status code and copies the body (up to
4 MiB; larger bodies are streamed through but not captured, and
`ResponseBody()` returns `nil`). Pooled contexts drop body buffers larger
than 64 KiB. `SetPath` copies the path, since callers may pass pooled
buffers.

## 3.4 Middleware chain

Defined in
[internal/dmz/edge/http/init.go](../internal/dmz/edge/http/init.go). Each
node embeds `*corechains.Node` (which holds `next`) and implements
`Handle(ctx) error`. Nodes are wrapped with `trace.Wrap(name, node)`, which
tags errors with the stage name (`tracer.StageError{Stage, Elapsed, Err}`)
and logs them once, at the stage that produced them.

| Order | Stage | Node | Behavior |
|---|---|---|---|
| 0 | — | `trace.Trace` | Generates request ID `senz:req:<xid>`, sets header `X-Request-Id` |
| 1 | `LOG` | `logging.Logging` | Calls `next` first, then logs an `httppb.Event` (JSON, kind `LOG_TYPE_HTTP`). Also looks up eBPF bandwidth by IP |
| 2 | `DMA` | `secure.DomainBased` | Host must be `<ns>.<SENZ_HOSTNAME>` (single level, port stripped). Otherwise → `403`. Sets `X().Namespace = ns` |
| 3 | `CDN` | `cdn.Cache` | If the namespace has an active CDN rule whose expression matches: look up the cache by `method+scheme+host+path+query+Accept-Encoding`; hit → return the cached response with `X-Cache: HIT`; miss → continue and store cacheable responses (TTL 1h, see [04-rule-engine.md](04-rule-engine.md#44-cdn-rules)) |
| 4 | `LMT` | `ratelimiter.Limiter` | `limiter.Allow(IP)`; exceeded → `429` |
| 5 | `ROM` | `room.WaitingRoom` | Placeholder, just forwards |
| 6 | `STC` | `static.Static` | After the rest of the chain, if the path has a static extension (`.css .js .png ...`) sets `Cache-Control: public, max-age=3600, immutable` |
| 7 | `RUL` | `secure.SecRule` | Walks the active SecRules by descending priority; the first matching rule with action `BLOCK` → `403`; other actions are only debug-logged and evaluation continues |
| 8 | `WAF` | `secure.WAF` | Coraza transaction: `ExecIngress` (headers/body) → rest of chain → `ExecEgress` (response). Blocked with status 403 → 403 page. Finally logs a `secrulepb.Event` (`LOG_TYPE_WAF`) if the transaction was interrupted |
| 9 | `ROU` | `routing.StandardRouter` | Finds the route, rewrites the path, sets headers, calls `ReverseProxy.Serve` |

Error responses (`400/403/404/429/500`) are rendered with `templ` templates
([core/common/render](../staging/src/github.com/sentinez/core/common/render/))
that include the `X-Request-ID`; if rendering fails they fall back to plain
text.

## 3.5 `proxy.yaml` configuration

The YAML file has a root key `setting`; it is converted to JSON and then
`protojson.Unmarshal`-ed into `edgepb.Setting`
([pkg/apps/dmz/edge/yaml/config.go](../pkg/apps/dmz/edge/yaml/config.go)).
Field names therefore follow the proto's **JSON camelCase** and enums are
written as strings. Full example:
[cmd/szedge/v1/proxy.yaml](../cmd/szedge/v1/proxy.yaml).

```yaml
setting:
  metadata: {}
  server:
    name: badcheese              # namespace = subdomain
    locations:
      - location: /api           # prefix to match
        proxyRewrite: /          # replaces the prefix before forwarding (default = location)
        proxyPass:
          - server: api.example.com
            protocol: PROXY_PROTOCOL_HTTPS   # or PROXY_PROTOCOL_HTTP
        balanceStrategy: BALANCE_STRATEGY_ROUND_ROBIN
        proxySetHeaders:
          X-Real-IP: $remote_addr
  security:
    rulesets:                    # any entry → enables CRS for the namespace
    rules:                       # SecRules (Lite form)
      - ingress:
          name: "Block /block"
          status: STATUS_ACTIVE
          expr:
            orCondition:
              - rules:
                  - condition:
                      source: FIELD_SOURCE_PATH
                      operator: OPERATOR_EQ
                      value: "/block"
          action:
            type: ACTION_TYPE_BLOCK
    limiters:
      - timeWindow: 5s
        maxRequests: 50
        timeout: 5s
  traffic_control: {}
  delivery:
    cdn:
      - rule:
          name: cache GET
          status: STATUS_ACTIVE
          expr: { orCondition: [ { rules: [ { condition: { source: FIELD_SOURCE_METHOD, operator: OPERATOR_EQ, value: GET } } ] } ] }
  personal: {}
```

Notes:

- The "Lite" types (`SecRuleLite`, `cdn.RuleLite`, `ExpressionLite`) use
  strings for `source/operator/status/value`. `NormalizeEdgeSetting` converts
  them to the runtime types (`SecRule`, `cdn.Rule`, `Expression`) with real
  enums, turns `value` into a string `structpb.Value`, and generates IDs
  `senz.rule.<nanoid>` / `senz.cond.<nanoid>`.
- Variables supported in `proxySetHeaders`
  ([core/http/variable](../staging/src/github.com/sentinez/core/http/variable/variable.go)):
  `$host`, `$remote_addr`, `$scheme`, `$request_uri`,
  `$proxy_add_x_forwarded_for`. They only work when the **whole value** is
  one variable; any other string containing `$` makes routing fail. Invalid
  header keys (RFC 7230) are dropped with a warning.
- The edge always adds `X-Forwarded-Prefix: <location>`.

## 3.6 Routing

[internal/memory/routes/routes.go](../internal/memory/routes/routes.go):

- `Store(server)`: copies locations, filters valid headers, sets the default
  `proxyRewrite`, and **sorts by `location` length, descending**, for
  nginx-style longest-prefix matching.
- `Match(ctx)`: loads the routes for the namespace, finds the first location
  that matches the path **on a segment boundary** (`/api` matches `/api` and
  `/api/x` but not `/apix`; a location ending in `/` matches any path under
  it), sets the new path to `rewrite + (path − location)` without doubling
  the `/` between them, sets headers, and returns `proxyPass[0].server`.

Examples:

| location | proxyRewrite | Request path | Forwarded path |
|---|---|---|---|
| `/` | `/` | `/about` | `/about` |
| `/api` | `/v1` | `/api/users` | `/v1/users` |
| `/api` | `/` | `/api/users` | `/users` |
| `/api` | `/` | `/api` | `/` |
| `/api` | `/v1/` | `/api/users` | `/v1/users` |
| `/api` | — | `/apix` | not matched by `/api` |

A reverse proxy is created once per `upstream.server` in
`MemStore.LoadReverseProxy` via `server.AcceptReverse(target)`:

- Hertz (`proxyhz.ReverseProxy`): two clients (TLS / plain); the director
  rewrites the URI with `JoinURLPath` and sets `Host` to the upstream host.
  Requests with `Upgrade: websocket` go to `WSReverseProxy`.
- `net/http` (`httpx.ReverseProxy`, used by the `Quic` and `Standard`
  engines): `httputil.NewSingleHostReverseProxy` with the given
  `http.RoundTripper` (`network.StandardTransporter()`, 30s dial timeout).
  The writer is wrapped so the context sees the upstream status and body;
  `Unwrap` keeps `Flush`/`Hijack` reachable for streaming and upgrades.

## 3.7 Memory store

`memory.MemStore` (a singleton) groups stores keyed by **namespace**
(`server.name`):

| Store | Type | Contents |
|---|---|---|
| `settings.Setting` | `Map[string,*edgepb.Setting]` + `cluster.DMap` | Original Settings; also written to the Olric DMap `sentinez.dmz.edge.setting` |
| `routes.Router` | `Map[string,[]*Location]` | Sorted locations |
| `reverseproxy.ReverseProxy` | `Map[string,corehttp.ReverseProxy]` | Key = upstream server |
| `secrules.SecRule` | `Map[string,[]Entry]` | Compiled `Entry{Eval, Rule}`, active only, sorted by priority |
| `ratelimiter.Limiter` | `Map[string,*RateLimiter]` | One limiter per namespace |
| `rulesets.Rulesets` | `Map[string,coraza.WAF]` | One WAF instance per namespace |
| `cdnrules.Rule` | `Map[string,EvalFunc]` + `Map[string,*cdn.CDN]` | One CDN rule per namespace |

All maps use `shared/sync.Map[K,V]` (a generic wrapper around `sync.Map`).

## 3.8 Cluster (Olric)

[internal/cluster](../internal/cluster/):

- `NewCluster` creates Olric with the `local` config. If both
  `SENZ_MEMBERSHIP_ADDRESS` and `SENZ_DISCOVERY_ADDRESS` are set, it sets
  `Peers` and binds memberlist to the discovery address.
- `DMap[T]` is a generic wrapper that stores `proto.Marshal`-ed messages.
  While the cluster is not ready, `Put` pushes into a channel buffered at
  1024 entries; the `wait()` goroutine waits for `dictReady`, creates the
  DMap and drains the queue.
- It is only used to distribute `Setting`s for now; nothing reads them back
  (`Get`) from other nodes yet.

## 3.9 EdgeService gRPC

`edgepb.EdgeService` has `Status` and `EvaluateRuleset`
([internal/dmz/edge/api/api.go](../internal/dmz/edge/api/api.go)), packaged
as `edge.Service` (`pkg/apps/dmz/edge/edge_service.go`). It is currently
**not started** by the edge `main`, and `EvaluateRuleset` evaluates an empty
`Expression`.

## 3.10 Choosing the engine

`main` calls `engine.Quic(c)`. To switch, replace it with `engine.Hertz(c)`
(Hertz, TCP only) or `engine.Standard(c)` (`net/http`, TCP only). The
`stdhttpx.Server` sets 15s read/write and 120s idle timeouts when TLS is on,
and uses `ConnContext` for `OnStdConnect`. The edge `Server.Start` always
opens both the TCP and the UDP listener, whichever engine is used.
