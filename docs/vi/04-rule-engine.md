# 4. Rule engine & WAF

Sentinéz có hai cơ chế lọc request độc lập:

1. **SecRule / CDN rule** — expression do người dùng định nghĩa, đánh giá bởi
   `core/szrule` (nhanh, không phụ thuộc thư viện ngoài).
2. **Rulesets (WAF)** — OWASP Core Rule Set chạy trên Coraza, nạp từ code Go
   sinh sẵn trong `core/szmodsec/gen`.

## 4.1 Mô hình expression

Định nghĩa trong
[types/rule/v1/rule.proto](../../api/proto/sentinez/types/rule/v1/rule.proto):

```
Expression
 └─ or_condition: []AndCondition        // OR giữa các phần tử
      ├─ rules: []Rule                   // AND giữa các rule
      │    └─ condition: Condition{source, key, operator, value}
      └─ or_condition: []AndCondition    // lồng nhau: AND(rules) && OR(con)
```

Ngữ nghĩa ([core/szrule/rule_expr.go](../../staging/src/github.com/sentinez/core/szrule/rule_expr.go)):

- `execOrCond(nil) == true` — expression rỗng **luôn khớp**.
- Một `AndCondition` khớp khi mọi `rules` khớp **và** `or_condition` con (nếu
  có) khớp.
- ID/tên các rule khớp được ghi vào `rulepb.MatchedRules` (nếu truyền vào).

API:

```go
eval := szrule.NewEval(expr)            // compile một lần
ok := eval(ctx szhttp.RequestContext, matched *rulepb.MatchedRules)
```

## 4.2 Nguồn dữ liệu (`FieldSource`) và toán tử

Bộ so khớp ở
[core/szrule/rule_match.go](../../staging/src/github.com/sentinez/core/szrule/rule_match.go).

| Source | Lấy từ | Toán tử hỗ trợ |
|---|---|---|
| `PATH` | `ctx.Path()` | EQ, NE, CONTAINS, PREFIX, SUFFIX, MATCHES |
| `BODY` | `ctx.Body()` | như trên |
| `HOST` | `ctx.Host()` | như trên |
| `METHOD` | `ctx.Method()` | IN, NOT_IN (list) + nhóm chuỗi |
| `HEADER` | `key` rỗng/`"header"`: kiểm tra **tồn tại** header; có `key`: so giá trị header | Tồn tại: IN/NOT_IN (list) hoặc EQ/NE (một tên). Giá trị: list → membership (IN/EQ = có, NOT_IN/NE = không); string → nhóm chuỗi |
| `QUERY` | tương tự HEADER với query param | như HEADER |
| `IP` | `ctx.RequestIP()` | EQ, NE (IP hoặc CIDR), IN, NOT_IN (list IP/CIDR) |
| `TLS` | `"true"`/`"false"` | EQ, NE |
| `JA4` | fingerprint TLS | nhóm chuỗi; không có JA4 → không khớp |

Ghi chú:

- `MATCHES` dùng `regexp` của Go, regex được cache toàn cục theo pattern
  (`sync.Map`). Pattern lỗi → không khớp.
- `GT/GTE/LT/LTE` so sánh hai vế như số (`strconv.ParseFloat`) trên các
  nguồn dạng chuỗi (`PATH`, `BODY`, `HOST`, `METHOD`, `JA4`, giá trị một
  `HEADER` hoặc `QUERY`). Vế nào không phải số thì không khớp.
- Source không xác định → không khớp (`bypass = false`).
- Trong YAML, `value` luôn là chuỗi (bản Lite) nên các toán tử list
  (`IN`/`NOT_IN`) chỉ dùng được khi tạo `Expression` trực tiếp (vd. qua API).

## 4.3 SecRule

Runtime type: `types.secrule.v1.SecRule{id, name, description, expr, action,
status, priority}`. Ở edge
([internal/memory/secrules/secrule.go](../../internal/memory/secrules/secrule.go)):

- Chỉ giữ rule `STATUS_ACTIVE`.
- Sắp xếp **priority giảm dần**, ổn định (cùng priority giữ thứ tự khai báo).
- Node `RUL`
  ([internal/handlers/secure](../../internal/handlers/secure/action.go))
  duyệt chuỗi theo thứ tự trên và giữ mọi rule khớp:
  - Action **kết thúc** dừng request; action mạnh nhất trong các rule khớp
    thắng, `BLOCK` > `CHALLENGE` > `REDIRECT`, bằng nhau thì rule priority
    cao hơn thắng. `BLOCK` dừng việc duyệt; `CHALLENGE` và `REDIRECT` thì
    không, nên một `BLOCK` phía sau vẫn áp dụng cho client đã có pass.
  - Action **không kết thúc** của mọi rule khớp đều được áp dụng, trước
    action kết thúc. Khi nhiều rule đặt cùng header, tag hay route, rule
    priority cao hơn thắng.
  - Mỗi request log một event `secrule: matched`, mang action kết thúc
    (nếu không có thì action của rule priority cao nhất) và các tag.

