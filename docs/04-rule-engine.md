# 4. Rule engine & WAF

Sentinéz has two independent request-filtering mechanisms:

1. **SecRule / CDN rules** — user-defined expressions evaluated by
   `core/szrule` (fast, no external dependency).
2. **Rulesets (WAF)** — the OWASP Core Rule Set running on Coraza, loaded
   from pre-generated Go code in `core/szmodsec/gen`.

## 4.1 Expression model

Defined in
[types/rule/v1/rule.proto](../api/proto/sentinez/types/rule/v1/rule.proto):

```
Expression
 └─ or_condition: []AndCondition        // OR across entries
      ├─ rules: []Rule                   // AND across rules
      │    └─ condition: Condition{source, key, operator, value}
      ├─ or_condition: []AndCondition    // nested: AND(rules) && OR(children)
      └─ and_condition: []AndCondition   // AND(each), every one with its own groups
```

Semantics ([core/szrule/rule_expr.go](../staging/src/github.com/sentinez/core/szrule/rule_expr.go)):

- `execOrCond(nil) == true` — an empty expression **always matches**.
- An `AndCondition` matches when all its `rules` match **and** its child
  `or_condition` (if any) matches **and** every one of its `and_condition`
  matches (each is an `AndCondition`, with its rules and groups).
  `(a or b) and (c or d)` is `{or_condition: [a, b], and_condition:
  [{or_condition: [c, d]}]}`: an `and` holds as many or groups as it needs,
  the first in `or_condition`, the others in `and_condition`.
- `ValidateExpression` allows 256 conditions and 8 levels of groups
  (`or_condition` and `and_condition` both count a level).
- An edge older than `and_condition` ignores it, so such a rule matches
  more than it says (a `BLOCK` blocks more): deploy the edges before a rule
  uses it.
- IDs/names of matching rules are appended to `rulepb.MatchedRules` (if one
  is passed in). The name is `szrule.ExprString` of the expression: its Lua
  condition (`ctx:path() == "/login" and ctx:header("X-Debug") ~= ""`, from
  `rulescript.ToCondition`), or a text form (`path matches "("`) for one a
  script cannot write (an invalid regex). Groups keep their parentheses:
  `ctx:path() == "/a" and (ctx:method() == "GET" or ctx:method() == "PUT")`.

API:

```go
eval := szrule.NewEval(expr)            // compile once
ok := eval(ctx szhttp.RequestContext, matched *rulepb.MatchedRules)
```

## 4.2 Field sources and operators

Matchers live in
[core/szrule/rule_match.go](../staging/src/github.com/sentinez/core/szrule/rule_match.go).

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
- `GT/GTE/LT/LTE` compare both sides as numbers (`strconv.ParseFloat`) on
  the string sources (`PATH`, `BODY`, `HOST`, `METHOD`, `JA4`, a `HEADER`
  or `QUERY` value). A side that is not a number never matches.
- Unknown sources never match (`bypass = false`).
- In YAML, a scalar `value` is matched as a string (numbers and bools are
  converted by `NormalizeEdgeSetting`); the list operators (`IN`/`NOT_IN`)
  take a YAML list, e.g. `value: [GET, HEAD]`.

## 4.3 SecRule

Runtime type: `types.secrule.v1.SecRule{id, name, description, expr, action,
status, priority, script}`. At the edge
([internal/memory/secrules/secrule.go](../internal/memory/secrules/secrule.go)):

- Only `STATUS_ACTIVE` rules are kept.
- Rules are sorted by **descending priority**, stable (equal priorities keep
  declaration order).
- The `secrule` node
  ([internal/handlers/secure](../internal/handlers/secure/action.go))
  walks the chain in that order and keeps every matching rule:
  - **Terminal** actions end the request; the strongest one matched wins,
    `BLOCK` > `CHALLENGE` > `REDIRECT`, ties going to the higher priority.
    `BLOCK` stops the scan; `CHALLENGE` and `REDIRECT` do not, so a later
    `BLOCK` still applies to a client holding a challenge pass.
  - **Non-terminal** actions of every matched rule are applied, before the
    terminal one. When rules set the same header, tag or route, the higher
    priority wins.
  - One `secrule: matched` event is logged per request, with the terminal
    action (else the highest priority one) and the tags.

