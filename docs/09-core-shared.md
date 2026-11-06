# 9. Framework & shared libraries

## 9.1 `core/szrunner`

[core/szrunner](../staging/src/github.com/sentinez/core/szrunner/) wraps
`uber/fx` to standardize service lifecycles and run several apps in one
binary.

| API | Description |
|---|---|
| `New(conf, scope)` | Creates the console logger for gRPC, sets the log level from `--log_level`, installs the OTLP exporter; returns a `*Runner` and an error |
| `Runner.Run(ctx, apps...)` | Starts the apps in order, serves until `ctx` is done, `SIGINT/SIGTERM`, or a `Serve` fails, then stops them in reverse order; returns the joined errors |
| `NewApp[T](setup...)` | Builds an `*App`; `T` is inferred from `setup func(*Context[T])` |
| `Context.Inject(fns...)` | `fx.Provide` |
| `Context.Invoke(fn)` | `fx.Invoke` |
| `Context.OnStart(func(ctx, *T) error)` | Start hook; runs synchronously in registration order and must not block |
| `Context.Serve(func(ctx, *T) error)` | Blocking main loop, started after every app has started; `http.ErrServerClosed` counts as a clean exit, any other error stops the whole runner |
| `Context.OnStop(func(ctx, *T) error)` | Stop hook; runs in reverse registration order |

Each app has its own fx container, so apps in the same process can provide
the same types (for example `config.Config`). Stop hooks get a 15 s timeout.
In any mode other than `dev`, fx logging is disabled (`fx.NopLogger`).

OTLP: system logs go to `SENZ_OTLP_INTERNAL_ENDPOINT` and are not exported
when it is unset; service name = `meta.service_key`. The exporter is shut
down after every app has stopped. The edge traffic logs (func chain `LOG`
node) go straight to Kafka when `SENZ_KAFKA_BROKERS` is set (raw proto
events, one topic per log type, see `shared/szkafka`), otherwise through their
own OTLP provider on `SENZ_OTLP_EXTERNAL_ENDPOINT`; the exporter is owned by
the edge server and flushed in its `Shutdown`. With neither set they are not
exported. An `https://` endpoint uses TLS.

## 9.2 `core/szhttp` and `core/szhttp/chains`

- `szhttp.Server`: `ListenAndServe`, `Shutdown`, `Use(middleware)`,
  `Handle(RequestHandler)`, `AcceptReverse(target)`. The Hertz server
  (`contrib/httphz`) wraps the handler with the `Use` middleware once, when
  it starts, so `Use` may come before or after `Handle`.
- `ServerOption`: `WithCertificate`, `WithTLSConfig`, `WithServerName`,
  `WithOnStdConnect`, `WithStdListener` (TCP), `WithOnQuicConnect`,
  `WithQuicListener` (`http3.QUICListener`).
- `DecoreServer(conf, s)`: a decorator that prints the ASCII banner
  (`console.INFO`) before listening.
- Error helpers: `Forbidden`, `BadRequest`, `NotFound`, `TooManyRequests`,
  `InternalServerError` (rendered with templ).
- `chains.Plugin{ID, Handle(ctx, next)}`: a chain plugin (Caddy-style). It
  is a zero value that pulls its dependencies in `Provision`
  (`Provisioner[E]`, implemented only when needed) and receives `next` on
  every call; `Cleaner.Cleanup` releases resources.
  `chains.Build(env, wrap, plugins...)` provisions the plugins with `env`,
  links them in order (each wrapped by `wrap` when set) and returns a
  `*Chain` (`Handle`, `Close` cleaning up in reverse). The legacy
  `ChainNode`/`Node` API is unused.
- `core/szhttp/variable`: nginx-style variables for proxy headers.
- `core/szhttp/httpconst`, `core/szbytestr`: header/value constants (as
  `string` and `[]byte`).

## 9.3 `core/szgrpc`

- `szgrpc.New(opts...)`: `WithXMeta`, `WithDiscorvery(bool)`,
  `WithGRPCServerOption(...)`. `Serve(conf, opts...)` listens on TCP
  `SENZ_ADDRESS` (`WithAddress(addr)` overrides it), prints the banner,
  optionally registers the listen address with Consul. `BufServe(bufLis)`
  serves over bufconn. `Shutdown` = `GracefulStop`.
- `core/szgrpc/grpcgateway`: see [06-gateway.md](06-gateway.md).

## 9.4 Other `core` packages

