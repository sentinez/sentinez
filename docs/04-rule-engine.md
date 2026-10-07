# 4. Rule engine & WAF

Sentinéz has two independent request-filtering mechanisms:

1. **SecRule / CDN rules** — user-defined expressions evaluated by
   `core/rules` (fast, no external dependency).
2. **Rulesets (WAF)** — the OWASP Core Rule Set running on Coraza, loaded
   from pre-generated Go code in `core/modsec/gen`.

## 4.1 Expression model

Defined in
[types/rule/v1/rule.proto](../api/proto/sentinez/types/rule/v1/rule.proto):

```
Expression
 └─ or_condition: []AndCondition        // OR across entries
      ├─ rules: []Rule                   // AND across rules
      │    └─ condition: Condition{source, key, operator, value}
      └─ or_condition: []AndCondition    // nested: AND(rules) && OR(children)
```

Semantics ([core/rules/rule_expr.go](../staging/src/github.com/sentinez/core/rules/rule_expr.go)):

- `execOrCond(nil) == true` — an empty expression **always matches**.
- An `AndCondition` matches when all its `rules` match **and** its child
  `or_condition` (if any) matches.
- IDs/names of matching rules are appended to `rulepb.MatchedRules` (if one
  is passed in).

API:

```go
eval := corerule.NewEval(expr)            // compile once
ok := eval(ctx corehttp.RequestContext, matched *rulepb.MatchedRules)
```

## 4.2 Field sources and operators

Matchers live in
[core/rules/rule_match.go](../staging/src/github.com/sentinez/core/rules/rule_match.go).

| Source | Read from | Supported operators |
|---|---|---|
| `PATH` | `ctx.Path()` | EQ, NE, CONTAINS, PREFIX, SUFFIX, MATCHES |
| `BODY` | `ctx.Body()` | same as above |
| `HOST` | `ctx.Host()` | same as above |
| `METHOD` | `ctx.Method()` | IN, NOT_IN (list) + the string group |
| `HEADER` | empty `key` or `"header"`: header **existence**; with `key`: compares the header value | Existence: IN/NOT_IN (list) or EQ/NE (single name). Value: list → membership (IN/EQ = present, NOT_IN/NE = absent); string → string group |
| `QUERY` | same as HEADER, for query params | same as HEADER |
| `IP` | `ctx.RequestIP()` | EQ, NE (IP or CIDR), IN, NOT_IN (list of IPs/CIDRs) |
| `TLS` | `"true"`/`"false"` | EQ, NE |
| `JA4` | TLS fingerprint | string group; no JA4 → no match |

Notes:

- `MATCHES` uses Go `regexp`; compiled regexes are cached globally by pattern
  (`sync.Map`). An invalid pattern never matches.
- `GT/GTE/LT/LTE` exist in the enum but are **not implemented** (never
  match).
- Unknown sources never match (`bypass = false`).
- In YAML, `value` is always a string (Lite form), so list operators
  (`IN`/`NOT_IN`) are only usable when building an `Expression` directly
  (e.g. via the API).

## 4.3 SecRule

Runtime type: `types.secrule.v1.SecRule{id, name, description, expr, action,
status, priority}`. At the edge
([internal/memory/secrules/secrule.go](../internal/memory/secrules/secrule.go)):

- Only `STATUS_ACTIVE` rules are kept.
- Rules are sorted by **descending priority**, stable (equal priorities keep
  declaration order).
- The `RUL` node walks the chain; the first matching rule with
  `ACTION_TYPE_BLOCK` → `403`. Other actions (`LOG`, `MODIFY_HEADER`,
  `REDIRECT`, `SET_TAG`, `ROUTE_TO`) are only debug-logged for now and not
  executed.

There are three SecRule shapes in the protos:

| Type | Used in |
|---|---|
| `apps.security.v1.SecRule` | API/DB model (control plane), with `metadata`, `action` as a string + `action_value` |
| `types.secrule.v1.SecRule` / `SecRuleLite` | Edge runtime / YAML input form |
| `dmz.edge.v1.SecRule` | Wrapper inside `Setting`: `{ingress: Lite, ingress_runtime: Runtime}` |

## 4.4 CDN rules

`types.cdn.v1.CDN{rule: RuleLite, rule_runtime: Rule}` lives in
`setting.controller.cdn`. Each namespace keeps **one** active CDN rule (a
later rule overwrites an earlier one). The `CDN` node only caches when the
rule is active and its expression matches. The cache is
`patrickmn/go-cache` with a 1h TTL and only stores responses with status
`< 400`.