Actions and their `params` (`Action.params`, a `Struct`; numbers and
booleans are formatted as strings):

| Action | `params` | Effect |
|---|---|---|
| `BLOCK` | — | `403` page |
| `CHALLENGE` | — | Proof-of-work page until the client holds a `pow-pass` cookie |
| `REDIRECT` | `location` (required), `status` (301, 302, 303, 307, 308; default 302) | Redirects the client. Without a usable `location` the rule is ignored |
| `LOG` | — | Only the matched event |
| `MODIFY_HEADER` | header → value | Sets request headers sent upstream. A location's `proxySetHeaders` are applied later and win on the same header |
| `SET_TAG` | key → value | Adds to `ContextExtra.tags` and the matched event |
| `ROUTE_TO` | `location` | Proxies to the `server.locations` entry with that `location` instead of the longest matching prefix; an unknown one falls back to prefix matching |

```yaml
secrules:
  - ingress:
      name: "Send old links to the new shop"
      status: STATUS_ACTIVE
      expr: {orCondition: [{rules: [{condition: {source: FIELD_SOURCE_PATH, operator: OPERATOR_PREFIX, value: "/old-shop"}}]}]}
      action:
        type: ACTION_TYPE_REDIRECT
        params: {location: "/shop", status: 301}
  - ingress:
      name: "Tag and route large pages"
      status: STATUS_ACTIVE
      expr: {orCondition: [{rules: [{condition: {source: FIELD_SOURCE_QUERY, key: "page", operator: OPERATOR_GT, value: "100"}}]}]}
      action:
        type: ACTION_TYPE_SET_TAG
        params: {deep_paging: true}
```

The control plane stores the action value as `ActionValue{str_value,
map_value}` (console: one value for `REDIRECT`/`ROUTE_TO`, key/value rows for
`MODIFY_HEADER`/`SET_TAG`); there is no sync to the edge yet, so the edge
only reads `params` from `proxy.yaml`.

### Lua rule scripts

A secrule (or a rate limiter, see 4.5) may set `script`, a Lua source, in
place of `expr` and `action`; a rule with both, or with neither, is refused
when the setting loads. A secrule script defines a global `match(ctx)`:

```yaml
secrules:
  - ingress:
      name: "Script: block X-Debug"
      status: STATUS_ACTIVE
      script: |
        ---@param ctx sz.Context
        function match(ctx)
          if ctx:header("X-Debug") == "1" then
            return SZ_BLOCK
          end
          if ctx:path() == "/old" then
            return {type = SZ_REDIRECT, params = {location = "/new"}}
          end
        end
```

- `nil` or `false`: the rule does not match.
- An action constant, as an eBPF program returns `XDP_DROP`: `SZ_BLOCK`,
  `SZ_CHALLENGE`, `SZ_REDIRECT`, `SZ_LOG`, `SZ_MODIFY_HEADER`, `SZ_SET_TAG`,
  `SZ_ROUTE_TO` (`RATE_LIMIT` is not one). Each holds the action name
  (`SZ_BLOCK == "block"`), which a script may return as well.
- A table `{type = SZ_REDIRECT, params = {...}}`: the action with the
  `params` of the table above.

