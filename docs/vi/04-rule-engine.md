# 4. Rule engine & WAF

Sentinéz có hai cơ chế lọc request độc lập:

1. **SecRule / CDN rule** — expression do người dùng định nghĩa, đánh giá bởi
   `core/rules` (nhanh, không phụ thuộc thư viện ngoài).
2. **Rulesets (WAF)** — OWASP Core Rule Set chạy trên Coraza, nạp từ code Go
   sinh sẵn trong `core/modsec/gen`.

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

Ngữ nghĩa ([core/rules/rule_expr.go](../../staging/src/github.com/sentinez/core/rules/rule_expr.go)):

- `execOrCond(nil) == true` — expression rỗng **luôn khớp**.
- Một `AndCondition` khớp khi mọi `rules` khớp **và** `or_condition` con (nếu
  có) khớp.
- ID/tên các rule khớp được ghi vào `rulepb.MatchedRules` (nếu truyền vào).

API:

```go
eval := corerule.NewEval(expr)            // compile một lần
ok := eval(ctx corehttp.RequestContext, matched *rulepb.MatchedRules)
```

## 4.2 Nguồn dữ liệu (`FieldSource`) và toán tử

Bộ so khớp ở
[core/rules/rule_match.go](../../staging/src/github.com/sentinez/core/rules/rule_match.go).

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
- `GT/GTE/LT/LTE` có trong enum nhưng **chưa được cài đặt** (luôn không khớp).
- Source không xác định → không khớp (`bypass = false`).
- Trong YAML, `value` luôn là chuỗi (bản Lite) nên các toán tử list
  (`IN`/`NOT_IN`) chỉ dùng được khi tạo `Expression` trực tiếp (vd. qua API).

## 4.3 SecRule

Runtime type: `types.secrule.v1.SecRule{id, name, description, expr, action,
status, priority}`. Ở edge
([internal/memory/secrules/secrule.go](../../internal/memory/secrules/secrule.go)):

- Chỉ giữ rule `STATUS_ACTIVE`.
- Sắp xếp **priority giảm dần**, ổn định (cùng priority giữ thứ tự khai báo).
- Node `RUL` duyệt chuỗi; rule khớp đầu tiên có `ACTION_TYPE_BLOCK` → `403`.
  Các action khác (`LOG`, `MODIFY_HEADER`, `REDIRECT`, `SET_TAG`, `ROUTE_TO`)
  hiện chỉ được log debug, chưa thực thi.

Có ba dạng SecRule trong proto:

| Type | Dùng ở |
|---|---|
| `apps.security.v1.SecRule` | Model API/DB (control plane), có `metadata`, `action` dạng string + `action_value` |
| `types.secrule.v1.SecRule` / `SecRuleLite` | Runtime ở edge / dạng nhập YAML |
| `dmz.edge.v1.SecRule` | Wrapper trong `Setting`: `{ingress: Lite, ingress_runtime: Runtime}` |

## 4.4 CDN rule

`types.cdn.v1.CDN{rule: RuleLite, rule_runtime: Rule}` nằm trong
`setting.delivery.cdn`. Mỗi namespace giữ **một** CDN rule active (rule sau
ghi đè rule trước). Node `CDN` chỉ cache khi rule active và expression khớp.
Cache là `patrickmn/go-cache` TTL 1h.

- Key: `method + scheme + host + path + query + Accept-Encoding` (body được
  lưu nguyên dạng upstream đã nén, ví dụ gzip hoặc br).
- Chỉ lưu response có thể trả lại cho client khác: status đúng `200`, body
  đã được giữ lại và khác rỗng (body trên 4 MiB không được giữ), không có
  `Set-Cookie`, và `Cache-Control` không chứa `no-store`, `private` hay
  `no-cache`.
- Không lưu `Alt-Svc`, vì server tự thêm header này vào mọi response.
- Khi hit, status code được ghi trước body.

## 4.5 Rate limiter

[core/limiter](../../staging/src/github.com/sentinez/core/limiter/) cài đặt
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
`ruleparser-sentinez`
([tools/cmd/ruleparser-sentinez](../../staging/src/github.com/sentinez/tools/cmd/ruleparser-sentinez/main.go)):

1. Parse SecLang bằng parser ANTLR
   ([core/modsec/ruleparser](../../staging/src/github.com/sentinez/core/modsec/ruleparser/)):
   mỗi statement là một rule; rule kết thúc bằng `chain"` được nối với
   statement sau; trích `ver`, `paranoia-level`, các action field.
2. Đổ vào `coreruleset.v1.CoreRulesets` rồi render template Go: mỗi rule thành
   một hàm trả về `*CoreRule` với `Configuration` mã hoá base64, cùng slice
   `XxxOrder` giữ thứ tự.
3. Ghi ra `core/modsec/gen[/v4-16-0|/v4-17-0]/*.sentinez_rules.gen.go`.

Chạy lại toàn bộ: `bash hack/ruleparser.sh`.

### Ghép directive

`corers.GenerateRulesets(version, flag)`
([waf_loader.go](../../staging/src/github.com/sentinez/core/rulesets/waf_loader.go))
nối theo thứ tự:

1. `setup` → 2. `REQUEST-901-INITIALIZATION` → 3. rule theo version + flag →
4. `audit` → 5. `default` → 6. `REQUEST-949-BLOCKING-EVALUATION`.

`Flag` là bitmask (`waf_flag.go`) cho từng file CRS (905…944, 950…980). Hiện
chỉ hai flag có rule sinh sẵn:

| Flag | v4.16.0 | v4.17.0 |
|---|---|---|
| `ReqAppAttackRCE` (932) | có | có |
| `ReqAppAttackSQLI` (942) | có | dùng lại bản v4.16.0 |

Kết quả cũng được ghi ra file `WAF.conf.lock` ở thư mục làm việc (dùng để
xem directive cuối cùng).

`corers.NewWAF(version, fs, flag)` tạo `coraza.WAF` với `RootFS` là các file
`*.data` nhúng trong [sentinez.go](../../sentinez.go) (`WAF4160()`,
`WAF4170()`).

### Xử lý trong request

[core/rulesets](../../staging/src/github.com/sentinez/core/rulesets/) — phỏng
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

`MemStore.LoadRulesets` tạo WAF cho mỗi phần tử `security.rulesets[]` của
namespace. Hiện tại luôn dùng **CRS v4.16.0** và flag `ReqAppAttackRCE`, chưa
đọc tuỳ chọn từ cấu hình.
