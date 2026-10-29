# 3. Edge proxy (`szedge`)

The edge carries the main load: it accepts HTTPS from the Internet, applies
the protection layers, then reverse-proxies to the tenant's origin.

Relevant source:

| Path | Role |
|---|---|
| [cmd/szedge/v1/main.go](../cmd/szedge/v1/main.go) | Entry point, enables pprof on `:6060` |
| [pkg/apps/dmz/edge/](../pkg/apps/dmz/edge/) | `Server` (Start/Shutdown), config, flags, YAML loading |
| [internal/dmz/edge/engine/](../internal/dmz/edge/engine/engine.go) | HTTP engine constructors, chosen by `--engine`: `Quic` (default, HTTP/3 + HTTP/1.1/2), `Hertz` or `Standard` (`net/http`) |
| [internal/handlers/](../internal/handlers/) | Middleware chain plugins |
| [internal/dmz/edge/transport/](../internal/dmz/edge/transport/) | TLS (JA4) and OnConnect hooks |
| [internal/memory/](../internal/memory/) | Per-namespace runtime state (routes, proxies, rules, limiters, WAF, CDN) |
| [pkg/cluster/](../pkg/cluster/) | Embedded Olric + generic `DMap[T]` |
| [contrib/httphz](../staging/src/github.com/sentinez/contrib/httphz/) | Hertz server + `szhttp.Context` + reverse proxy |
| [pkg/network/](../pkg/network/) | TCP (`StdListener`/`StdConn`, with IDs) and QUIC (`QuicListener`) listeners, `net/http` transport |
| [pkg/network/httpx/](../pkg/network/httpx/) | Shared `net/http` `Context` + `ReverseProxy`; servers in `std/` (`stdhttpx`) and `quic/` (`quichttpx`) |

## 3.1 Startup