The edge runs no Lua for a rule. When the setting loads,
[szrule/rulescript](../staging/src/github.com/sentinez/core/szrule/rulescript/rulescript.go)
converts the script to the `expr` + `action` it declares
([scripts.Rule](../internal/memory/scripts/scripts.go)), and `szrule`
evaluates it like any `expr` rule (priority, terminal ranking, the
`secrule: matched` event, which lists its conditions). The script must be of
the declarative form of `scripting.ParseMatch`: branches of
`if <conditions> then return <action> end` over `ctx:path()`,
`ctx:header(k)`, `ctx:body()`, `ctx:ja4()`... Every condition of an `expr`
has its Lua form: `MATCHES` is `re.match(ctx:path(), "^/a")`, an IP CIDR
`ctx:ip_in("10.0.0.0/8")`, TLS `ctx:tls()`, a header or query presence
`ctx:has_header("X-A")` (`not` for NE, NOT_IN), an IN list an `or` of its
values. `not`, `and`, `or` and parentheses nest as deep as the expression
does, up to 256 conditions (`scripting.MaxConditions`): no group is
multiplied out into branches. A script outside of the form (a local, a
loop...) fails the load with an error wrapping `scripting.ErrNotDeclarative`,
and the previous chain stays.
`ToScript`/`FromScript` (`ToKeyScript`/`FromKeyScript` for a limiter)
convert both ways. The `---@stage` hooks still run as Lua.

### SecRule shapes

There are three SecRule shapes in the protos:

| Type | Used in |
|---|---|
| `apps.security.v1.SecRule` | API/DB model (control plane), with `metadata`, `action` as a string + `action_value` |
| `types.secrule.v1.SecRule` | Edge runtime and YAML input form |
| `dmz.edge.v1.SecRule` | Wrapper inside `Setting`: `{ingress: SecRule}` |

## 4.4 CDN rules

`types.cdn.v1.CDN{rule: Rule}` lives in
`setting.delivery.cdn`. As for SecRules, each namespace keeps a **chain**
of CDN rules
([internal/memory/cdnrules](../internal/memory/cdnrules/cdnrules.go)): only
`STATUS_ACTIVE` rules, sorted by **descending priority**, stable (equal
priorities keep declaration order). The `caching` node caches a request when at
least one rule matches, tried in that order. System paths (`/.sentinez/`)
are never cached. The cache is `patrickmn/go-cache` with a 1h TTL.

- Key: `method + scheme + host + path + query + Accept-Encoding` (the body
  is stored as encoded by the upstream, e.g. gzip or br).
- Only responses that can be replayed to other clients are stored: status
  exactly `200`, a non-empty captured body (bodies over 4 MiB are not
  captured), no `Set-Cookie`, and no `Cache-Control` containing `no-store`,
  `private` or `no-cache`.
- `Alt-Svc` is not stored, since the server adds it to every response.
- On a hit the status code is written before the body.

```yaml
delivery:
  cdn:
    - rule:
        name: "Static assets"
        status: STATUS_ACTIVE
        priority: 10
        expr: {orCondition: [{rules: [{condition: {source: FIELD_SOURCE_PATH, operator: OPERATOR_PREFIX, value: "/static/"}}]}]}
    - rule:
        name: "Other GET pages"
        status: STATUS_ACTIVE
        expr: {orCondition: [{rules: [{condition: {source: FIELD_SOURCE_METHOD, operator: OPERATOR_EQ, value: "GET"}}]}]}
```

## 4.5 Rate limiter

[core/szlimiter](../staging/src/github.com/sentinez/core/szlimiter/) implements
the **sliding window counter** algorithm:

```
count = curr.count + prev.count × (size − elapsed_in_curr) / size
allow  ⇔ count + n ≤ limit
```

- `SlidingWindow` holds two `LocalWindow`s (prev/curr); `advance` shifts the
  windows using `now.Truncate(size)`.
- `RateLimiter` keeps one `SlidingWindow` per key and a `timeout`
  mechanism: after a key exceeds the limit, every request from that key is
  rejected for `timeout` ms. Other keys are not affected.
- `Window.Sync` is a hook for syncing with a central store in the future
  (currently a no-op).

Configured in `security.limiters[]`: `timeWindow`, `maxRequests`, `timeout`
(`time.ParseDuration` strings). Every limiter needs an `ingress.expr` or an
`ingress.script` (not both): it only counts the requests it selects. An
`expr` limiter counts them under the client IP. A script defines a global
`key(ctx)` returning `true` for the requests to count, converted to its
`expr` when it loads like a secrule script; the request is counted under the
client IP too:

