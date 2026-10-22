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
node) use their own provider on `SENZ_OTLP_EXTERNAL_ENDPOINT`, owned by the
edge server and flushed in its `Shutdown`; unset, they are not exported
either. An `https://` endpoint uses TLS.

## 9.2 `core/szhttp` and `core/szhttp/szchains`

- `szhttp.Server`: `ListenAndServe`, `Shutdown`, `Use(middleware)`,
  `Handle(RequestHandler)`, `AcceptReverse(target)`.
- `ServerOption`: `WithCertificate`, `WithTLSConfig`, `WithServerName`,
  `WithOnStdConnect`, `WithStdListener` (TCP), `WithOnQuicConnect`,
  `WithQuicListener` (`http3.QUICListener`).
- `DecoreServer(conf, s)`: a decorator that prints the ASCII banner
  (`szconsole.INFO`) before listening.
- Error helpers: `Forbidden`, `BadRequest`, `NotFound`, `TooManyRequests`,
  `InternalServerError` (rendered with templ).
- `szchains.ChainNode{SetNext, Handle}` + `Node{next}` with `HandleNext`,
  `GetNext` — the basis of the edge middleware chain.
- `core/szhttp/szvariable`: nginx-style variables for proxy headers.
- `core/szhttp/szhttpconst`, `core/szcmn/szbytestr`: header/value constants (as
  `string` and `[]byte`).

## 9.3 `core/szgrpc`

- `szgrpc.New(opts...)`: `WithXMeta`, `WithDiscorvery(bool)`,
  `WithGRPCServerOption(...)`. `Serve(conf)` listens on TCP `SENZ_ADDRESS`,
  prints the banner, optionally registers with Consul. `BufServe(bufLis)`
  serves over bufconn. `Shutdown` = `GracefulStop`.
- `core/szgrpc/szgrpcgateway`: see [06-gateway.md](06-gateway.md).

## 9.4 Other `core` packages

| Package | Contents |
|---|---|
| `core` | Constants `Name` (`sentinez/1.0`, used as the `Server` header), `Code` (`SENZ`), `Version`, request ID prefix, `EnvPattern` |
| `core/szcontext` | `WithTransportValue` / `GetTransport` for `netpb.Transport` |
| `core/szcmn` | `NormalizeEdgeSetting` (Lite → runtime) |
| `core/szcmn/szconsole` | ASCII banner (`go-figure`) printed once at startup |
| `core/szcmn/szrender` | `templ` templates for the 400/403/404/429/500 pages |
| `core/szstorage/szcache/szmem` | Generic cache on top of `patrickmn/go-cache` |
| `core/szstorage/szdbx` | See [07-controlplane.md](07-controlplane.md#71-storage-layer-coreszstorageszdbx) |
| `core/szstorage/{szcache/szkeyval,szdbx/szclickhouse,szmsgx/szkafka,szsearch/szelastic}` | Empty packages (planned) |
| `core/szlimiter`, `core/szrule`, `core/szrs`, `core/szmodsec` | See [04-rule-engine.md](04-rule-engine.md) |

## 9.5 `shared`

| Package | Contents |
|---|---|
| `zlog` | Logging (details below) |
| `config` | `LoadEnv(file, keys...)` loads `.env` with godotenv and snapshots `SENZ_*` values into a global map; `GetEnv`/`SetEnv` are mutex-protected |
| `flagx` | Common flags `--mode`, `--log_level`; `Parse(meta)` (usage prints the banner); `Validate` with protovalidate |
| `errorx` | Standardized gRPC statuses — see [07-controlplane.md](07-controlplane.md#77-error-codes) |
| `perms` | `Allow(XMethod, ControlPlane)` |
| `sync` | `Map[K,V]` (generic `sync.Map`), `Pool[T]` (`NewPool`, `NewPoolCtr`) |
| `rand` | `RandomString` (crypto/rand), `NewId` (UUID), `NewNanoId`, `NewXId`, `NewTimeId` (ULID) — all with prefixes; results are copied out of the pooled buffer |
| `bytesconv` | Fork of fasthttp/Hertz: zero-copy `B2s`/`S2b`, character tables, number parsing |
| `jsonx` | `Marshal`/`Unmarshal` via `bytedance/sonic` |
| `protobuf` | Timestamp/Duration conversion, `Compare` for go-cmp, `Validate` |
| `protobuf/protox` | Binary `Marshal`/`Unmarshal`, `Struct(any)` → `structpb.Struct` |
| `copier` | Copies proto/JSON values via marshaling |
| `cron` | `Start(ctx, interval, job)`: runs the job immediately, then on a ticker until ctx is cancelled |
| `eventq` | In-process pub/sub per namespace on channels (no callers yet) |
| `store/ja4` | JA4 fingerprint cache keyed by connection ID (24h TTL) |
| `color` | ANSI colors for console logs |
| `topic` | Empty package |

### `zlog`

Two kinds of logger:

1. **Console (sugar)** — used through the global functions
   `zlog.Info/Debugf/Errorf/Fatal...`. Colored console encoder; the level is
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

## 9.6 `pkg/` (root module)

| Package | Contents |
|---|---|
| `pkg/apps/**/config` | Singleton `Config()`: parses flags + loads env + attaches `XMeta` |
| `pkg/apps/**/flags` | Service-specific flags, default `--env_file` |
| `pkg/network` | `StdListen` (TCP, `StdConn`s with IDs), `QuicListen` (UDP, needs `WithTLSConfig`, adds the `h3` ALPN), `StandardTransporter` (30s dial timeout), `QuicTransporter` (HTTP/3 client, unused), `GetInterface` |
| `pkg/network/httpx` | `szhttp.Context` and `ReverseProxy` implemented with `net/http`, shared by both servers below |
| `pkg/network/httpx/std` | `szhttp.Server` on `net/http` (TCP) |
| `pkg/network/httpx/quic` | `szhttp.Server` serving HTTP/3 (UDP) and HTTP/1.1/2 (TCP) on one address, with `Alt-Svc` |
| `pkg/network/wsz` | WebSocket server + client manager |
| `pkg/pools/{request,ruleevent}` | Pools for `httppb.Event`, `secrulepb.Event` |
| `pkg/protocol` | `Upstream2Target`: `PROXY_PROTOCOL_HTTP(S)` + server → URL |
| `pkg/tracer` | `StageError{Stage, Elapsed, Err}`, `Origin(err)` |
| `pkg/queue` | Empty `Queue{}` (for the waiting room) |

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
| `ruleparser-sentinez` | `-file <crs.conf> -out <dir>` → Go file containing base64-encoded CRS rules |
| [tools.go](../staging/src/github.com/sentinez/tools/tools.go) | Pins tool versions: templ, buf, bpf2go, bombardier, protoc-go-inject-tag, grpc-gateway, openapiv2, vtproto, sqlc, mockery, protoc-gen-go(-grpc) |
