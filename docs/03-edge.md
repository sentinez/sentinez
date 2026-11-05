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
| [pkg/cluster/](../pkg/cluster/) | Embedded Olric + `SyncMap[T]` (memory synced between edges) |
| [pkg/cluster/leader/](../pkg/cluster/leader/) | Zone leader election (Consul session lock) |
| [contrib/httphz](../staging/src/github.com/sentinez/contrib/httphz/) | Hertz server + `szhttp.Context` + reverse proxy |
| [pkg/network/](../pkg/network/) | TCP (`StdListener`/`StdConn`, with IDs) and QUIC (`QuicListener`) listeners, `net/http` transport |
| [pkg/network/httpx/](../pkg/network/httpx/) | Shared `net/http` `Context` + `ReverseProxy`; servers in `std/` (`stdhttpx`) and `quic/` (`quichttpx`) |

## 3.1 Startup

```
main: run() → runner.Run(ctx, NewApp(edgeServer), NewApp(grpcServer), NewApp(leader), NewApp(pprof))
 ├─ edgeServer
 │   ├─ Inject(config.Config, edgeyaml.LoadSetting, edge.NewServer)
 │   └─ Serve(server.Start)
 ├─ grpcServer
 │   ├─ Inject(config.Config, edge.NewService)
 │   └─ Serve(service.Start)   // gRPC on defaults.EdgeAddress
 └─ leader
     ├─ Inject(config.Config, leader.New)
     └─ Serve(e.Run(ctx, e.PublishStatus()))  // zone leader, see 3.8

edge.NewServer(conf)
 ├─ edgeyaml.LoadSetting(conf)              // raw *Lite setting of proxy.yaml
 ├─ memory.NewMemStore(conf)                // singleton; cluster.NewCluster + settings SyncMap
 └─ switch conf.Flag.Engine                 // --engine
     hertz → engine.Hertz(conf)            → httphz.NewServer + WithOnStdConnect(OnHertzConnect)
     quic  → engine.Quic(conf)             → quichttpx.NewServer + WithOnStdConnect(OnStandardConnect)
     std, other → engine.Standard(conf)    → stdhttpx.NewServer + WithOnStdConnect(OnStandardConnect)

Server.Start()
 ├─ initialize()
 │   ├─ mem.Sync(ctx, server, setting)  // 30s: wait for Olric, load the cluster's settings,
 │   │                                  // share this one; each change rebuilds routes,
 │   │                                  // reverse proxies, CDN, SecRule, limiter, WAF
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
| `httpreq.RequestContext` | `core/szhttp/httpreq/context_value.go` | Evaluating rules against an `httppb.Request` (only tests use it now) |

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
| 1 | `LOG` | `logging.Logging` | Calls `next` first, then logs an `httppb.Event` (JSON, kind `LOG_TYPE_HTTP`) |
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

- `Store(server)`: copies locations, skips (with a warning) those without
  `proxyPass`, filters valid headers, sets the default `proxyRewrite`, and
  **sorts by `location` length, descending**, for nginx-style
  longest-prefix matching. A server left with no location is unloaded.
- `Match(ctx)`: loads the routes for the namespace, finds the first location
  that matches the path **on a segment boundary** (`/api` matches `/api` and
  `/api/x` but not `/apix`; a location ending in `/` matches any path under
  it), sets the new path to `rewrite + (path − location)` without doubling
  the `/` between them, sets headers, and returns the next upstream of
  `proxyPass`, **round robin** (per location, atomic counter; the only
  `balanceStrategy`, also used when it is unspecified).

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
  Requests with `Upgrade: websocket` go to `WSReverseProxy`, which builds
  the target (`upstream + URI`) per request.
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
| `settings.Setting` | `Map[string,*edgepb.Setting]` | Normalized Settings by namespace, rebuilt from the synced settings (3.8) |
| `routes.Router` | `Map[string,[]*Location]` | Sorted locations |
| `reverseproxy.ReverseProxy` | `Map[string,szhttp.ReverseProxy]` | Key = upstream server |
| `secrules.SecRule` | `Map[string,[]Entry]` | Compiled `Entry{Eval, Rule}`, active only, sorted by priority |
| `ratelimiter.Limiter` | `Map[string,*RateLimiter]` | One limiter per namespace |
| `rulesets.Rulesets` | `Map[string,coraza.WAF]` | One WAF instance per namespace |
| `cdnrules.Rule` | `Map[string,EvalFunc]` + `Map[string,*cdn.CDN]` | One CDN rule per namespace |

All maps use `shared/szsync.Map[K,V]` (a generic wrapper around `sync.Map`).

## 3.8 Cluster (Olric)

[pkg/cluster](../pkg/cluster/):

- `NewCluster` creates Olric with the `local` config. If the config is
  `Clustered()` (both `SENZ_MEMBERSHIP_ADDRESS` and `SENZ_DISCOVERY_ADDRESS`
  are set), it sets `Peers` and binds memberlist to the discovery address.
- Olric has no compare-and-swap, so it must not hold state a leader
  writes; that goes to Consul, where the writes are fenced (see below).
- `Cluster.Ready()` is closed once Olric accepts connections, clustered or
  not, and `Cluster.Node()` names the process (`<hostname>-<6 chars>`).

### Syncing memory between edges

`cluster.SyncMap[T]` ([syncmap.go](../pkg/cluster/syncmap.go),
[syncmap_run.go](../pkg/cluster/syncmap_run.go)) keeps a map of proto
values the same on every edge of the cluster, without putting Olric on the
request path. Measured on one node, an embedded Olric `Get` takes ~500 ns
against ~20 ns for a `sync.Map` `Load` (~1.3 µs against 2 ns on a hot key
read in parallel), and a key owned by another node costs a network round
trip on top.

| Operation | What it does |
|---|---|
| `Load`, `Range` | Read the local `szsync.Map` only: lock-free, never Olric |
| `Store(ctx, key, v)`, `Delete(ctx, key)` | Apply locally, `Put` the record in the DMap `<name>`, then `Publish` it on the Olric PubSub channel `sentinez.sync.<name>` |
| `Start(ctx)` | Wait for `Cluster.Ready()`, subscribe, then load the whole DMap (`Scan` + `Get`) |
| every 30s | Reconcile: take the DMap records newer than the local ones, write back the local ones the DMap lacks or holds older |
| `Close()` | Stop syncing; the local values stay |

- **Ordering.** Each write is a record `{version, origin, deleted,
  payload}`. `version` comes from a hybrid logical clock: wall nanoseconds
  that never go back and move past every version the node applied, so a
  node's next write orders after what it saw even with clock drift. Ties
  break on `origin` (the node). A record applies only if it is newer:
  last writer wins, and every node converges to the same value.
- **Lost messages.** Olric PubSub is at most once: a node that is not
  subscribed (restart, partition) misses the message. `Start` subscribes
  before it loads, so nothing written in between is lost, and the periodic
  reconcile repairs the rest, within 30s.
- **Concurrent writes.** Two nodes writing a key at once may leave the
  older record in the DMap (the `Put`s race); the reconcile of the node that
  holds the newer one writes it back.
- **Deletes** are tombstones: kept 24h in the DMap (`EX`) and locally, so an
  older value cannot come back.
- **Writes before `Start`** stay local; the first load shares them.

What it is not for: counters (two increments at once lose one: use
Olric `Incr` instead), and state that must be strictly ordered across nodes
(clock drift beyond the HLC bound can reorder concurrent writes): that goes
to Consul with fencing.

`OnChange(fn(key, prev, next))` reports each change of a key on the node,
whether written there, received or reconciled; `prev`/`next` is `nil` when
the key had or has no value. Calls are serialized in the order changes
apply.

### Settings between edges

The hot-path stores (routes, reverse proxies, CDN, SecRule, limiter, WAF)
hold compiled values that cannot cross the network (`coraza.WAF`, proxies
with their transports, limiters with their counters). So the edges sync
the **source**, the raw `edgepb.Setting`, and each node compiles its own
stores from it ([internal/memory/sync.go](../internal/memory/sync.go)):

- `MemStore` holds a `SyncMap` `sentinez.dmz.edge.settings` keyed by
  `metadata.resource_id`, the id of the domain resource. A setting without
  one is loaded on its edge only, with a warning.
- `MemStore.Sync(ctx, server, local...)` starts the map (waits for Olric,
  loads the cluster's settings), then shares `local`. `SyncMap.Store`
  skips a value equal to the one it holds (same deterministic encoding), so
  a restart does not make every edge rebuild it.
- On each change (`OnChange`), the node normalizes a copy
  (`sznormalize.NormalizeEdgeSetting`) and rebuilds the stores of its
  namespace (`server.name`), which requests still route by. A namespace
  that changed name, or a deleted resource, is unloaded first.
- A namespace belongs to one resource: a second resource claiming it is
  not loaded (logged), and cannot unload it.
- Requests keep reading the local `szsync.Map`s (~20 ns); Olric is never on
  the request path.

Rate-limiter counters and JA4 fingerprints are not synced: each edge limits
its own share of the traffic, and a TLS connection only lands on one edge.

### Zone leader election

[pkg/cluster/leader](../pkg/cluster/leader/) elects the one edge that leads
its zone (`SENZ_ZONE`). Olric is not used for this: its own docs call its
locks approximate (a lock lives on one partition owner, with no replica or
quorum by default and no monotonic token), so they cannot guarantee a single
leader. The election goes through a Consul session lock instead, which
Consul serves from its Raft log.

| Configuration | Behavior |
|---|---|
| `SENZ_CONSUL_URI` + `SENZ_ZONE` | Campaigns on the Consul key `sentinez/dmz/edge/leader/<zone>`; the zone is lowercased like the `szrand` id prefix, so `SGN` and `sgn` share one key |
| `conf.Clustered()` without `SENZ_CONSUL_URI` | Startup fails (`errNoConsul`): a cluster never runs without an election |
| `SENZ_CONSUL_URI` without `SENZ_ZONE` | Startup fails (`errNoZone`) |
| Neither | Standalone: the node leads at once (term 1) and never loses |

`conf.Clustered()` ([config.go](../api/proto/sentinez/types/setting/v1/config.go))
is true when both `SENZ_MEMBERSHIP_ADDRESS` and `SENZ_DISCOVERY_ADDRESS` are
set; `cluster.newConfig` uses the same check to join peers.

A term:

1. `consul.Client.Acquire` ([api/client/consul/lock.go](../api/client/consul/lock.go))
   creates a session (TTL 15s, renewed by its holder) and blocks until it
   holds the key. Each wait on a held lock is capped at 2s, so a follower
   notices shutdown well within the runner's 15s stop timeout.
2. The winner reads the `LockIndex` of the key with a consistent read: it
   becomes `Term.Index`, which grows each time the lock changes hands. The
   lock value is the node id (`<hostname>-<6 random chars>`); it is compared
   so the index belongs to this node.
3. `Elector.Run(ctx, task)` publishes the term (`Elector.Leader()` returns
   it) and runs `task(ctx, term)`.
4. The term ends when the lock is lost (`Lost()` closed, after 3 monitor
   retries), the task returns an error, or `ctx` ends. The elector clears
   `Leader()` **before** canceling the task's `ctx`, waits for the task,
   releases the lock, then campaigns again (after 5s if the term failed).

A task returning `nil` keeps the leadership until it is lost;
`Run(ctx, nil)` only holds it.

Failover:

- Graceful stop: the lock is released and another node leads within
  milliseconds.
- Crash or lost link to Consul: the session expires after its TTL, then
  Consul's lock-delay (15s) applies, so the zone has no leader for about 30s.

**Fencing.** A deposed leader can still act for a moment (GC pause,
partition) before it sees `Lost()`. Every write the leader makes to shared
state must carry `term.Index`, and the store must reject an index lower than
the highest it has seen. The elector cannot enforce this on its own;
`consul.Client.PutFenced` ([api/client/consul/fenced.go](../api/client/consul/fenced.go))
does, for Consul KV:

- The value is stored as JSON `{"token": <term.Index>, "value": ...}`
  (`consul.Fenced[T]`, read back with `consul.GetFenced[T]`).
- Each write reads the key (consistent read), fails with `consul.ErrFenced`
  if the stored token is higher than the writer's (equal is the same
  leader), then
  writes with a CAS on the `ModifyIndex` it read, so the check and the write
  are atomic. A CAS lost to another writer is retried a few times (`_casRetries`).
- Olric is not fenced: it has no compare-and-swap (only `Put` `NX`/`XX`,
  `GetPut`, `Incr` and an approximate `Lock`), so a token check there could
  still be raced. Shared state the leader writes lives in Consul instead.

**Zone records.** The leader keeps its zone under
`sentinez/dmz/edge/zone/<zone>/` ([zone.go](../pkg/cluster/leader/zone.go)),
every write through `PutFenced` with `term.Index`:

- `Elector.PublishStatus()` ([status.go](../pkg/cluster/leader/status.go))
  writes `<zone>/status` at once, then every 10s. It is the `Task` the edge
  runs.
- `Elector.Publish(name, value)` writes `<zone>/<name>` once per term,
  retried every 10s until written, for zone-wide records to come.
- `leader.All(tasks...)` ([task.go](../pkg/cluster/leader/task.go)) runs
  several tasks in one term; the first failure cancels the others and ends
  the term.

The `status` record:

| Field | Value |
|---|---|
| `zone` | Lowercased `SENZ_ZONE` |
| `node`, `term` | The leader and its `term.Index` |
| `updated_at` | UTC time of the write; a record older than ~10s means no leader is writing |

A fenced write (`ErrFenced`) ends the term: a newer leader wrote the record,
so the node gives up the lock and campaigns again. Other errors (Consul
unreachable) are logged and retried at the next tick. A standalone edge has
no Consul, so `PublishStatus` returns `nil` and it only leads.

**Wiring.** [cmd/szedge/v1/server/leader.go](../cmd/szedge/v1/server/leader.go)
runs it as its own szrunner app (`Inject(leader.New)`,
`Serve(e.Run(ctx, e.PublishStatus()))`). Each app has its own fx container, so the `Elector` is not reachable
from the `edgeServer`/`grpcServer` apps; zone-wide work goes in the `Task`.

**Deploy.** [deploy/docker/cluster/docker-compose.yaml](../deploy/docker/cluster/docker-compose.yaml)
runs two edges in zone `SGN` with a dev-mode Consul. Run 3 or 5 Consul
servers in production: without a Consul quorum no zone can elect a leader.

## 3.9 EdgeService gRPC

`edgepb.EdgeService` only has `Status`, which returns `OK`
([internal/dmz/edge/api.go](../internal/dmz/edge/api.go), package `edge`).
It is packaged as `edge.Service` (`pkg/apps/dmz/edge/edge_service.go`), and
`main` starts it as the `grpcServer` app on `defaults.EdgeAddress`. The
`EvalCoreRuleset` RPC (replaying a request against a CRS setting) was removed
together with `szcrs.Evaluate`. `internal/dmz/edge/events/{pub,sub}` are
empty packages for now.

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