Các action và `params` (`Action.params`, kiểu `Struct`; số và boolean được
đổi thành chuỗi):

| Action | `params` | Tác dụng |
|---|---|---|
| `BLOCK` | — | Trang `403` |
| `CHALLENGE` | — | Trang proof-of-work cho tới khi client có cookie `pow-pass` |
| `REDIRECT` | `location` (bắt buộc), `status` (301, 302, 303, 307, 308; mặc định 302) | Chuyển hướng client. Không có `location` dùng được thì rule bị bỏ qua |
| `LOG` | — | Chỉ ghi event |
| `MODIFY_HEADER` | header → giá trị | Đặt header của request gửi lên upstream. `proxySetHeaders` của location chạy sau và thắng nếu trùng header |
| `SET_TAG` | key → giá trị | Thêm vào `ContextExtra.tags` và event |
| `ROUTE_TO` | `location` | Proxy tới mục `server.locations` có `location` đó thay vì prefix dài nhất; location không tồn tại thì quay về khớp prefix |

```yaml
sec_rules:
  - ingress:
      name: "Chuyển link cũ sang shop mới"
      status: STATUS_ACTIVE
      expr: {orCondition: [{rules: [{condition: {source: FIELD_SOURCE_PATH, operator: OPERATOR_PREFIX, value: "/old-shop"}}]}]}
      action:
        type: ACTION_TYPE_REDIRECT
        params: {location: "/shop", status: 301}
  - ingress:
      name: "Gắn tag trang phân trang sâu"
      status: STATUS_ACTIVE
      expr: {orCondition: [{rules: [{condition: {source: FIELD_SOURCE_QUERY, key: "page", operator: OPERATOR_GT, value: "100"}}]}]}
      action:
        type: ACTION_TYPE_SET_TAG
        params: {deep_paging: true}
```

Control plane lưu giá trị action dạng `ActionValue{str_value, map_value}`
(console: một giá trị cho `REDIRECT`/`ROUTE_TO`, các dòng key/value cho
`MODIFY_HEADER`/`SET_TAG`); chưa có kênh đồng bộ xuống edge, nên edge chỉ đọc
`params` từ `proxy.yaml`.

Có ba dạng SecRule trong proto:

| Type | Dùng ở |
|---|---|
| `apps.security.v1.SecRule` | Model API/DB (control plane), có `metadata`, `action` dạng string + `action_value` |
| `types.secrule.v1.SecRule` / `SecRuleLite` | Runtime ở edge / dạng nhập YAML |
| `dmz.edge.v1.SecRule` | Wrapper trong `Setting`: `{ingress: Lite, ingress_runtime: Runtime}` |

## 4.4 CDN rule

`types.cdn.v1.CDN{rule: RuleLite, rule_runtime: Rule}` nằm trong
`setting.delivery.cdn`. Như SecRule, mỗi namespace giữ một **chuỗi** CDN
rule
([internal/memory/cdnrules](../../internal/memory/cdnrules/cdnrules.go)):
chỉ rule `STATUS_ACTIVE`, sắp xếp **priority giảm dần**, ổn định (cùng
priority giữ thứ tự khai báo). Node `CDN` cache request khi có ít nhất một
rule khớp, thử theo thứ tự đó. Đường dẫn hệ thống (`/.sentinez/`) không bao
giờ được cache. Cache là `patrickmn/go-cache` TTL 1h.

- Key: `method + scheme + host + path + query + Accept-Encoding` (body được
  lưu nguyên dạng upstream đã nén, ví dụ gzip hoặc br).
- Chỉ lưu response có thể trả lại cho client khác: status đúng `200`, body
  đã được giữ lại và khác rỗng (body trên 4 MiB không được giữ), không có
  `Set-Cookie`, và `Cache-Control` không chứa `no-store`, `private` hay
  `no-cache`.
- Không lưu `Alt-Svc`, vì server tự thêm header này vào mọi response.
- Khi hit, status code được ghi trước body.

```yaml
delivery:
  cdn:
    - rule:
        name: "Ảnh tĩnh"
        status: STATUS_ACTIVE
        priority: 10
        expr: {orCondition: [{rules: [{condition: {source: FIELD_SOURCE_PATH, operator: OPERATOR_PREFIX, value: "/static/"}}]}]}
    - rule:
        name: "Trang GET còn lại"
        status: STATUS_ACTIVE
        expr: {orCondition: [{rules: [{condition: {source: FIELD_SOURCE_METHOD, operator: OPERATOR_EQ, value: "GET"}}]}]}
```