| Package | Contents |
|---|---|
| `core` | Constants `Name` (`sentinez/1.0`, used as the `Server` header), `Code` (`SENZ`), `Version`, request ID prefix, `EnvPattern` |
| `core/szcontext` | `WithTransportValue` / `GetTransport` for `netpb.Transport` |
| `shared/sznormalize` | `NormalizeEdgeSetting` (Lite → runtime) |
| `core/internal/console` | ASCII banner (`go-figure`) printed once at startup |
| `core/internal/render` | `templ` templates for the 400/403/404/429/500 pages |
| `shared/szstorage/cache/mem` | Generic cache on top of `patrickmn/go-cache` |
| `shared/szstorage/dbx` | See [07-controlplane.md](07-controlplane.md#71-storage-layer-sharedszstoragedbx) |
| `shared/szstorage/dbx/timescale` | Hypertables and continuous aggregates, see [07-controlplane.md](07-controlplane.md#timescaledb-hypertables-and-continuous-aggregates) |
| `shared/szstorage/{cache/keyval,dbx/clickhouse,search/elastic}` | Empty packages (planned) |
| `core/szlimiter`, `core/szrule`, `core/szcrs`, `core/szmodsec` | See [04-rule-engine.md](04-rule-engine.md) |
| `shared/sztracer` | `StageError{Stage, Elapsed, Err}`, `Origin(err)` |
| `shared/szqueue` | Empty `Queue{}` (for the waiting room) |
| `shared/szsdklog` | `New(conf, opts...)`: `Exporter` of the edge traffic logs: Kafka on `SENZ_KAFKA_BROKERS`, else OTLP on `SENZ_OTLP_EXTERNAL_ENDPOINT`, nil when neither is set; `opts` (`szkafka.ConfigOption`) configure the producer |
| `shared/szkafka` | Kafka producer of the edge events and batch consumer, see [`szkafka`](#szkafka) |

## 9.5 `shared`

| Package | Contents |
|---|---|
| `szlog` | Logging (details below) |
| `szconfig` | `LoadEnv(file, keys...)` loads `.env` with godotenv and snapshots `SENZ_*` values into a global map; `GetEnv`/`SetEnv` are mutex-protected |
| `szflag` | Common flags `--mode`, `--log_level`; `Parse(meta)` (usage prints the banner); `Validate` with protovalidate |
| `szerror` | Standardized gRPC statuses — see [07-controlplane.md](07-controlplane.md#77-error-codes) |
| `szperms` | `Allow(XMethod, ControlPlane)` |
| `szsync` | `Map[K,V]` (generic `sync.Map`), `Pool[T]` (`NewPool`, `NewPoolCtr`) |
| `szrand` | `RandomString` (crypto/rand), `NewUUID` (record keys, never zoned), `NewNanoId`, `NewULID` (request/event ids) — each takes a `szprefix.Prefix` and joins it with `szprefix.Separator` |
| `bytesconv` | Fork of fasthttp/Hertz: zero-copy `B2s`/`S2b`, character tables, number parsing |
| `szjson` | `Marshal`/`Unmarshal` via `bytedance/sonic` |
| `szprotobuf` | Timestamp/Duration conversion, `Compare` for go-cmp, `Validate` |
| `szprotobuf/protox` | Binary `Marshal`/`Unmarshal`, `Struct(any)` → `structpb.Struct` |
| `szcopier` | Copies proto/JSON values via marshaling |
| `szcron` | `Start(ctx, interval, job)`: runs the job immediately, then on a ticker until ctx is cancelled |
| `szeventq` | In-process pub/sub per namespace on channels (no callers yet) |
| `szstore/ja4` | JA4 fingerprint cache keyed by connection ID (24h TTL) |
| `szcolor` | ANSI colors for console logs |
| `sztopic` | `Topic` type and the topic names of the edge traffic events (`EventHTTPIngress`, `EventHTTPEgress`, `EventHTTPSecRule`, ...) |

### `szlog`

Two kinds of logger:

1. **Console (sugar)** — used through the global functions
   `szlog.Info/Debugf/Errorf/Fatal...`. Colored console encoder; the level is
   set by `SetScopeLogLevel`.
2. **Structured events** — `NewLog(name, kind, level)` /
   `NewLogCloser(...)` write JSON to stdout with a `kind` field (`LogType`)
   and an `event` field (the proto marshaled via `protoreflect`). `LogCloser`
   calls `Close()` on the object after logging (returning it to its pool).

Both are teed into the `otelzap` bridge, so once `SetupOTLP` has run, logs
are exported over OTLP/gRPC (`NewOTLPProvider` + batch processor). A single
logger can use its own provider via `WithOTLPProvider(p)`.

Log levels: `debug=0, info=1, warn=2, error=3, fatal=4`; a record is written
when `level >= verbosity`.

### `szkafka`

[shared/szkafka](../staging/src/github.com/sentinez/shared/szkafka/) talks to Kafka with franz-go.

**Producer** (`NewProducer(Config)`):

- A `szlog.EventExporter` producing each event (vtproto bytes) to the
  topic of its `LogType` (`DefaultTopics()`), with the metadata in the
  `sz-kind`/`sz-level`/`sz-msg`/`sz-service` headers. It never blocks and
  counts the drops (`Dropped`).
- On creation it creates the missing topics it produces to (kadm, 10 s
  timeout), with `Partitions` partitions (zero takes
  `DefaultTopicPartitions` = 6, also the most collector replicas sharing a
  topic) and `ReplicationFactor` replicas (zero takes the broker default).
  An existing topic is left as is. A failed creation is only logged: the
  producer still starts and drops its events until the topics exist.
- `NewProducer` therefore reaches the brokers at once: with Kafka
  unreachable, the edge start waits up to 10 s.
- `ConfigOf(brokers, service, opts...)` builds the `Config`, with
  `WithTopics(map)` (copied), `WithTopicPartitions(n)`,
  `WithReplicationFactor(n)`. `Brokers(s)` splits a comma-separated broker
  list.

**Consumer** (`NewConsumer(ConsumerConfig)`, `Run(ctx, BatchHandler)`,
`Close()`):

- Consumes the `Topics` in the consumer group `Group`, from the earliest
  offset when the group has no commit yet.
- At least once: a batch is committed only after its handler returns nil
  (`AutoCommitMarks` + `MarkCommitRecords`; the marks are committed
  periodically, on partition revoke and on `Close`). A handler error
  retries the same batch, with a backoff from 0.5 s to 30 s, until `ctx`
  is done; no other batch is polled meanwhile.
- `Close` commits the marks, leaves the group and closes the client; it can
  be called more than once.

The limits bound the memory of the fetched records not handled yet,
however many records wait on the topics; a zero field takes its default:

| `ConsumerConfig` field | Default | franz-go default |
|---|---|---|
| `MaxPollRecords` | 1000 | unbounded |
| `FetchMaxBytes` | 16 MiB | 50 MiB |
| `FetchMaxPartitionBytes` | 1 MiB | 1 MiB |
| `MaxConcurrentFetches` | 3 | unbounded (one per broker) |
| `FetchMaxWait` | 500 ms | 5 s |
| (broker read limit) | 2 × `FetchMaxBytes` | 100 MiB |

The bound is about `MaxConcurrentFetches × FetchMaxBytes` = 48 MiB of
compressed data; decompressed (the producer uses zstd) it is larger.

## 9.6 `pkg/` (root module)

| Package | Contents |
|---|---|
| `pkg/apps/**/config` | Singleton `Config()`: parses flags + loads env + attaches `XMeta` |
| `pkg/apps/**/flags` | Service-specific flags, default `--env_file` |
| `pkg/network` | `StdListen` (TCP, `StdConn`s with IDs), `QuicListen` (UDP, needs `WithTLSConfig`, adds the `h3` ALPN, own `quic.Transport` that gives each connection an ID read with `ConnId(ctx)`), `StandardTransporter` (30s dial timeout), `QuicTransporter` (HTTP/3 client, unused), `GetInterface` |
| `pkg/network/httpx` | `szhttp.Context` and `ReverseProxy` implemented with `net/http`, shared by both servers below |
| `pkg/network/httpx/std` | `szhttp.Server` on `net/http` (TCP) |
| `pkg/network/httpx/quic` | `szhttp.Server` serving HTTP/3 (UDP) and HTTP/1.1/2 (TCP) on one address, with `Alt-Svc` |
| `pkg/network/wsz` | WebSocket server + client manager |
| `pkg/pools/{ingress,egress,secrule,coreruleset,matchedrule}` | Pools of the edge events (`httppb.IngressEvent`, `httppb.EgressEvent`, `secrulepb.RuleEvent`, `corerulesetpb.CoreRuleEvent`, `rulepb.MatchedRules`), reset with `ResetVT` |
| `pkg/protocol` | `Upstream2Target`: `PROXY_PROTOCOL_HTTP(S)` + server → URL |
| `pkg/cluster` | Embedded Olric; `SyncMap[T]` syncs a local map between edges, see [03-edge.md](03-edge.md#38-cluster-olric) |
| `pkg/cluster/leader` | Zone leader election through a Consul session lock; the leader keeps the zone status in Consul KV with fenced writes (`api/client/consul` `PutFenced`), see [03-edge.md](03-edge.md#zone-leader-election) |

## 9.7 `contrib/httphz`

Hertz adapter for `szhttp`:

- `NewServer(conf)`: `XServer` implements `szhttp.Server` with Hertz (TLS,
  ALPN, H2C, HTTP/2, body streaming, custom transport).
- `Context`: implements `szhttp.Context` on top of `*app.RequestContext`.
- `proxy`: HTTP(S) and WebSocket reverse proxies built on
  `hertz-contrib/reverseproxy`, with functional options (`WithTLS`,
  `WithTimeout`, `WithStreamResponseBody`, ...).
- `net/std`: transport/connection forked from Hertz (CloudWeGo license) to
  expose `TLSConn` so the original `network.StdConn` can be recovered.

## 9.8 `tools`

| Tool | Description |
|---|---|
| `protoc-gen-go-senz` | protoc plugin handling `x_meta`, `x_message`, `x_method` |
| `szruleparser` | `-file <crs.conf> -out <dir>` → Go file containing base64-encoded CRS rules |
| [tools.go](../staging/src/github.com/sentinez/tools/tools.go) | Pins tool versions: templ, buf, bpf2go, bombardier, protoc-go-inject-tag, grpc-gateway, openapiv2, vtproto, sqlc, mockery, protoc-gen-go(-grpc) |