```
main: run() → runner.Run(ctx, NewApp(edgeServer), NewApp(grpcServer), NewApp(pprof))
 ├─ edgeServer
 │   ├─ Inject(config.Config, edgeyaml.LoadSetting, edge.NewServer)
 │   └─ Serve(server.Start)
 └─ grpcServer
     ├─ Inject(config.Config, edge.NewService)
     └─ Serve(service.Start)   // gRPC on defaults.EdgeAddress

edge.NewServer(conf, setting)
 ├─ sznormalize.NormalizeEdgeSetting(setting)   // *Lite (YAML) → runtime protos
 ├─ memory.NewMemStore(setting)             // singleton; stores Setting in a map + DMap
 └─ switch conf.Flag.Engine                 // --engine
     hertz → engine.Hertz(conf)            → httphz.NewServer + WithOnStdConnect(OnHertzConnect)
     quic  → engine.Quic(conf)             → quichttpx.NewServer + WithOnStdConnect(OnStandardConnect)
     std, other → engine.Standard(conf)    → stdhttpx.NewServer + WithOnStdConnect(OnStandardConnect)

Server.Start()
 ├─ initialize()
 │   ├─ mem.Start(conf)        → cluster.Start: olric.New + db.Start (goroutine)
 │   ├─ mem.LoadServer(server) → for each Setting: routes, reverse proxies, CDN, SecRule, limiter, WAF
 │   └─ server.Handle(http.Init(conf, mem).Handle)
 ├─ newTLSConfig(cert, key)  → {GetConfigForClient: transport.TLSConfig, MinVersion: TLS1.3, Certificates}
 ├─ network.StdListen(SENZ_ADDRESS | 0.0.0.0:443)   // TCP
 ├─ network.QuicListen(same addr, WithTLSConfig)    // UDP
 └─ server.ListenAndServe(addr,
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

Every TLS connection gets an ID before its handshake. `transport.TLSConfig`
(the `GetConfigForClient` callback) stores the JA4 fingerprint under that ID
in `shared/szstore/ja4` (in-memory, 24h TTL). The connect hook puts the same
ID into the request context as `netpb.Transport{ConnId, ServerName(SNI)}`, and
`httpx.NewContext` / `httphz.NewContext` sets
`Request.Fingerprint = ja4.Get(connId)`. TCP and QUIC only differ in where
the ID lives:

| | TCP (HTTP/1.1, HTTP/2) | QUIC (HTTP/3) |
|---|---|---|
| ID created by | `network.StdListener.Accept`, as `*network.StdConn{Id}` | `quic.Transport.ConnContext` in `network.QuicListen`, put on the connection context |
| `TLSConfig` reads it with `network.HelloConnId(chi)` from | `chi.Conn.(*network.StdConn).Id` | `network.ConnId(chi.Context())` |
| Connect hook | `OnStandardConnect` / `OnHertzConnect` (`WithOnStdConnect`) unwrap `TLSConn → tls.Conn → network.StdConn` | `OnQuicConnect` (`WithOnQuicConnect`) reads `network.ConnId(conn.Context())` |
| JA4 prefix | `t` | `q` |
| Removed from the cache | `network.StdConn.Close` | `context.AfterFunc` when quic-go cancels the connection context (close or failed handshake) |

IDs are 8 random bytes in hex (`network.newConnId`).

**QUIC details**

- `QuicListen` builds its own `quic.Transport` on a UDP socket (instead of
  `quic.ListenAddr`) so it can set `ConnContext`. quic-go derives the context
  passed to `crypto/tls` (`tls.ClientHelloInfo.Context`) and the one returned
  by `quic.Conn.Context` from the context `ConnContext` returns, so the
  handshake and the requests see the same ID. `QuicListener.Close` closes the
  transport (which closes the listener) and the UDP socket, which quic-go
  does not close for a socket it did not create.
- The ID is generated by the server, not the QUIC Connection ID. quic-go only
  exposes the client-chosen original Destination Connection ID to the qlog
  `Tracer`, which cannot change the context. A client could also reuse a
  DCID on purpose to overwrite another connection's entry. The generated ID
  does not depend on the client address, so it stays the same after
  connection migration.
- `ja4plus` (v0.0.3) picks the protocol character from
  `LocalAddr().Network()`, which is `"udp"` for QUIC, and so writes `d`
  (DTLS). For a `"udp"` handshake `TLSConfig` passes `ja4plus` a copy of the
  `ClientHelloInfo` whose conn reports `"quic"`, so it writes `q` as the JA4
  spec requires. The edge serves no DTLS.
- Test: [transport_quic_test.go](../internal/dmz/edge/transport/transport_quic_test.go)
  runs a real QUIC handshake and checks the `q13…` fingerprint, the SNI and
  the removal after close.

## 3.3 `szhttp.Context`

An interface that abstracts request/response
([core/szhttp/context.go](../staging/src/github.com/sentinez/core/szhttp/context.go)):
`RequestContext` (read/write headers, query, path, body, IP, JA4, ...) +
`ResponseWriter` + `X()`, which returns `edgepb.ContextExtra` (holding the
`namespace`).

There are three implementations:

| Implementation | File | Used for |
|---|---|---|
| `httphz.Context` | `contrib/httphz/request_context.go` | Edge (Hertz) |
| `httpx.Context` | `pkg/network/httpx/context.go` | Edge in `Quic` (HTTP/3 and TCP) and `Standard` modes, realtime |
| `httpreq.RequestContext` | `core/szhttp/httpreq/context_value.go` | Evaluating rules against an `httppb.Request` (gRPC `EvaluateRuleset`) |

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

Built in [internal/handlers/handlers.go](../internal/handlers/handlers.go) with
`chains.Build(envplugin.NewEnv(mem, opts...), Stage, plugins()...)`. Each
plugin implements `chains.Plugin` (`ID()`, `Handle(ctx, next)`) and, when it
needs dependencies, `Provision(*envplugin.Env)`: the
[Env](../internal/envplugin/env.go) resolves the options once and hands out
the sub-stores of `MemStore`, the hostname and event loggers
(`Env.Logger(kind)`), so plugins take no constructor arguments. Every plugin
is wrapped by `handlers.Stage`, which tags errors with the plugin `ID()`
(`sztracer.StageError{Stage, Elapsed, Err}`) and logs them once, at the
stage that produced them; the stage names are exported constants (e.g.
`secure.WAFId`). The edge closes the chain (`Chain.Close`) on shutdown.

| Order | Stage | Plugin | Behavior |
|---|---|---|---|
| 0 | `TRC` | `trace.Tracer` | Generates request ID `senz:req:<xid>` when the context has none |
| 1 | `LOG` | `logging.Logging` | Calls `next` first, then logs an `httppb.Event` (JSON, kind `LOG_TYPE_HTTP`). Also looks up eBPF bandwidth by IP |
| 2 | `DMA` | `secure.DomainBased` | Host must be `<ns>.<SENZ_HOSTNAME>` (single level, port stripped). Otherwise → `403`. Sets `X().Namespace = ns` |
| 3 | `PUB` | `cdn.Public` | Serves the images embedded in the binary under `/cdn/sentinez/public/`; a load failure at provisioning stops the edge |
| 4 | `CDN` | `cdn.Cache` | If the namespace has an active CDN rule whose expression matches: look up the cache by `method+scheme+host+path+query+Accept-Encoding`; hit → return the cached response with `X-Cache: HIT`; miss → continue and store cacheable responses (TTL 1h, see [04-rule-engine.md](04-rule-engine.md#44-cdn-rules)) |
| 5 | `LMT` | `secure.Limiter` | for each limiter whose required `ingress.expr` matches (priority order, like `RUL`): `limiter.Allow(IP)`; exceeded → `429` |
| 6 | `ROM` | `room.WaitingRoom` | Placeholder, just forwards |
| 7 | `STC` | `static.Static` | After the rest of the chain, if the path has a static extension (`.css .js .png ...`) sets `Cache-Control: public, max-age=3600, immutable` |
| 8 | `RUL` | `secure.SecRule` | Walks the active SecRules by descending priority; the first matching rule with action `BLOCK` → `403`; other actions are only debug-logged and evaluation continues |
| 9 | `WAF` | `secure.WAF` | Coraza transaction: `ExecIngress` (headers/body) → rest of chain → `ExecEgress` (response). Blocked with status 403 → 403 page. Finally logs a `secrulepb.Event` (`LOG_TYPE_WAF`) if the transaction was interrupted |
| 10 | `ROU` | `routing.StandardRouter` | Ends the chain: finds the route, rewrites the path, sets headers, calls `ReverseProxy.Serve` |

Error responses (`400/403/404/429/500`) are rendered with `templ` templates
([core/internal/render](../staging/src/github.com/sentinez/core/internal/render/))
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
    core_ruleset:                # OWASP CRS, one configuration per namespace
      status: STATUS_ACTIVE
      mode: ENGINE_MODE_ON       # ENGINE_MODE_OFF | ENGINE_MODE_DETECTION_ONLY | ENGINE_MODE_ON
      version: VERSION_V4_17_0   # VERSION_V4_16_0 (RCE, SQLi only) | VERSION_V4_17_0 (default)
      paranoiaLevel: 1           # 1 (fewest false positives) .. 4 (strictest)
      detectionParanoiaLevel: 2  # log up to this level, block up to paranoiaLevel; 0 = same
      threshold:                 # anomaly score needed to block, 0 = setup.conf value
        inbound: 5
        outbound: 4
        earlyBlocking: true
      categories:                # enabled rule groups, empty = all
        - CATEGORY_RCE           # 932
        - CATEGORY_XSS           # 941
        - CATEGORY_SQLI          # 942
      request:
        bodyAccess: true         # inspect request bodies
        bodyLimitBytes: 13107200 # 12.5 MiB, 0 = default
        allowedMethods: [GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS]
        # allowedContentTypes: [application/json]   # empty = setup.conf value
      exclusions:                # false-positive tuning
        - location: /api/search  # skip SQLi rules under this prefix
          tags: [attack-sqli]
        - targets:               # no location = whole namespace
            - ruleId: 942100
              variable: ARGS:password
      overrides:                 # per-rule state
        - id: 920420
          state: RULE_STATE_DISABLED        # removed under /webhook only
          location: /webhook
        - id: 942430
          state: RULE_STATE_DETECTION_ONLY  # logged, never counted
        - id: 942130
          state: RULE_STATE_ENABLED         # forced on above paranoiaLevel
    sec_rules:                   # SecRules (Lite form)
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
        maxRequests: 500
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

- `security.core_ruleset` is a `coreruleset.v1.CoreRuleset`
  ([setting.proto](../api/proto/sentinez/types/coreruleset/v1/setting.proto)).
  `version` picks the embedded rule files (`VERSION_V4_16_0` only ships the
  RCE and SQLi groups; unset = `VERSION_V4_17_0`). `paranoiaLevel` and
  `detectionParanoiaLevel` are limited to 0–4; rule IDs in `exclusions` and
  `overrides` must be in 900000–999999; `location` must start with `/`.
- `categories` values map 1:1 to the CRS files: request groups
  `COMMON_EXCEPTIONS` (905), `METHOD_ENFORCEMENT` (911), `SCANNER_DETECTION`
  (913), `PROTOCOL_ENFORCEMENT` (920), `PROTOCOL_ATTACK` (921),
  `MULTIPART_ATTACK` (922), `LFI` (930), `RFI` (931), `RCE` (932), `PHP`
  (933), `GENERIC` (934), `XSS` (941), `SQLI` (942), `SESSION_FIXATION`
  (943), `JAVA` (944); response groups `DATA_LEAKAGES` (950),
  `DATA_LEAKAGES_SQL` (951), `DATA_LEAKAGES_JAVA` (952), `DATA_LEAKAGES_PHP`
  (953), `DATA_LEAKAGES_IIS` (954), `WEB_SHELLS` (955), `DATA_LEAKAGES_RUBY`
  (956). All are prefixed with `CATEGORY_`.
- `overrides[].state`: `RULE_STATE_DISABLED` can be scoped to a `location`;
  `RULE_STATE_DETECTION_ONLY` and `RULE_STATE_ENABLED` always apply to the
  whole namespace.
- The "Lite" types (`SecRuleLite`, `cdn.RuleLite`, `ExpressionLite`) use
  strings for `source/operator/status/value`. `NormalizeEdgeSetting` converts
  them to the runtime types (`SecRule`, `cdn.Rule`, `Expression`) with real
  enums, turns `value` into a string `structpb.Value`, and generates IDs
  `srule_<uuid>` / `cond_<nanoid>`.
- Variables supported in `proxySetHeaders`
  ([core/szhttp/variable](../staging/src/github.com/sentinez/core/szhttp/variable/variable.go)):
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
| `reverseproxy.ReverseProxy` | `Map[string,szhttp.ReverseProxy]` | Key = upstream server |
| `secrules.SecRule` | `Map[string,[]Entry]` | Compiled `Entry{Eval, Rule}`, active only, sorted by priority |
| `ratelimiter.Limiter` | `Map[string,*RateLimiter]` | One limiter per namespace |
| `rulesets.Rulesets` | `Map[string,coraza.WAF]` | One WAF instance per namespace |
| `cdnrules.Rule` | `Map[string,EvalFunc]` + `Map[string,*cdn.CDN]` | One CDN rule per namespace |

All maps use `shared/szsync.Map[K,V]` (a generic wrapper around `sync.Map`).

## 3.8 Cluster (Olric)

[pkg/cluster](../pkg/cluster/):

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
as `edge.Service` (`pkg/apps/dmz/edge/edge_service.go`). `main` starts it as
the `grpcServer` app on `defaults.EdgeAddress`; `EvaluateRuleset` evaluates an
empty `Expression`.

## 3.10 Choosing the engine

The engine is picked at startup with the `--engine` flag (default `quic`);
`edge.NewServer` switches on `conf.Flag.Engine`:

| `--engine` | Constructor | Server | Protocols |
|---|---|---|---|
| `quic` (default) | `engine.Quic` | `quichttpx.Server` | HTTP/3 (UDP) + HTTP/1.1/2 (TCP), needs a certificate |
| `hertz` | `engine.Hertz` | `httphz.XServer` | HTTP/1.1/2 (TCP only) |
| `std` | `engine.Standard` | `stdhttpx.Server` (`net/http`) | HTTP/1.1/2 (TCP only) |

```sh
./cmd/szedge/v1/bin/main --engine=hertz ...
```

Values outside `std|hertz|quic` are rejected by protovalidate when flags are
parsed; an empty value falls back to `Standard`. The startup log shows the
engine in use (`https[quic] running on 0.0.0.0:7443`). The
`stdhttpx.Server` sets 15s read/write and 120s idle timeouts when TLS is on,
and uses `ConnContext` for `OnStdConnect`. The edge `Server.Start` always
opens both the TCP and the UDP listener, whichever engine is used.
