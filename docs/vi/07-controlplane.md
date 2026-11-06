# 7. Control plane

Module `github.com/sentinez/controlplane`
([staging/src/github.com/sentinez/controlplane](../../staging/src/github.com/sentinez/controlplane/))
chứa các domain service. Mỗi domain theo cùng cấu trúc:

```
<domain>/v1/
├── <domain>.go     # gRPC module: NewService(ctx, conf) + Start() phục vụ qua bufconn
├── factory/        # NewDefaultHandler / NewDefaultService: dựng repo → service → handler
├── handler/        # cài <X>ServiceServer, kiểm tra quyền, log
├── service/        # nghiệp vụ
└── repos/<name>/   # truy cập PostgreSQL (interface IXxx + struct)
```

Scaffold mẫu cho từng lớp nằm trong `hack/.agent/skills/` (`create-service`,
`create-handler`, `create-repo`, `create-factory`, `create-module`).

## 7.1 Lớp lưu trữ (`shared/szstorage/dbx`)

[shared/szstorage/dbx](../../staging/src/github.com/sentinez/shared/szstorage/dbx/):

- `dbx.Database[T]`: `Insert`, `Select`, `Delete`, `Total`, `Table`, cùng
  `Exec`, `Query`, `CollectRows`, `CollectOneRow` nhận `query.Query`
  (interface `ToSql()` của squirrel).
- `postgres.New[T](ctx, conf, schema)`, với `schema` là một `dbx.Schema`
  (bảng, cột có kiểu, index) lấy từ
  [controlplane/pkg/schema](../../staging/src/github.com/sentinez/controlplane/pkg/schema/)
  (`schema.Tenants()`, ...; `schema.All()` liệt kê tất cả):
  - Tên bảng thật = `<mode>.sentinez.<name>` thay `.` bằng `_`, ví dụ
    `dev_sentinez_iam_users`; phải khớp regex
    `^(dev|sandbox|prod)_sentinez_[a-z]+(_[a-z]+)*$`.
  - Tự tạo bảng (`id TEXT PK, created_at, updated_at`), thêm cột còn thiếu
    theo thứ tự rồi tạo index, trong một round trip — **không xoá/đổi kiểu
    cột**.
  - Cùng schema đó dựng được bảng go-memdb (`dbx.MemDBSchema`): centraldata
    giữ mọi bảng controlplane trong bộ nhớ theo cách này.
  - Dùng một `pgxpool.Pool` toàn cục từ `SENZ_POSTGRES_URI`.
- Helper: `SelectBuilder` (+ phân trang `LIMIT/OFFSET ORDER BY created_at DESC`
  khi `page.index` và `page.size` > 0), `LimitOffset` (cùng `LIMIT/OFFSET`
  đó, thứ tự để caller tự đặt), `InsertBuilder` (`RETURNING id`),
  `UpdateBuilder` (tự set `updated_at`).
- `Schema.ColumnNames()` trả về tên các cột của schema theo thứ tự, để repo
  không phải chép lại.
- Transaction: `postgres.NewTX(conf).Begin(ctx)` → `*TxSession`;
  `postgres.WithTx(ss, db)` trả về `Database[T]` chạy trong transaction; repo
  cung cấp `WithTX(ss)`.
- ID bản ghi: `szrand.NewUUID(szprefix.<Entity>.Key())` =
  `<prefix>_<uuid>` (`user_<uuid>`), mỗi bảng một prefix trong
  [shared/szprefix](../../staging/src/github.com/sentinez/shared/szprefix/prefix.go)
  (a-z, 2 đến 5 ký tự, không đổi sau khi công khai); rule prefix `id` trong
  proto phải khớp.