```yaml
limiters:
  - ingress:
      name: "Script: rate limit the login"
      status: STATUS_ACTIVE
      script: |
        ---@param ctx sz.Context
        function key(ctx)
          if ctx:path() == "/login" then
            return true
          end
        end
    timeWindow: 5s
    maxRequests: 20
    timeout: 5s
```

## 4.6 WAF — OWASP CRS on Coraza

### Rule generation

The original CRS (`.conf`) lives in [deploy/ruleroot](../deploy/ruleroot/).
The `szruleparser` tool
([tools/cmd/szruleparser](../staging/src/github.com/sentinez/tools/cmd/szruleparser/main.go)):

1. Parses SecLang with an ANTLR parser
   ([core/szmodsec/ruleparser](../staging/src/github.com/sentinez/core/szmodsec/ruleparser/)):
   each statement is a rule; a rule ending in `chain"` is joined with the
   next statement; `ver`, `paranoia-level` and action fields are extracted.
2. Loads the result into `ruleset.v1.Rulesets` and renders a Go
   template: each rule becomes a function returning a `*Rule` whose
   `Configuration` is base64-encoded, plus an `XxxOrder` slice that keeps the
   order.
3. Writes `core/szmodsec/gen[/v4-16-0|/v4-17-0]/*.sentinez_rules.gen.go`.

Regenerate everything with `bash hack/ruleparser.sh`.

### Assembling directives

`szcrs.GenerateRulesets(crs)`
([waf_loader.go](../staging/src/github.com/sentinez/core/szcrs/waf_loader.go))
maps the CRS setting (`ruleset.v1.Crs`) to SecLang directives: engine off →
`SecRuleEngine Off`; an unknown version → `ErrUnsupportedVersion`; the
setting is validated, then a `szruleset.Loader` concatenates, in order:

1. `setup` rules + the directives of the setting (`setupDirectives`) →
2. `REQUEST-901-INITIALIZATION` → 3. request groups of the version, each
loaded when its flag is on → 4. `audit` → 5. `default` →
6. `REQUEST-949-BLOCKING-EVALUATION` → 7. response groups →
8. global exclusions and rule overrides (`globalDirectives`, written last
because they change rules that must already be defined).

The order matters. The loader's `Rewrite` hook (`ruleOverrides`) applies the
rule states: `DETECTION_ONLY` moves the anomaly score of a rule to
`tx.detection_only_*` counters the blocking evaluation never reads, and a rule
forced `ENABLED` above the blocking paranoia level is scored at level 1 and
hoisted in front of its file.

`Flag` is a bitmask (`waf_flag.go`) with one bit per CRS file (905…944,
950…980). Only two flags have generated rules so far:

| Flag | v4.16.0 | v4.17.0 |
|---|---|---|
| `ReqAppAttackRCE` (932) | yes | yes |
| `ReqAppAttackSQLI` (942) | yes | reuses the v4.16.0 rules |

`szcrs.NewWAF(rootFs, crs)` generates the directives and hands them to
`szruleset.NewWAF`, which builds a `coraza.WAF` whose `RootFS` is the set of
`*.data` files embedded in [sentinez.go](../sentinez.go) (`WAF4160()`,
`WAF4170()`).

### Per-request processing

[core/szruleset](../staging/src/github.com/sentinez/core/szruleset/) — the
provider-agnostic Coraza runtime (`Loader`, `NewWAF`, `Rulesets`), modeled
on Coraza's HTTP middleware. `szcrs` keeps what is specific to the OWASP CRS
(flags, paranoia levels, setting, `AnomalyScore`, `IsFrameworkRule`):

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

`MemStore.LoadRulesets` creates one WAF per namespace from
`security.ruleset` (version, mode, paranoia levels, thresholds,
categories, exclusions, overrides — see
[03-edge.md](03-edge.md#35-proxyyaml-configuration)). `version` selects the
embedded CRS files: `CRS_VERSION_V4_16_0`, otherwise **v4.17.0**.