## 4.5 Rate limiter

[core/szlimiter](../../staging/src/github.com/sentinez/core/szlimiter/) cài đặt
thuật toán **sliding window counter**:

```
count = curr.count + prev.count × (size − elapsed_in_curr) / size
allow  ⇔ count + n ≤ limit
```

- `SlidingWindow` giữ 2 `LocalWindow` (prev/curr); `advance` dịch cửa sổ theo
  `now.Truncate(size)`.
- `RateLimiter` giữ một `SlidingWindow` cho mỗi key (IP) và cơ chế
  `timeout`: sau khi một key vượt ngưỡng, mọi request của key đó bị từ chối
  trong `timeout` ms.
- `Window.Sync` là hook để đồng bộ với kho tập trung trong tương lai (hiện
  no-op).

Cấu hình ở `security.limiters[]`: `timeWindow`, `maxRequests`, `timeout` (chuỗi
`time.ParseDuration`).

## 4.6 WAF — OWASP CRS trên Coraza

### Sinh rule

CRS gốc (`.conf`) nằm ở [deploy/ruleroot](../../deploy/ruleroot/). Công cụ
`szruleparser`
([tools/cmd/szruleparser](../../staging/src/github.com/sentinez/tools/cmd/szruleparser/main.go)):

1. Parse SecLang bằng parser ANTLR
   ([core/szmodsec/ruleparser](../../staging/src/github.com/sentinez/core/szmodsec/ruleparser/)):
   mỗi statement là một rule; rule kết thúc bằng `chain"` được nối với
   statement sau; trích `ver`, `paranoia-level`, các action field.
2. Đổ vào `coreruleset.v1.CoreRulesets` rồi render template Go: mỗi rule thành
   một hàm trả về `*CoreRule` với `Configuration` mã hoá base64, cùng slice
   `XxxOrder` giữ thứ tự.
3. Ghi ra `core/szmodsec/gen[/v4-16-0|/v4-17-0]/*.sentinez_rules.gen.go`.

Chạy lại toàn bộ: `bash hack/ruleparser.sh`.

### Ghép directive

`szcrs.GenerateRulesets(version, flag)`
([waf_loader.go](../../staging/src/github.com/sentinez/core/szcrs/waf_loader.go))
nối theo thứ tự:

1. `setup` → 2. `REQUEST-901-INITIALIZATION` → 3. rule theo version + flag →
4. `audit` → 5. `default` → 6. `REQUEST-949-BLOCKING-EVALUATION`.

`Flag` là bitmask (`waf_flag.go`) cho từng file CRS (905…944, 950…980). Hiện
chỉ hai flag có rule sinh sẵn:

| Flag | v4.16.0 | v4.17.0 |
|---|---|---|
| `ReqAppAttackRCE` (932) | có | có |
| `ReqAppAttackSQLI` (942) | có | dùng lại bản v4.16.0 |

`szcrs.NewWAF(version, fs, flag)` tạo `coraza.WAF` với `RootFS` là các file
`*.data` nhúng trong [sentinez.go](../../sentinez.go) (`WAF4160()`,
`WAF4170()`).

### Xử lý trong request

[core/szcrs](../../staging/src/github.com/sentinez/core/szcrs/) — phỏng
theo middleware HTTP của Coraza:

| Bước | Hàm | Chi tiết |
|---|---|---|
| Tạo transaction | `NewRulesets(ctx, waf)` | Lấy từ `sync.Pool`; dùng `NewTransactionWithOptions` nếu WAF hỗ trợ context |
| Ingress | `ExecIngress` | `ProcessConnection` (IP:port), `ProcessURI`, thêm toàn bộ header + `Host` + `Transfer-Encoding`, `ProcessRequestHeaders`, đọc body nếu được phép, `ProcessRequestBody`. Interrupt → set status (mặc định 403 nếu action `deny`) |
| Egress | `ExecEgress` | Thêm response header, `ProcessResponseHeaders`, ghi body nếu được xử lý, `ProcessResponseBody`, copy body đã qua WAF |
| Kết thúc | `Final(cb)` | `ProcessLogging`, callback ghi event, `tx.Close()` |

Event WAF (`secrulepb.Event`): danh sách rule ID/severity/message (trừ rule
gây interrupt), `score` = data của rule interrupt (anomaly score), IP, domain,
transaction ID, request ID, thời gian.

### Bật WAF cho namespace

`MemStore.LoadRulesets` tạo một WAF cho mỗi namespace từ
`security.core_ruleset` (version, mode, paranoia level, threshold, categories,
exclusions, overrides — xem [03-edge.md](03-edge.md#35-cấu-hình-proxyyaml)).
`version` chọn bộ CRS nhúng trong binary: `VERSION_V4_16_0`, còn lại là
**v4.17.0**.