Tên bảng ([shared/sztable](../../staging/src/github.com/sentinez/shared/sztable/tables.go)):
`iam.users`, `iam.accounts`, `tenant.resources`, `analytic.activities`,
`security.sec_rules`, cùng các bảng event của edge `analytic.secrules.events`,
`analytic.ratelimiters.events`, `analytic.corerulesets.events`,
`analytic.httpingress.events`, `analytic.httpegress.events` (xem
[Collector](#collector-thu-nạp-event-của-edge)).

Hằng tên cột (`User_Email`, `SecRule_Expr`, ...) được sinh bởi
`protoc-gen-go-senz` cho message có option `x_message.database_model = true`.

### TimescaleDB: hypertable và continuous aggregate

[shared/szstorage/dbx/timescale](../../staging/src/github.com/sentinez/shared/szstorage/dbx/timescale/)
xây trên `postgres.Open`: dùng chung pool và tham gia được transaction của
`postgres.WithTx`. Cần TimescaleDB (`SENZ_POSTGRES_URI` trỏ vào container
`sentinez.timescale` mà `hack/timescale-up.sh` bật).

- `timescale.New[T](ctx, conf, timescale.Hypertable{Schema, TimeColumn,
  ChunkInterval, Retention})` tạo extension `timescaledb`, bảng, các cột
  còn thiếu, hypertable, index và retention policy, trong một round trip.
  Bảng thường đã có được chuyển tại chỗ (`migrate_data`, bỏ primary key).
  - `TimeColumn` dùng để chia partition: cột `TIMESTAMPTZ` (mặc định
    `created_at`) hoặc `BIGINT` là unix mili giây; với loại sau, hàm
    `<table>_now()` mà các policy cần được tạo tự động. `ChunkInterval`
    mặc định 7 ngày; đổi giá trị chỉ áp dụng cho chunk tạo sau đó.
  - Unique index của hypertable bắt buộc chứa cột thời gian, nên `id` một
    mình không còn unique: khoá là `(id, <cột thời gian>)`, và insert bỏ
    qua bản ghi trùng bằng `Hypertable.OnConflictDoNothing()`
    (`ON CONFLICT (id, <cột thời gian>) DO NOTHING`).
  - `Retention` xoá các chunk cũ hơn mốc đó (không đặt khi bằng 0).
- `timescale.NewView[T](ctx, conf, timescale.ViewSchema{...})` tạo
  continuous aggregate (materialized view) trên một hypertable đã tồn tại:
  `time_bucket(Bucket, cột thời gian)` đặt tên `bucket`, các cột `GroupBy`,
  rồi các `Aggregates` (`Expr`, `As`); view được làm mới theo policy
  `Refresh` (`StartOffset`, `EndOffset`, `Schedule`) và xoá bucket sau
  `Retention`. `RealTime` cho phép khi query vẫn tổng hợp thêm các bucket
  chưa được làm mới.
  - View được tạo `WITH NO DATA`, và backfill một lần, tới
    `now - Refresh.EndOffset`, khi nó chưa tồn tại. Chế độ real time, policy
    refresh và retention được áp dụng ở mỗi lần khởi động; đổi định nghĩa
    (nguồn, bucket, `GroupBy`, `Aggregates`) thì phải dùng tên mới.
  - Không tạo hay refresh được trong transaction. `View.Refresh(ctx, from,
    to)` làm mới thủ công một khoảng; `View.CollectRows` để query.

## 7.2 Xác thực & phân quyền

[controlplane/pkg](../../staging/src/github.com/sentinez/controlplane/pkg/):

| Gói | Chức năng |
|---|---|
| `crypto` | `TokenGenerator(*typepb.Context)` ký JWT HS256 với secret = base64-decode(`SENZ_SECRET_KEY`); claim `auth` = `Context.String()` (prototext), `exp`. `BearerTokenVerifier` bỏ tiền tố `Bearer `, kiểm tra HMAC, parse lại `Context`. `HashPassword`/`CheckPasswordHash` dùng bcrypt |
| `headers` | `GetAuth(ctx)` đọc metadata `Authorization`, verify → `Auth{*typepb.Context}`; `Auth.Check(method)` gọi `perms.Allow`; `Allow(ctx, method)` làm cả hai |
| `passkey` | `NewWebAuthn(conf)` (RP = `core.Name`, RPID = `SENZ_HOSTNAME`, origin = `SENZ_CLIENT_ORIGIN`); `Store` in-memory (session WebAuthn + `Registration` đang chờ, xem 7.3) |

`typepb.Context` = `{name, expire_at, user_id, control_plane}`; JWT hết hạn
sau 1 giờ.

Phân quyền dựa trên **control plane** (`PORTAL` / `ADMIN`):

- Mỗi RPC có thể khai báo `option (sentinez.types.v1.x_method) = {ignore |
  control_planes: [...]}`; codegen sinh hàm `Get<Service><Method>()`.
- `perms.Allow(method, cp)`: cho phép nếu method nil, `ignore`, danh sách rỗng
  hoặc `cp` nằm trong danh sách.
- Handler phải **tự gọi** `headers.Allow(ctx, method)` — hiện
  `IAM.ListAccounts`, các handler Tenant và mọi RPC của Analytic làm việc
  này. Không có interceptor gRPC chung.

## 7.3 IAM

Proto: [apps/iam/v1](../../api/proto/sentinez/apps/iam/v1/). Code:
[apps/iam/v1](../../staging/src/github.com/sentinez/controlplane/apps/iam/v1/) — logic
account và user nằm ở `service/service.go`, các luồng WebAuthn ở
`service/passkey.go`.

### Mô hình dữ liệu: một user, mỗi phương thức đăng nhập một account

```
User (iam.users)                 Account (iam.accounts)
┌──────────────────────┐         ┌────────────────────────────────────┐
│ id                   │◄────────┤ user_id                            │
│ email  (danh tính)   │   1 : N │ provider = LOCAL   username,       │
│ full_name            │         │                    password_hash   │
│ control_plane        │         ├────────────────────────────────────┤
└──────────────────────┘         │ provider = PASSKEY credentials[],  │
                                 │                    provider_user_id│
                                 └────────────────────────────────────┘
```

- **User** là danh tính: `iam.users` (`id, email, full_name,
  control_plane`). Email xác định user — mỗi email một user.
- **Account** là một cách đăng nhập của user đó: `iam.accounts` (`id,
  user_id, email, username, password_hash, credentials TEXT[], provider
  INTEGER, provider_user_id`).
- `provider` là enum `iampb.Provider`:

| Provider | Giá trị | Trường sử dụng |
|---|---|---|
| `PROVIDER_LOCAL` | 1 | `username`, `password_hash` (bcrypt) |
| `PROVIDER_PASSKEY` | 2 | `credentials` (mỗi passkey một JSON `webauthn.Credential`), `provider_user_id` (WebAuthn user handle) |

Các quy tắc service đảm bảo:

- Mỗi user có tối đa một account cho mỗi provider. Nhiều passkey của một
  user là nhiều phần tử trong `credentials` của account passkey duy nhất.
- Mọi account mang email của user sở hữu nó.
- Đăng nhập tra account **trong phạm vi một provider**: mật khẩu không bao
  giờ khớp với account passkey và ngược lại.
- Access token được tạo từ `User`, bất kể đăng nhập bằng account nào:
  `newAccessToken` lấy control plane từ `User.control_plane` (`PORTAL` nếu
  chưa đặt). Ngoại lệ duy nhất là admin dựng sẵn: không có dòng nào trong
  DB, xem `Login`.

### Thêm phương thức đăng nhập

Email chưa được xác minh, nên không bao giờ gắn account vào user chỉ vì
trùng email:

- **Đăng ký** (không có token) chỉ được chấp nhận với email chưa thuộc user
  nào. Nó tạo `User` và `Account` đầu tiên trong một transaction
  (`createUserAccount`).
- **User đã tồn tại** chỉ thêm phương thức khi đang đăng nhập; user lấy từ
  token, không lấy từ request. Hiện mới có cho passkey
  (`PasskeyLinkChallenge`).

### RPC

| RPC | REST | Hành vi |
|---|---|---|
| `Login` | `PUT /iam/login` | Nếu username = `SENZ_ADMIN_USERNAME`: so mật khẩu với `SENZ_ADMIN_PASSWORD`, cấp token ADMIN (`user_id = sentinez.admin`). Ngược lại: tìm account `LOCAL` theo username/email, bcrypt, lấy user, cấp token |
| `CreateAccount` | `POST /iam/account` | Đăng ký bằng mật khẩu. Từ chối (`AlreadyExists`) username hoặc email đã có account `LOCAL`, username dành riêng `SENZ_ADMIN_USERNAME`, và email đã thuộc một user; sau đó tạo `User` (PORTAL) + account `LOCAL` |
| `CreateUser` | — (chỉ gRPC) | Tạo `User` không kèm account nếu email chưa có |
| `GetUser` | `GET /iam/user` | Theo `id` hoặc `email` (từ `iam.users`) |
| `ListUsers` | `GET /iam/users` | Lọc `ids`, `emails`, phân trang |
| `UpdateUser` | — | Cập nhật `full_name`, `email` |
| `DeleteUser` | `DELETE /iam/user/{id}` | Xoá user và toàn bộ account của nó trong một transaction |
| `ListAccounts` | `GET /iam/accounts` | Yêu cầu token; lọc `ids/emails/user_ids/usernames`; mỗi phần tử có `provider` |
| `Status` | `GET /iam/status` | Gọi Greeter.Status qua bufconn, trả về `Context` từ token (nếu có) |
| `PasskeyRegisterChallenge` | `GET /iam/passkey/register/challenge` | Đăng ký bằng passkey, bước 1. Cần email chưa thuộc user nào |
| `PasskeyLinkChallenge` | `GET /iam/passkey/link/challenge` | Thêm passkey cho user đang đăng nhập, bước 1. Yêu cầu token |
| `PasskeyRegisterVerify` | `POST /iam/passkey/register/verify` | Bước 2 của cả hai: xác minh phản hồi của authenticator rồi mới ghi |
| `PasskeyLoginChallenge` | `GET /iam/passkey/login/challenge` | `BeginLogin` cho account `PASSKEY` của username/email |
| `PasskeyLoginVerify` | `PUT /iam/passkey/login/verify` | `ValidateLogin`, cập nhật credential (sign count), cấp token |

### Các luồng passkey

Trạng thái giữa hai bước nằm trong `passkey.Store` (bộ nhớ, TTL 1 giờ) theo
`session_id` trả về: `SessionData` của WebAuthn và, với đăng ký, một
`passkey.Registration{UserId, Email, Handle}`.

**Đăng ký** (`PasskeyRegisterChallenge` → `PasskeyRegisterVerify`)

1. Challenge: đầu vào phải là email (ngược lại `InvalidArgument`) chưa
   thuộc user nào (`AlreadyExists`). Sinh một handle ngẫu nhiên; không ghi
   gì vào DB.
2. Verify: registration được lấy ra khỏi store — mỗi session chỉ hoàn tất
   được một lần. `CreateCredential` kiểm tra phản hồi của authenticator.
   Chỉ sau đó, và sau khi kiểm tra lại email vẫn còn trống, `User`
   (`full_name` = email) và account `PASSKEY` mới được tạo trong một
   transaction. Phản hồi sai hoặc giả mạo không để lại dòng nào.

**Thêm passkey** (`PasskeyLinkChallenge` → `PasskeyRegisterVerify`)

1. Challenge: user đọc từ token. Nếu đã có account passkey thì dùng lại
   handle của nó, chưa có thì sinh mới. `Registration.UserId` được đặt.
2. Verify: vì có `UserId`, lời gọi phải mang token của đúng user đó (ngược
   lại `Unauthenticated` / `PermissionDenied`). Credential được thêm vào
   account passkey của user, hoặc tạo account nếu đây là passkey đầu tiên.
   Nếu handle của account khác với handle đã cấp cho authenticator, lời gọi
   lỗi `InvalidArgument` và client làm lại từ đầu.

**Đăng nhập** (`PasskeyLoginChallenge` → `PasskeyLoginVerify`)

1. Challenge: tìm account `PASSKEY` theo username/email (`NotFound` nếu
   user chưa có) và lưu session.
2. Verify: tìm account theo handle lưu trong session
   (`GetByProviderUserId`), kiểm tra assertion, lưu credential đã cập nhật,
   cấp token cho user của account.

`AccountX` bọc `iampb.Account` để cài `webauthn.User`. WebAuthn user handle
là `provider_user_id`, một giá trị không mang nghĩa, nên không đổi khi
email đổi. Mọi passkey của một user dùng chung handle này.

### Các hàm tra cứu của repo

| Hàm | Dùng cho |
|---|---|
| `accounts.GetByProvider(provider, usernameOrEmail)` | Đăng nhập và kiểm tra trùng của một phương thức |
| `accounts.GetByProviderUserId(provider, id)` | Đăng nhập passkey: account của một WebAuthn handle |
| `accounts.GetByUserId(provider, userId)` | Account của một user theo provider |
| `accounts.GetByUsernameOrEmail(input)` | Mọi provider — chỉ để kiểm tra tồn tại, không dùng để xác thực |
| `accounts.DeleteByUserId(userId)` | Dùng bởi `DeleteUser` |
| `users.GetByEmail(email)` | Email thuộc user nào |

"Không tìm thấy" là `szerror.ErrNotFound`; helper `found` của service đổi nó
thành kết quả nil.

### Nâng cấp DB đang có

`provider` trước đây là cột `TEXT` không bao giờ được ghi. Tầng storage
thêm cột còn thiếu nhưng không bao giờ đổi kiểu cột (7.1), nên DB tạo trước
mô hình này vẫn giữ cột cũ và mọi truy vấn theo provider sẽ lỗi. Hãy tạo
lại bảng accounts (`<mode>_sentinez_iam_accounts`) hoặc tự đổi kiểu cột.
Passkey đăng ký trước thay đổi này được cấp email làm user handle, nên
authenticator của chúng chỉ hoạt động khi `provider_user_id` = email.

## 7.4 Tenant

Proto: [apps/tenant/v1](../../api/proto/sentinez/apps/tenant/v1/). Bảng
`tenant.resources` (`id, resource_setting BYTEA, resource_name,
resource_domain, plan, status`). `resource_setting` là `edgepb.Setting` lưu
dạng protobuf nhị phân — tức chính cấu hình edge của tenant.

| RPC | REST | Ghi chú |
|---|---|---|
| `ListResource` | `GET /tenant/resources` | Lọc `status`, `plan`, `resource_name`, `resource_domain` |
| `CreateResource` | `POST /tenant/resource` | `resource_setting` rỗng → `Setting{}` mặc định |
| `UpdateResource` | `PUT /tenant/resource` | Cập nhật domain/name/status/plan (không cập nhật setting) |
| `DeleteResource` | `DELETE /tenant/resource` | |
| `GetResource` | `GET /tenant/resource` | `default=true` → thay setting bằng mặc định |
| `GetResourceByDomain` | `GET /tenant/resource/{resource_domain}` | Lấy bản ghi đầu tiên khớp domain |
| `Status` | `GET /tenant/status` | |

Các RPC trên khai báo `control_planes: [PORTAL, ADMIN]` nhưng handler chưa
kiểm tra token.

## 7.5 Security

Proto: [apps/security/v1](../../api/proto/sentinez/apps/security/v1/). Bảng
`security.sec_rules` (`id, name, description, expr JSONB, status, priority,
action JSONB`).

| RPC | REST |
|---|---|
| `CreateSecRule` | `POST /security/secrule` |
| `GetSecRule` | `GET /security/secrule/{id}` |
| `UpdateSecRule` | `PUT /security/secrule/{id}` |
| `DeleteSecRule` | `DELETE /security/secrule/{id}` |
| `ListSecRules` | `GET /security/secrules` (lọc `ids`, phân trang) |
| `Status` | `GET /security/status` |

Service kiểm tra `sec_rule` bắt buộc khi tạo/sửa; `UpdateSecRule` đọc bản
ghi, ghi đè các trường từ request rồi gọi repo `Update`. Rule lưu ở đây
**chưa được đẩy xuống edge**.

## 7.6 Analytic, Greeter, CentralData

| Domain | Trạng thái |
|---|---|
| Analytic | `ListActivities` (`GET /analytic/activities`, cần token) trên bảng `analytic.activities` (`id, resource_id, unique_visitor INTEGER[]`), cùng các API đọc event của edge và histogram của chúng (xem [Analytic: đọc event](#analytic-đọc-event)). Chạy trong tiến trình apiserver |
| Greeter | Service mẫu: `SayHello` (`GET /greeter/say?name=`) trả lời kèm thời gian; `Status`. Có binary riêng `szgreeter` và cũng chạy in-process trong apiserver |
| CentralData | Chỉ có `Status`; binary `szcentraldata` chưa có `main` |
| Collector | `Status`, cùng phần thu nạp event của edge từ Kafka (bên dưới). Binary `szcollector` |

### Collector: thu nạp event của edge

Binary `szcollector` chạy hai app: gRPC server `CollectorService` và
`ingest` ([apps/collector/v1/ingest](../../staging/src/github.com/sentinez/controlplane/apps/collector/v1/ingest/)), app
này tiêu thụ event traffic của edge từ Kafka rồi insert theo batch vào các
bảng event analytic.

| Topic | Giá trị record | Bảng | Repository |
|---|---|---|---|
| `sentinez.event.http.secrule` | `secrulepb.SecRuleEvent` | `analytic.secrules.events` | `secruleeventsrepo.NewSecRules` |
| `sentinez.event.http.ratelimit` | `secrulepb.SecRuleEvent` | `analytic.ratelimiters.events` | `secruleeventsrepo.NewRateLimits` |
| `sentinez.event.http.rulesets` | `corerulesetpb.CoreRuleEvent` | `analytic.corerulesets.events` | `coreruleeventsrepo.New` |
| `sentinez.event.http.ingress` | `httppb.IngressEvent` (mỗi request một event, từ handler `logging`) | `analytic.httpingress.events` | `httpingressrepo.New` |
| `sentinez.event.http.egress` | `httppb.EgressEvent` (mỗi request được proxy tới upstream origin một event, từ handler `logging`) | `analytic.httpegress.events` | `httpegresseventsrepo.New` |

Luồng xử lý một batch:

1. `szkafka.Consumer` poll tối đa 1000 record. Consumer group là service
   key (`sentinez.apps.collector.v1`), nên các replica collector chia nhau
   partition.
2. `Ingest.Handle` gom record theo topic và chạy sink của từng topic song
   song (mỗi topic một bảng).
3. Sink decode record (`UnmarshalVT`), bỏ record không decode được kèm
   cảnh báo, và insert tối đa 500 dòng mỗi câu lệnh. Edge gán ULID cho mọi
   event; event nào thiếu id sẽ nhận `<prefix>_<topic>-<partition>-<offset>`,
   giống nhau ở mọi lần đọc lại. Event được decode vào message lấy từ pool
   (xem [Pool event](#pool-event)) và trả lại pool khi batch đã insert
   xong, nên `Create` không được giữ lại event.
4. Khi cả batch đã insert xong, các record được đánh dấu; offset đã đánh
   dấu được commit định kỳ, và trước khi partition chuyển đi hoặc consumer
   đóng.

Mỗi event được ghi ít nhất một lần (at least once). Insert lỗi thì thử lại
cả batch, chờ tăng dần từ 0.5 s đến 30 s, và trong lúc đó không poll batch
khác. Repo insert với `ON CONFLICT (id, <cột thời gian>) DO NOTHING` (xem
[TimescaleDB](#timescaledb-hypertable-và-continuous-aggregate)), nên batch
bị đọc lại không tạo bản ghi trùng.

Các bảng ([pkg/schema/analytic.go](../../staging/src/github.com/sentinez/controlplane/pkg/schema/analytic.go):
`AnalyticSecRulesEvent`, `AnalyticRateLimiterEvent`,
`AnalyticCoreRulesetsEvent`, `AnalyticHttpIngressEvent`,
`AnalyticHttpEgressEvent`) lấy theo các trường
của message Kafka: mảng là `TEXT[]`/`INTEGER[]`, `action` và `tags` là
`JSONB`, enum là `INTEGER`, `matched_time` (mili giây) là `BIGINT`, có index
`BTREE` trên `matched_time` và `resource_id`. Bảng ingress có các trường
scalar của `httppb.IngressEvent` (`status` là `INTEGER`, `asn` và `timestamp` (mili
giây) là `BIGINT`, còn lại là `TEXT`, header không gồm các header chứa
credential), có index trên `timestamp` và `resource_id`; `resource_id`/
`tenant_id` của nó do stage `DomainBased` của edge đặt
(`SetResourceId`/`SetTenantId`). Chúng không nằm trong `schema.All()`:
centraldata không giữ event trong bộ nhớ. Hằng tên cột lấy từ
`event_senz.pb.go`, sinh bởi `generate.sh` của `types/secrule/v1`,
`types/coreruleset/v1` và `types/net/http/v1`.

Bảng egress có một dòng cho mỗi request mà edge proxy tới upstream origin
(`httppb.EgressEvent`): `id` chính là id của event ingress của request đó,
nên hai bảng nối với nhau qua `id`; tiếp theo là `upstream` (vd
`http://10.0.0.1:8080`), `method`, `scheme` (của upstream), `path`,
`queries`, `status` (của origin, 502 khi không kết nối được), `content_type`,
`content_length` (-1 khi stream), `latency_micros`, `timestamp` (lúc gửi
request, mili giây), `resource_id`, `tenant_id` và `protocol` (phiên bản
HTTP client dùng để tới edge). Request do edge tự trả lời (cache, static, bị
chặn) không có dòng egress.

Mọi bảng event đều là hypertable
(`schema.EventHypertable(schema, timeColumn)`), chia theo cột thời gian
(`matched_time`, `timestamp`) thành chunk một ngày, và giữ trong
`schema.EventRetention`, 30 ngày.

Repository ([apps/collector/v1/repos](../../staging/src/github.com/sentinez/controlplane/apps/collector/v1/repos/)):

- `Create(ctx, events...)` insert các event (bắt buộc đã có id) trong một
  câu lệnh; `Get`, `Delete`, `WithTX`.
- `Histogram(ctx, *histogram.Filter)`: số lượng theo bucket thời gian, xem
  [Histogram](#histogram).
- `List(ctx, *eventfilter.Filter)`: lọc theo `ResourceIds` và theo
  `matched_time` (`From` tính cả mốc, `To` không tính; event ingress dùng
  `timestamp`), phân trang bằng `Page`, sắp theo `matched_time DESC`.
- Danh sách cột lấy từ `Schema.ColumnNames()` của schema bảng.

#### Histogram

[repos/histogram](../../staging/src/github.com/sentinez/controlplane/apps/collector/v1/repos/histogram/)
giữ, cho mỗi bảng event, ba continuous aggregate đếm số dòng theo bucket,
resource và key, để vẽ chart:

| View | Bucket | Làm mới | Khoảng dùng | Giữ |
|---|---|---|---|---|
| `<table>_secondly` | 1 giây | mỗi 10 giây, 3 phút gần nhất tới 1 giây trước | zoom tới 15 phút | 7 ngày |
| `<table>_minutely` | 1 phút | mỗi phút, 2 giờ gần nhất tới 1 phút trước | tới một ngày | 30 ngày |
| `<table>_hourly` | 1 giờ | mỗi 30 phút, 3 ngày gần nhất tới 1 giờ trước | tới 30 ngày | 30 ngày |

Ví dụ `dev_sentinez_analytic_httpingress_events_minutely`. View chạy real
time, nên các bucket mới nhất gồm cả những dòng chưa được làm mới; view
phút và giờ được giữ trong `schema.EventRetention`, view giây giữ 7 ngày
(`_secondRetention`). Các cột là `bucket` (unix mili giây),
`resource_id`, `tenant_id`, key và `total`.

Key chia nhỏ số đếm thêm: `status` http với event ingress và egress,
`behavior` với event core ruleset, không có với event secrule và rate limit
(`Point.Key` bằng 0).

`Histogram.Points(ctx, *Filter{Resolution, ResourceIds, TenantId, From, To})`
cộng `total` qua các tenant và trả về `[]*Point{Time, ResourceId, Key,
Total}`, cũ nhất trước. `Resolution` là `histogram.Second`, `Minute` hoặc
`Hour`, đánh số giống `HistogramResolution` của proto. Khi để 0, khoảng thời
gian sẽ quyết định: `Second` nếu `From`..`To` (`To` mặc định là hiện tại)
không quá 15 phút (`_secondSpan`) và `From` còn trong 7 ngày của view giây,
ngược lại là `Minute`. `From` tính cả mốc, `To` không tính, theo đầu bucket.

### Analytic: đọc event

Analytic đọc các bảng của collector qua
[apps/analytic/v1/repos/events](../../staging/src/github.com/sentinez/controlplane/apps/analytic/v1/repos/events/):
`eventsrepo.Events` giữ một `Reader[T]` chỉ đọc (`List`, `Histogram`) cho
mỗi bảng, dựng từ các repo của collector, nên analytic không ghi được
event. Khi dựng, nó sync bảng và view như collector.

| RPC | REST | Đọc |
|---|---|---|
| `ListHttpIngressEvents` | `GET /analytic/events/http-ingress` | `analytic.httpingress.events` |
| `ListHttpEgressEvents` | `GET /analytic/events/http-egress` | `analytic.httpegress.events` |
| `ListSecRuleEvents` | `GET /analytic/events/secrule` | `analytic.secrules.events` |
| `ListRateLimitEvents` | `GET /analytic/events/ratelimit` | `analytic.ratelimiters.events` |
| `ListCoreRuleEvents` | `GET /analytic/events/coreruleset` | `analytic.corerulesets.events` |
| `GetEventHistogram` | `GET /analytic/events/histogram` | các view của `kind` |

- Request list có `page`, `resource_ids`, `from` và `to`; event mới nhất
  trước. Một trang có 50 event khi không truyền, tối đa 500.
- `GetEventHistogram` nhận `kind` (`EventKind`), `resolution`
  (`HistogramResolution`, `UNSPECIFIED` thì chọn theo khoảng thời gian, xem
  [Histogram](#histogram)), `resource_ids`, `from` và `to`, trả về
  `HistogramPoint{time, resource_id, key, total}`, cũ nhất trước.
- Mọi RPC cần token (`headers.Allow`, `ignore: true`). Việc đọc chưa lọc
  theo tenant: không truyền `resource_ids` thì trả event của mọi tenant.

#### Pool event

Sink decode vào các message được tái sử dụng
([controlplane/pkg/pools](../../staging/src/github.com/sentinez/controlplane/pkg/pools/)):

| Package | Message | Sink |
|---|---|---|
| `pools/secrule` | `secrulepb.SecRuleEvent` | secrule, ratelimit |
| `pools/coreruleset` | `corerulesetpb.CoreRuleEvent` | rulesets |
| `pools/httpingress` | `httppb.IngressEvent` | ingress |
| `pools/httpegress` | `httppb.EgressEvent` | egress |

Mỗi package có `Pool()`, `Acquire()`, `Release(obj)`, dựa trên pool mà
vtproto sinh cho message (`<Msg>FromVTPool`, `ReturnToVTPool`, xem
[8.3](08-api-protobuf.md#83-sinh-code)). `ReturnToVTPool` gọi `ResetVT`, hàm
này giữ dung lượng của các trường repeated cho lần `UnmarshalVT` sau. Sink
gom event của một batch vào `pools.Batch` và `defer` `Close` của nó để trả
tất cả về pool.

Cấu hình: `SENZ_KAFKA_BROKERS` (không đặt: ingest bị tắt kèm cảnh báo,
gRPC server vẫn chạy) và database của `SENZ_POSTGRES_URI`. Các giới hạn
giữ bộ nhớ của consumer nằm ở
[09-core-shared.md](09-core-shared.md#szkafka).

## 7.7 Mã lỗi

[shared/szerror](../../staging/src/github.com/sentinez/shared/szerror/errors.go)
cung cấp gRPC status có định dạng `SENZ-<code>: <ERRORS_...>` và các hàm
`StatusXxxF(format, ...)` (`NotFound`, `Unauthorized`, `Forbidden`,
`InvalidData`/`InvalidArgument`, `AlreadyExists`, `Internal`,
`Unimplemented`). Sentinel `ErrNotFound`, `ErrInvalidData`, `ErrUnimplemented`
được gateway map sang HTTP. `IsNoRows(err)` nhận diện `pgx.ErrNoRows`.

## 7.8 Kiểm thử

- Mock repository sinh bằng mockery: `iam/v1/repos/{accounts,users}/mock`,
  `tenant/v1/repos/resources/mock`.
- `iam/v1/service/service_test.go`, `iam/v1/handler/iam_test.go`,
  `pkg/crypto/crypto_test.go`.