## 4.5 Rate limiter

[core/limiter](../staging/src/github.com/sentinez/core/limiter/) implements
the **sliding window counter** algorithm:

```
count = curr.count + prev.count × (size − elapsed_in_curr) / size
allow  ⇔ count + n ≤ limit
```

- `SlidingWindow` holds two `LocalWindow`s (prev/curr); `advance` shifts the
  windows using `now.Truncate(size)`.
- `RateLimiter` keeps one `SlidingWindow` per key (IP) and a `timeout`
  mechanism: after a key exceeds the limit, every request from that key is
  rejected for `timeout` ms. Other keys are not affected.
- `Window.Sync` is a hook for syncing with a central store in the future
  (currently a no-op).

Configured in `security.limiters[]`: `timeWindow`, `limit`, `timeout`
(`time.ParseDuration` strings).

## 4.6 WAF — OWASP CRS on Coraza

### Rule generation

The original CRS (`.conf`) lives in [deploy/ruleroot](../deploy/ruleroot/).
The `ruleparser-sentinez` tool
([tools/cmd/ruleparser-sentinez](../staging/src/github.com/sentinez/tools/cmd/ruleparser-sentinez/main.go)):

1. Parses SecLang with an ANTLR parser
   ([core/modsec/ruleparser](../staging/src/github.com/sentinez/core/modsec/ruleparser/)):
   each statement is a rule; a rule ending in `chain"` is joined with the
   next statement; `ver`, `paranoia-level` and action fields are extracted.
2. Loads the result into `coreruleset.v1.CoreRulesets` and renders a Go
   template: each rule becomes a function returning a `*CoreRule` whose
   `Configuration` is base64-encoded, plus an `XxxOrder` slice that keeps the
   order.
3. Writes `core/modsec/gen[/v4-16-0|/v4-17-0]/*.sentinez_rules.gen.go`.

Regenerate everything with `bash hack/ruleparser.sh`.

### Assembling directives

`corers.GenerateRulesets(version, flag)`
([waf_loader.go](../staging/src/github.com/sentinez/core/rulesets/waf_loader.go))
concatenates, in order:

1. `setup` → 2. `REQUEST-901-INITIALIZATION` → 3. version + flag rules →
4. `audit` → 5. `default` → 6. `REQUEST-949-BLOCKING-EVALUATION`.

`Flag` is a bitmask (`waf_flag.go`) with one bit per CRS file (905…944,
950…980). Only two flags have generated rules so far:

| Flag | v4.16.0 | v4.17.0 |
|---|---|---|
| `ReqAppAttackRCE` (932) | yes | yes |
| `ReqAppAttackSQLI` (942) | yes | reuses the v4.16.0 rules |

The result is also written to `WAF.conf.lock` in the working directory (handy
for inspecting the final directives).

`corers.NewWAF(version, fs, flag)` builds a `coraza.WAF` whose `RootFS` is the
set of `*.data` files embedded in [sentinez.go](../sentinez.go) (`WAF4160()`,
`WAF4170()`).

### Per-request processing

[core/rulesets](../staging/src/github.com/sentinez/core/rulesets/) — modeled
on Coraza's HTTP middleware:

| Step | Function | Details |
|---|---|---|
| Create transaction | `NewRulesets(ctx, waf)` | Taken from a `sync.Pool`; uses `NewTransactionWithOptions` when the WAF supports a context |
| Ingress | `ExecIngress` | `ProcessConnection` (IP:port), `ProcessURI`, adds all headers + `Host` + `Transfer-Encoding`, `ProcessRequestHeaders`, reads the body if allowed, `ProcessRequestBody`. On interruption sets the status (403 by default for a `deny` action) |
| Egress | `ExecEgress` | Adds response headers, `ProcessResponseHeaders`, writes the body when processable, `ProcessResponseBody`, copies the WAF-processed body |
| Finish | `Final(cb)` | `ProcessLogging`, the event-logging callback, `tx.Close()` |

WAF event (`secrulepb.Event`): matched rule IDs/severities/messages (except
the interrupting rule), `score` = data of the interrupting rule (anomaly
score), IP, domain, transaction ID, request ID, time.

### Enabling the WAF for a namespace

`MemStore.LoadRulesets` creates a WAF for each entry in the namespace's
`security.rulesets[]`. It currently always uses **CRS v4.16.0** with the
`ReqAppAttackRCE` flag and does not read options from the config.
