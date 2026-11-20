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

## 7.1 Lớp lưu trữ (`pkg/szstorage/dbx`)

[pkg/szstorage/dbx](../../staging/src/github.com/sentinez/pkg/szstorage/dbx/):

- `dbx.Database[T]`: `Insert`, `Select`, `Delete`, `Total`, `Table`,
  `CreatedAt`, `UpdatedAt` (tên hai cột timestamp của bảng), cùng
  `Exec`, `Query`, `CollectRows`, `CollectOneRow` nhận `query.Query`
  (interface `ToSql()` của squirrel).
- `postgres.New[T](ctx, conf, schema)`, với `schema` là một `dbx.Schema`
  (bảng, cột có kiểu, index, tên cột timestamp) lấy từ
  [controlplane/pkg/schema](../../staging/src/github.com/sentinez/controlplane/pkg/schema/)
  (`schema.Tenants()`, ...; `schema.All()` liệt kê tất cả):
  - Tên bảng thật = `<mode>.sentinez.<name>` thay `.` bằng `_`, ví dụ
    `dev_sentinez_iam_users`; phải khớp regex
    `^(dev|sandbox|prod)_sentinez_[a-z]+(_[a-z]+)*$`.
  - Tự tạo bảng (`id TEXT PK`, rồi hai cột `TIMESTAMPTZ` đặt tên theo
    `Schema.CreatedAt`/`UpdatedAt`, mặc định `created_at` và `updated_at`),
    thêm cột còn thiếu theo thứ tự rồi tạo index, trong một round trip —
    **không xoá/đổi tên/đổi kiểu cột**.
  - Model có metadata được flatten đặt tên cột timestamp theo hằng được
    sinh: `metadata_created_at` (iam users và accounts, tenants, members,
    resources), `owner_created_at` (secrules, rate limits, rulesets,
    activities).
  - Cùng schema đó dựng được bảng go-memdb (`dbx.MemDBSchema`): centraldata
    giữ mọi bảng controlplane trong bộ nhớ theo cách này.
  - Dùng một `pgxpool.Pool` toàn cục từ `SZ_POSTGRES_URI`.
- Helper: `SelectBuilder` (+ phân trang `LIMIT/OFFSET ORDER BY <created at> DESC`
  khi `page.index` và `page.size` > 0), `LimitOffset` (cùng `LIMIT/OFFSET`
  đó, thứ tự để caller tự đặt), `InsertBuilder` (`RETURNING id`),
  `InsertRowsBuilder(db, columns, rows, valuesOf)` (nhiều dòng trong một
  `INSERT`, ví dụ `<Msg>Columns()` với `<Msg>Values`), `UpsertBuilder` và
  `UpdateBuilder` (tự set cột updated-at). Các builder insert thay byte NUL
  và UTF-8 không hợp lệ trong giá trị chuỗi bằng U+FFFD, vì text của
  postgres không chứa được (SQLSTATE 22021). `SelectCoalesce` đã bị bỏ: cột
  nullable được khai báo trên field proto (`x_field.nullable`).
- `Schema.ColumnNames()` trả về tên các cột của schema theo thứ tự, để repo
  không phải chép lại.
- Transaction: `postgres.NewTX(conf).Begin(ctx)` → `*TxSession`;
  `postgres.WithTx(ss, db)` trả về `Database[T]` chạy trong transaction; repo
  cung cấp `WithTX(ss)`.
- ID bản ghi: `szrand.NewUUID(szprefix.<Entity>.Key())` =
  `<prefix>_<uuid>` (`user_<uuid>`), mỗi bảng một prefix trong
  [pkg/szprefix](../../staging/src/github.com/sentinez/pkg/szprefix/prefix.go)
  (a-z, 2 đến 5 ký tự, không đổi sau khi công khai); rule prefix `id` trong
  proto phải khớp.

Tên bảng ([pkg/sztable](../../staging/src/github.com/sentinez/pkg/sztable/tables.go)):
`iam.users`, `iam.accounts`, `tenant.resources`, `analytic.activities`,
`security.secrules`, cùng các bảng event của edge `analytic.secrules.events`,
`analytic.secratelimits.events`, `analytic.secrulesets.events`,
`analytic.httpingress.events`, `analytic.httpegress.events` (xem
[Collector](#collector-thu-nạp-event-của-edge)).

Hằng tên cột (`User_Email`, `SecRule_Expr`, ...) được sinh bởi
`protoc-gen-go-sz` cho message có option `x_message.database_model = true`.
Model đặt thêm `x_message.scanner = true` (mọi model controlplane và event
của edge) có thêm row scanner được sinh: `<Msg>Columns()`, `<Msg>Values(m)`,
`ScanOne<Msg>(row)` và `Scan<Msg>(rows)` (xem
[8.2](08-api-protobuf.md#row-scanner-được-sinh)): repo select
`<Msg>Columns()...` rồi truyền `ScanOne<Msg>` cho `Select`/`CollectOneRow`
hoặc `Scan<Msg>` cho `CollectRows`, không còn `row.Scan` viết tay.
`schema_test.go` kiểm tra mọi cột mà scanner select đều có trong schema
bảng.

### TimescaleDB: hypertable và continuous aggregate

[pkg/szstorage/dbx/timescale](../../staging/src/github.com/sentinez/pkg/szstorage/dbx/timescale/)
xây trên `postgres.Open`: dùng chung pool và tham gia được transaction của
`postgres.WithTx`. Cần TimescaleDB (`SZ_POSTGRES_URI` trỏ vào container
`sentinez.timescale` mà `hack/timescale-up.sh` bật).

- `timescale.New[T](ctx, conf, timescale.Hypertable{Schema, TimeColumn,
  ChunkInterval, Retention})` tạo extension `timescaledb`, bảng, các cột
  còn thiếu, hypertable, index và retention policy, trong một round trip.
  Bảng thường đã có được chuyển tại chỗ (`migrate_data`, bỏ primary key).
  - `TimeColumn` dùng để chia partition: cột `TIMESTAMPTZ` (mặc định là
    cột created-at của schema) hoặc `BIGINT` là unix mili giây; với loại sau, hàm
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
| `crypto` | `TokenGenerator(*typepb.Context)` ký JWT HS256 với secret = base64-decode(`SZ_SECRET_KEY`); claim `auth` = `Context.String()` (prototext), `exp`, `iat`, `iss` = `sentinez.iam`, `aud` = `sentinez.controlplane`. `BearerTokenVerifier` bỏ tiền tố `Bearer `, kiểm tra HMAC (chỉ HS256), hạn dùng (bắt buộc), issuer và audience, rồi parse lại `Context`. `ServiceTokenGenerator(name, ttl)` / `ServiceTokenVerifier` làm điều tương tự cho token service (control plane `SERVICE`), với `SZ_SERVICE_KEY` và `iss` = `aud` = `sentinez.service`, nên không loại nào dùng thay được loại kia. `HashPassword`/`CheckPasswordHash` dùng bcrypt |
| `headers` | `GetAuth(ctx)` trả người gọi mà interceptor `authz` đã xác thực, nếu không thì đọc và verify metadata `Authorization` (token user, nếu không thì token service) → `Auth{*typepb.Context}`; `Auth.Check(method)` gọi `perms.Allow` (lỗi `PermissionDenied`); `Auth.IsAdmin()`; `Auth.IsService()`; `WithAuth(ctx, auth)` |
| `authz` | `Unary()` / `Stream()`, cài cùng nhau bằng `ServerOptions()`: các interceptor từ chối mặc định (xem dưới); `Authorizer`: người gọi được làm gì trong các tenant; `Scoped`: giới hạn request list theo resource của người gọi |
| `loginlimit` | `Limiter`: chặn một key sau N lần đăng nhập thất bại trong một cửa sổ; IAM dùng |
| `passkey` | `NewWebAuthn(conf)` (RP = `core.Name`, RPID = `SZ_HOSTNAME`, origins = `ClientOrigins(conf)`, tức `SZ_CLIENT_ORIGIN` cách nhau bằng dấu phẩy); `Store` in-memory (session WebAuthn + `Registration` đang chờ, xem 7.3) |

`typepb.Context` = `{name, expire_at, user_id, control_plane}`; JWT hết hạn
sau 1 giờ.

Phân quyền có hai lớp.

**1. Theo từng call: control plane (`PORTAL` / `ADMIN` / `SERVICE`).**
`authz.Unary()` và `authz.Stream()` (stream được kiểm tra một lần, lúc mở)
được cài trên mọi gRPC server của control plane bằng
`szgrpc.New(authz.ServerOptions()...)`, `authz.Unary()` trên local conn của
gateway, và từ chối mặc định:

- Mỗi RPC có thể khai báo `option (sentinez.types.v1.x_method) = {public |
  ignore | control_planes: [...]}`; codegen sinh hàm `Get<Service><Method>()`.
- `public` không cần token (token hợp lệ vẫn được đọc). RPC khác, **kể cả
  RPC không có `x_method`**, cần token hợp lệ.
- Sau đó `perms.Allow(method, cp)`: cho phép nếu method nil, `public` hoặc
  `ignore`, danh sách rỗng hoặc `cp` nằm trong danh sách. `ignore` chỉ bỏ
  qua bước này, không bỏ qua token.
- Các RPC public: `Login`, `CreateAccount`, các RPC passkey register/login
  và các RPC `Status`/`SayHello`. `CreateUser` và `DeleteUser` chỉ cho
  control plane `ADMIN`.
- Token service (`SERVICE`) chỉ mở được các method liệt kê
  `CONTROL_PLANE_SERVICE` (hiện là `Sync` của centraldata), không bao giờ mở
  method có danh sách rỗng. Trên kết nối có client certificate đã verify,
  tên trong token phải trùng CN của certificate.
- `szgrpc.New` và `grpcgateway.NewLocalConn` đặt `recovery.Unary()` đầu
  chuỗi (và `szgrpc.New` thêm `recovery.Stream()`): panic của handler làm call
  lỗi `Internal` thay vì dừng process.

**2. Theo từng tenant: vai trò của người gọi.** Token không chứa tenant hay
vai trò. `authz.Authorizer` đọc membership (`tenant.members`) của người gọi
ở mỗi call, nên đổi vai trò có hiệu lực ngay.

| Cần | Vai trò (hoặc cao hơn) | Call |
|---|---|---|
| đọc | `MEMBER` | xem tenant, member, resource, rule và event |
| ghi | `LEADER` | tạo/sửa/xoá resource, SecRule, SecRateLimit, SecRuleset |
| quản trị | `ADMIN` | sửa/xoá tenant, tạo/sửa/xoá member |
| thanh toán | `SUPER_ADMIN` | đổi `plan` của tenant (thứ tenant trả tiền) |

Tenant mới luôn bắt đầu ở plan `FREE` (người gọi `PORTAL` xin plan khác nhận
`PermissionDenied`); chỉ control plane `ADMIN` mới tạo được tenant ở plan khác.

- Thứ bậc `SUPER_ADMIN` > `ADMIN` > `LEADER` > `MEMBER`. Control plane
  `ADMIN` làm được mọi thứ.
- Người gọi ngoài tenant nhận `NotFound` nên không biết gì về tenant; người
  trong tenant nhưng thiếu vai trò nhận `PermissionDenied`.
- Chỉ gán được vai trò tối đa bằng vai trò của chính người gọi, không bao
  giờ `SUPER_ADMIN` (chỉ control plane `ADMIN` gán được), và tenant không
  được mất admin cuối cùng.
- Các list bị giới hạn theo tenant/resource của người gọi (`TenantIds`,
  `ResourceIds`, `Scoped`). Người gọi không thấy gì nhận list rỗng mà không
  query: filter rỗng sẽ khớp mọi thứ.
- Rule security lấy tenant từ resource của nó, không từ request.

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
| `Login` | `PUT /iam/login` | Nếu username = `SZ_ADMIN_USERNAME`: so mật khẩu với `SZ_ADMIN_PASSWORD` trong thời gian hằng số (từ chối khi chưa đặt), cấp token ADMIN (`user_id = sentinez.admin`). Ngược lại: tìm account `LOCAL` theo username/email, bcrypt, lấy user, cấp token. Số lần thất bại bị giới hạn: 10 mỗi account và 50 tổng mỗi client trong 15 phút, sau đó `ResourceExhausted` (HTTP 429) tới hết cửa sổ. Client là hop `X-Forwarded-For` cuối cùng |
| `CreateAccount` | `POST /iam/account` | Đăng ký bằng mật khẩu. Từ chối (`AlreadyExists`) username hoặc email đã có account `LOCAL`, username dành riêng `SZ_ADMIN_USERNAME`, và email đã thuộc một user; sau đó tạo `User` (PORTAL) + account `LOCAL` |
| `CreateUser` | — (chỉ gRPC) | Chỉ control plane `ADMIN`. Tạo `User` không kèm account nếu email chưa có |
| `GetUser` | `GET /iam/user` | Theo `id` (từ `iam.users`); người gọi `PORTAL` chỉ lấy được chính mình hoặc user chung tenant, ngoài ra `NotFound` |
| `ListUsers` | `GET /iam/users` | Lọc `ids`, `emails`, phân trang. Người gọi `PORTAL`: với `ids` chỉ giữ chính mình và user chung tenant; với `emails` là khớp chính xác (tối đa 5, để mời); không truyền gì thì chỉ chính mình |
| `UpdateUser` | — | Cập nhật `full_name`, `email` |
| `DeleteUser` | `DELETE /iam/user/{id}` | Chỉ control plane `ADMIN`. Xoá user, các membership tenant và toàn bộ account của nó trong một transaction |
| `ListAccounts` | `GET /iam/accounts` | Yêu cầu token; lọc `ids/emails/user_ids/usernames`; mỗi phần tử có `provider`. Người gọi `PORTAL` chỉ lấy account của chính mình |
| `Status` | `GET /iam/status` | Yêu cầu token. Gọi Greeter.Status qua bufconn, trả về `Context` của token |
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

Proto: [apps/tenant/v1](../../api/proto/sentinez/apps/tenant/v1/). Một
`Resource` gồm `id, tenant_id, name, domain, locations, status, tls_mode`, lưu
trong bảng `tenant.resources` mỗi field một cột (`locations` dạng JSONB,
các `edgepb.Location` của domain). `CreateResource` mang một
`edgepb.Resource` (`resource`), được `sznormalize.TenantResource` tách
thành các field này; `UpdateResource` mang một `Resource` và một
`update_mask` (xem [Cập nhật](#cập-nhật-và-update_mask));
edge nhận lại setting qua `sznormalize.EdgeSetting` / `EdgeResource`, với
`id` luôn là `id` của resource. Service
đổi domain về chữ thường như edge so khớp, và từ chối resource không có
domain.

| RPC | REST | Ghi chú |
|---|---|---|
| `ListResource` | `GET /tenant/resources` | Lọc `status`, `resource_name` (`name`), `resource_domain` (`domain`), `tenant_ids` |
| `CreateResource` | `POST /tenant/resource` | Bắt buộc `resource` có `domain` |
| `UpdateResource` | `PUT /tenant/resource` | Ghi các trường của `resource` mà `update_mask` nêu, trong `name`, `domain`, `locations`, `status`, `tls_mode` |
| `DeleteResource` | `DELETE /tenant/resource` | |
| `GetResource` | `GET /tenant/resource` | `default=true` → resource bỏ các location |
| `GetResourceByDomain` | `GET /tenant/resource/{resource_domain}` | Lấy bản ghi đầu tiên khớp domain |
| `Status` | `GET /tenant/status` | |

Các RPC trên khai báo `control_planes: [PORTAL, ADMIN]`, và service kiểm tra
vai trò của người gọi trong tenant (xem 7.2).

### CA certificate

Proto: [apps/tenant/v1](../../api/proto/sentinez/apps/tenant/v1/). Một
`CaCertificate` là một chứng chỉ của resource, cùng các trường được parse từ
PEM (`subject`, `issuer`, `provider`, `serial_number`, `fingerprint_sha256`,
`not_before`, `not_after`, `self_signed`) và `usage`: `UPSTREAM` (kiểm tra các
máy chủ origin), `CLIENT` (kiểm tra chứng chỉ client, mTLS), `SERVER`
(chứng chỉ TLS mà edge trình ra) hoặc `UPSTREAM_CLIENT` (chứng chỉ client mà
edge trình cho máy chủ origin, mTLS). Hai usage sau là leaf: chứng chỉ không
cần là CA, và luôn có khoá riêng. Bản
ghi bất biến: chỉ tạo, đọc và xoá, không cập nhật. PEM và khoá riêng được mã
hoá bằng `SZ_DATA_KEY` trong database, API không bao giờ trả khoá về
(`has_private_key` cho biết có khoá hay không). Fingerprint là duy nhất trong
một resource.

| RPC | REST | Ghi chú |
|---|---|---|
| `CreateCaCertificate` | `POST /tenant/cacertificate` | `pem` chứa một hoặc nhiều chứng chỉ, mỗi chứng chỉ lưu thành một bản ghi |
| `GetCaCertificate` | `GET /tenant/cacertificate` | Theo `id` (`cacrt_...`) |
| `ListCaCertificate` | `GET /tenant/cacertificates` | Phân trang, giới hạn theo các resource của người gọi |
| `DeleteCaCertificate` | `DELETE /tenant/cacertificate` | |

`CreateCaCertificate` nhận:

- Một bundle gồm số chứng chỉ bất kỳ. Mỗi chứng chỉ phải chưa hết hạn, phải là
  CA (trừ khi usage là `SERVER` hoặc `UPSTREAM_CLIENT`), và không được lặp. Chữ nằm giữa các khối PEM
  bị từ chối.
- Một khoá riêng, trong `private_key` hoặc là một khối của `pem` (tối đa một
  khoá, tính cả hai nơi). Khoá được gắn vào chứng chỉ khớp với nó trong bundle,
  nên có thể tải cả chain (leaf + intermediate) kèm khoá của leaf; khoá không
  khớp chứng chỉ nào, hoặc được mã hoá bằng passphrase, bị từ chối. Usage
  `SERVER` và `UPSTREAM_CLIENT` bắt buộc có khoá.

## 7.5 Security

Proto: [apps/security/v1](../../api/proto/sentinez/apps/security/v1/). Bảng
`security.secrules` (`id, name, description, expr JSONB, status, priority,
action JSONB`).

| RPC | REST |
|---|---|
| `CreateSecRule` | `POST /security/secrule` |
| `GetSecRule` | `GET /security/secrule/{id}` |
| `UpdateSecRule` | `PUT /security/secrule/{id}` |
| `DeleteSecRule` | `DELETE /security/secrule/{id}` |
| `ListSecRules` | `GET /security/secrules` (lọc `ids`, phân trang) |
| `Status` | `GET /security/status` |

Service kiểm tra `secrule` bắt buộc khi tạo/sửa; `UpdateSecRule` đọc bản
ghi rồi ghi các trường mà `update_mask` nêu (xem
[Cập nhật](#cập-nhật-và-update_mask)). Rule lưu ở đây tới
edge qua centraldata (xem [CentralData](#centraldata-đồng-bộ-các-bảng-xuống-edge)).

### Cập nhật và update_mask

`UpdateResource`, `UpdateTenant`, `UpdateSecRule`, `UpdateSecRateLimit` và
`UpdateSecRuleset` nhận model (`resource`, `tenant`, `secrule`, ...) và
một `google.protobuf.FieldMask` `update_mask`, đọc theo AIP-134:

- không có path: các trường được set trong model (mask ngầm định)
- `*`: mọi trường được phép cập nhật
- còn lại: các path của nó; trường được nêu mà model để trống thì bị xoá,
  path không phải trường được phép cập nhật trả về `InvalidArgument`

Trong body JSON, mask là chuỗi tên lowerCamelCase:
`"updateMask": "name,timeWindow"`. Service đọc bản ghi đang lưu, chép các
trường trong mask lên nó bằng `szprotobuf.ApplyMask`, kiểm tra kết quả
(domain của resource và plan của tenant là bắt buộc, rate limit và core
ruleset được validate toàn bộ), rồi repo `Update(ctx, model, paths)` chỉ
ghi các cột của các path đó (`postgres.M.Only`). Mỗi service tự liệt kê
các trường được phép cập nhật; id, owner và tenant không bao giờ đổi.
`UpdateMember` và `UpdateUser` chỉ đặt một trường (`roles`, `full_name`)
nên không nhận mask.

## 7.6 Analytic, Greeter, CentralData

| Domain | Trạng thái |
|---|---|
| Analytic | `ListActivities` (`GET /analytic/activities`, cần token) trên bảng `analytic.activities` (`id, resource_id, unique_visitor INTEGER[]`), cùng các API đọc event của edge và histogram của chúng (xem [Analytic: đọc event](#analytic-đọc-event)). Chạy trong tiến trình apiserver |
| Greeter | Service mẫu: `SayHello` (`GET /greeter/say?name=`) trả lời kèm thời gian; `Status`. Có binary riêng `szgreeter` và cũng chạy in-process trong apiserver |
| CentralData | `Status`, và `Sync`: stream nạp các bảng cho edge, xem dưới. Binary `szcentraldata` |
| Collector | `Status`, cùng phần thu nạp event của edge từ Kafka (bên dưới). Binary `szcollector` |

### CentralData: đồng bộ các bảng xuống edge

`szcentraldata` giữ các bảng trong `SZ_SYNC_TABLES` (mặc định:
`security.secrules`, `security.secratelimits`, `security.secrulesets`,
`tenant.resources`, `tenant.cacertificates`) trong go-memdb, và edge leader
của mỗi zone đi theo nó thay cho các edge khác:

```
postgres ──trigger: pg_notify(id)──► centraldata ──Olric pub/sub──► edge leader
                                         ▲                              │
                                         └──── Sync (gRPC stream) ◄─────┤ khi được bầu
                                                                        ▼
                                        mọi edge ◄── SyncMap sentinez.dmz.edge.central
```

- **Phát hiện thay đổi.** Khi khởi động, `postgres.NewNotifying` tạo từng
  bảng và một trigger `AFTER INSERT OR UPDATE OR DELETE` gọi `pg_notify` với
  id của row. `Syncer`
  ([service/syncer.go](../../staging/src/github.com/sentinez/controlplane/apps/centraldata/v1/service/syncer.go))
  `LISTEN` trước rồi mới nạp các bảng, nên không thay đổi nào lọt vào giữa.
  Với mỗi notification nó đọc lại row và áp dụng (hoặc xoá khi row không
  còn). Mỗi phút nó so các bảng với bộ nhớ để bắt notification postgres làm
  rơi. Mất kết nối thì nạp lại các bảng và bắt đầu epoch mới. Mỗi lần nạp
  log tiến độ từng bảng (`centraldata: load <table> i/n: rows, upserted,
  deleted in ...`); lần đối soát chỉ log các bảng có thay đổi.
- **Version.** Mỗi lần ghi vào bộ nhớ tăng version; một `Change` mang
  `(table, op, id, row, version, epoch)`. `epoch` đổi khi lịch sử không còn
  liền mạch.
- **Leader được bầu: `Sync`**, RPC stream duy nhất. Nó gửi mọi row dưới dạng
  `Change`, rồi một `Snapshot{version, epoch}`. Các row là của version đó:
  edge áp các thay đổi được publish sau nó.
- **Live: Olric.** Centraldata publish mỗi `Change` trong một `Envelope` ký
  HMAC-SHA256 bằng `SZ_SYNC_KEY`, trên channel
  `sentinez.centraldata.changes` của cluster edge (`SZ_OLRIC_ADDRESSES`,
  qua Olric cluster client: nó không là member). Pub/sub là at-most-once, nên
  edge ([internal/centralsync](../../internal/centralsync/follower.go))
  subscribe trước khi gọi `Sync`, bỏ những gì không ký bằng key hoặc không
  nằm sau snapshot, và nạp lại khi version bị hụt, epoch đổi, hoặc mỗi 5
  phút.
- **Chỉ leader theo centraldata.** Follower chạy trong một `leader.Task`
  của leader zone ([03-edge.md](03-edge.md), `Server.Lead`) và dừng khi hết
  term, nên centraldata chỉ phục vụ một `Sync` mỗi zone thay vì mỗi edge.
  `memory.CentralSink` của nó giữ các row và dựng lại `edgepb.Setting` của
  mỗi resource có row thay đổi (server của resource, rule, limiter và core
  ruleset), trên nền setting đã nạp cho domain đó, nên CDN của file được
  giữ. Nó ghi setting vào `SyncMap` `sentinez.dmz.edge.central` theo
  resource id (setting không đổi thì không ghi lại), và mọi edge, kể cả
  leader, nạp những gì map giữ (`MemStore.applyCentral`); setting từ các
  bảng chiếm domain của setting từ file. Map nằm trong DMap của Olric, nên
  edge khởi động lúc chưa có leader vẫn nạp được setting ghi gần nhất.
  Leader mới bắt đầu với sink rỗng: lần nạp đầu xoá khỏi map những gì
  snapshot không có, tức các resource bị xoá trong lúc zone không có leader.
  Leader đã bị hạ vẫn có thể ghi cho tới khi nhận ra mất lock (TTL session
  Consul); nó ghi đúng những gì centraldata giữ, và lần nạp của leader kế
  tiếp sẽ sửa lại. Không có `SZ_SERVICE_TOKEN` thì leader không theo
  centraldata; nếu không tới được centraldata, các edge chạy bằng file và
  cluster, và leader thử lại nền. Lần resync chỉ dựng lại các resource có
  row thay đổi, và những resource map còn giữ mà snapshot không còn; các lần
  ghi của một lượt dựng lại dùng chung timeout 30s, quá hạn thì chỉ áp trên
  leader và lần đối soát của map sẽ chia sẻ sau.
- **Độ trễ.** Thay đổi tới leader qua Olric pub/sub, rồi tới các edge khác
  qua `SyncMap` (một lần `Put` vào DMap, rồi publish): vài mili-giây, cộng
  thời gian dựng lại store trên mỗi edge. Tin mất giữa leader và một edge
  được lần đối soát của map sửa, trong vòng 30s; tin mất giữa centraldata
  và leader được sửa ở thay đổi kế tiếp (hụt version) hoặc lần resync,
  trong vòng 5 phút.
- **Log debug** (`edge: centraldata: ...`): leader log những gì nó nạp
  (`loaded N rows`, `received <op> <table> <id>, version`) và những gì nó
  ghi (`rebuild resource` / `remove resource`); mọi edge log từng thay đổi
  của map, `written here as the leader` hoặc `synced from the leader <node>`
  (`SyncMap.Origin`).
- **Tin cậy.** `Sync` chỉ mở cho `CONTROL_PLANE_SERVICE`. Edge gọi qua mTLS
  (certificate của CA nội bộ, `hack/mkcert-internal.sh`) với token service
  (`go run ./cmd/szservicetoken -name <CN của certificate>`) ký bằng
  `SZ_SERVICE_KEY`, chỉ centraldata giữ: edge giữ token, không giữ key.
  Tên trong token phải là CN của certificate của kết nối. Token service không
  mở method nào không liệt kê control plane service, và token user không bao
  giờ mở được `Sync`.
- **CA certificate** đi đúng như bảng lưu: pem và private key vẫn được seal
  bằng `SZ_DATA_KEY`, thứ centraldata không bao giờ đọc. Edge đặt chúng vào
  `ingress.tls` của setting theo usage (`SERVER` → `certificates`,
  `UPSTREAM` → `upstream_cas`, `UPSTREAM_CLIENT` → `upstream_certificates`)
  và chỉ mở khi nạp setting (xem [03-edge.md](03-edge.md#upstream-tls-ca-riêng-mtls)). Ai
  subscribe được channel Olric của edge cũng chỉ thấy bản đã seal.

### Collector: thu nạp event của edge

Binary `szcollector` chạy hai app: gRPC server `CollectorService` và
`ingest` ([apps/collector/v1/ingest](../../staging/src/github.com/sentinez/controlplane/apps/collector/v1/ingest/)), app
này tiêu thụ event traffic của edge từ Kafka rồi insert theo batch vào các
bảng event analytic.

| Topic | Giá trị record | Bảng | Repository |
|---|---|---|---|
| `sentinez.event.http.secrule` | `secrulepb.SecRuleEvent` | `analytic.secrules.events` | `secruleeventsrepo.NewSecRules` |
| `sentinez.event.http.secratelimit` | `secrulepb.SecRuleEvent` | `analytic.secratelimits.events` | `secruleeventsrepo.NewSecRateLimits` |
| `sentinez.event.http.secruleset` | `rulesetpb.SecRulesetEvent` | `analytic.secrulesets.events` | `secruleseteventsrepo.New` |
| `sentinez.event.http.ingress` | `httppb.IngressEvent` (mỗi request một event, từ handler `logging`) | `analytic.httpingress.events` | `httpingresseventrepo.New` |
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
`AnalyticSecRulesEvent`, `AnalyticSecRateLimitsEvent`,
`AnalyticSecRulesetsEvent`, `AnalyticHttpIngressEvent`,
`AnalyticHttpEgressEvent`) lấy theo các trường
của message Kafka: mảng là `TEXT[]`/`INTEGER[]`, `action` và `tags` là
`JSONB`, enum là `INTEGER`, `matched_time` (mili giây) là `BIGINT`, có index
`BTREE` trên cột thời gian, index phức hợp trên `(resource_id, <time>)`,
và `request_id` trên các bảng event quy tắc (`secruleEvents` và
`AnalyticSecRulesetsEvent`; bảng HTTP được bao phủ bởi index unique `(id, time)`). Bảng ingress có các trường
scalar của `httppb.IngressEvent` (`status` là `INTEGER`, `asn` và `timestamp` (mili
giây) là `BIGINT`, còn lại là `TEXT`, header không gồm các header chứa
credential); `resource_id`/
`tenant_id` của nó do stage `DomainBased` của edge đặt
(`SetResourceId`/`SetTenantId`). Chúng không nằm trong `schema.All()`:
centraldata không giữ event trong bộ nhớ. Hằng tên cột lấy từ
`event_sz.pb.go`, sinh bởi `generate.sh` của `types/secrule/v1`,
`types/ruleset/v1` và `types/net/http/v1`.

Bảng egress có một dòng cho mỗi request mà edge proxy tới upstream origin
(`httppb.EgressEvent`): `id` chính là id của event ingress của request đó,
nên hai bảng nối với nhau qua `id`; tiếp theo là `upstream` (vd
`http://10.0.0.1:8080`), `method`, `scheme` (của upstream), `path`,
`queries`, `status` (của origin, 502 khi không kết nối được), `content_type`,
`content_length` (-1 khi stream), `latency_micros`, `timestamp` (lúc gửi
request, mili giây), `resource_id`, `tenant_id`, `protocol` (phiên bản
HTTP client dùng để tới edge), và `headers` (các request header edge đặt
cho upstream, sắp xếp theo tên, lược bỏ header nhạy cảm). Request do edge
tự trả lời (cache, static, bị chặn) không có dòng egress.

Mọi bảng event đều là hypertable
(`schema.EventHypertable(schema, timeColumn)`), chia theo cột thời gian
(`matched_time`, `timestamp`) thành chunk một ngày, và giữ trong
`schema.EventRetention`, 30 ngày.

Repository ([apps/collector/v1/repos](../../staging/src/github.com/sentinez/controlplane/apps/collector/v1/repos/)) nhúng `*eventreader.Reader[T]` ([eventreader](../../staging/src/github.com/sentinez/controlplane/apps/collector/v1/repos/eventreader/)):

- `Create(ctx, events...)` insert các event (bắt buộc đã có id) trong một
  câu lệnh; `Get`, `Delete`, `WithTx`, `Select()`.
- `Histogram(ctx, *histogram.Filter, *eventfilter.Filter)`: số lượng theo
  bucket thời gian, xem [Histogram](#histogram); filter thứ hai chỉ thu hẹp
  theo request và field, có thể là nil.
- `List(ctx, *eventfilter.Filter)`: lọc theo `ResourceIds`, `RequestIds`, và theo
  `matched_time` (`From` tính cả mốc, `To` không tính; event ingress dùng
  `timestamp`), phân trang bằng `Page`, sắp theo `matched_time DESC`.
- `TopValues(ctx, *eventfilter.Filter, names, limit)`: đếm các giá trị phổ biến
  nhất của các trường trên tập event đã lọc.
- `eventfilter` so sánh trường theo đúng kiểu dữ liệu cột (`INTEGER`,
  `INTEGER[]`, `BOOLEAN`, `TEXT`; `eventfilter.TypedColumn(name, type)`, kiểu
  mảng thì khớp bất kỳ phần tử nào, `Field.IsArray()`) thay vì so sánh chuỗi,
  và `Filter.Cond`
  biên dịch các điều kiện thành `sq.Sqlizer`. `eventfilter.Selection` là
  interface được thoả mãn bởi cả các field phẳng của request và `EventSelection`.
- Repo egress cung cấp thêm `LatencyReader` (xem [Analytic: đọc event](#analytic-đọc-event)).
- Cột và việc scan dòng dùng các hàm được sinh `<Event>Columns()`,
  `<Event>Values` (cho `Create`, qua `postgres.InsertRowsBuilder`) và
  `ScanOne<Event>`/`Scan<Event>`.

#### Histogram

[pkg/szstorage/dbx/timescale/histogram](../../staging/src/github.com/sentinez/pkg/szstorage/dbx/timescale/histogram/)
giữ, cho mỗi bảng event, một continuous aggregate cho mỗi
`HistogramInterval`, đếm số dòng theo bucket, resource và key; mỗi
`HistogramResolution` (khoảng thời gian một chart phủ) đọc view của
interval của nó, các resolution cùng interval dùng chung view:

| Resolution | Interval (hậu tố view) | Số bucket |
|---|---|---|
| 5 phút | 10 giây (`_10s`) | 30 |
| 15 phút | 30 giây (`_30s`) | 30 |
| 30 phút | 1 phút (`_1m`) | 30 |
| 1 giờ | 1 phút (`_1m`) | 60 |
| 3 giờ | 5 phút (`_5m`) | 36 |
| 6 giờ | 10 phút (`_10m`) | 36 |
| 12 giờ | 15 phút (`_15m`) | 48 |
| 24 giờ | 30 phút (`_30m`) | 48 |
| 2 ngày | 1 giờ (`_1h`) | 48 |
| 7 ngày | 3 giờ (`_3h`) | 56 |
| 30 ngày | 12 giờ (`_12h`) | 60 |

Ví dụ `dev_sentinez_analytic_httpingress_events_1m`. View được làm mới
theo độ rộng bucket (10 giây tới 30 phút), tới một phút trước (hoặc một
bucket, nếu ngắn hơn), trên 2 giờ gần nhất (`_lateEvents`; 3 ngày với
bucket giờ): các event collector ghi muộn, dưới watermark (ví dụ backlog
Kafka sau khi khởi động lại), được đếm trong khoảng đó. Không dài hơn vì
TimescaleDB invalidate cả khoảng thời gian của một transaction: một batch
lẫn một event cũ sẽ khiến mỗi lần refresh tính lại toàn bộ khoảng đó. View
giây giữ 7 ngày (`_secondRetention`), các view khác giữ
`schema.EventRetention`. View chạy real time, nên các bucket mới nhất gồm
cả những dòng chưa được làm mới. Các cột là `bucket` (unix mili giây),
`resource_id`, `tenant_id`, key và `total`.

Key chia nhỏ số đếm thêm: `status` http với event ingress và egress,
`behavior` với event ruleset, không có với event secrule và rate limit
(`Point.Key` bằng 0).

`Histogram.Points(ctx, *Filter{Resolution, ResourceIds, TenantId, From, To})`
cộng `total` qua các tenant và trả về `[]*Point{Time, ResourceId, Key,
Total}`, cũ nhất trước: các bucket của resolution, kết thúc ở bucket hiện
tại hoặc ở bucket chứa `To` (không tính). `Resolution` là
`typepb.HistogramResolution` của proto; khi để trống, khoảng thời gian sẽ
quyết định: resolution ngắn nhất có span chứa `From`..`To` (`To` mặc định
là hiện tại) với dung sai một bucket, bỏ qua view giây khi `From` cũ hơn 7
ngày, ngược lại là 30 ngày; 1 giờ khi không có `From`. `Filter.Resolve(now)`
tôn trọng `From` khi được đặt (căn chỉnh cả hai đầu theo biên bucket) và giới
hạn cửa sổ tối đa `MaxBuckets = 1000`, đẩy `From` muộn hơn và bật `Truncated = true`
nếu vượt quá; nó trả về `Window{Interval, From, To, Truncated}`, được các RPC
histogram và visitor của analytic phản chiếu trong các trường `interval`, `from`, `to`,
và `truncated` của response. Các RPC này từ chối `from`
hoặc `to` cũ hơn `schema.EventRetention` (`INVALID_ARGUMENT`).

Các view không thu hẹp được theo request hay field, nên
`Histogram.PointsWhere(ctx, *Filter, sq.Sqlizer)` đếm trực tiếp các dòng của bảng
nguồn (`SELECT time_bucket(<interval>, <time>), resource_id, key, count(*) ...
GROUP BY`), `sq.Sqlizer` thêm các điều kiện. Cửa sổ, độ phân giải và các bucket
rỗng giống `Points`. `eventreader.Reader[T].Histogram` chọn giữa hai cách:
dùng view khi filter `events` là nil hoặc không có điều kiện, ngược lại là
`PointsWhere` với `events.Cond(spec.Columns)`.

### Analytic: đọc event

Analytic đọc các bảng của collector qua
[apps/analytic/v1/repos/events](../../staging/src/github.com/sentinez/controlplane/apps/analytic/v1/repos/events/):
`eventsrepo.Events` giữ một `Reader[T]` chỉ đọc (`List`, `Histogram`,
`TopValues`) cho mỗi bảng, dựng từ các repo của collector, kèm `IngressReader`
(bổ sung `VisitorReader`) và `EgressReader` (bổ sung `LatencyReader`).
Khi dựng, nó sync bảng và view như collector.

| RPC | REST | Đọc |
|---|---|---|
| `ListHttpIngressEvents` | `GET /analytic/events/http-ingress` | `analytic.httpingress.events` |
| `ListHttpEgressEvents` | `GET /analytic/events/http-egress` | `analytic.httpegress.events` |
| `ListSecRuleEvents` | `GET /analytic/events/secrule` | `analytic.secrules.events` |
| `ListSecRateLimitEvents` | `GET /analytic/events/secratelimit` | `analytic.secratelimits.events` |
| `ListSecRulesetEvents` | `GET /analytic/events/secruleset` | `analytic.secrulesets.events` |
| `GetEventHistogram` | `GET /analytic/events/histogram` | các view hoặc bảng của `kind` |
| `GetEventTopValues` | `GET /analytic/events/top-values` | bảng của `kind` |
| `GetEgressLatency` | `GET /analytic/events/http-egress/latency` | `analytic.httpegress.events` |

- Request list có `page`, `resource_ids`, `request_ids`, `filters`, `from` và `to`;
  event mới nhất trước. Một trang có 50 event khi không truyền, tối đa 500.
- `GetEventHistogram` nhận `kind` (`EventKind`), `resolution`
  (`HistogramResolution`), `resource_ids`, `request_ids`, `filters`, `from` và `to`,
  trả về `HistogramPoint{time, resource_id, key, total}`, cũ nhất trước, cùng
  với `from`, `to`, `interval`, và `truncated`.
- `GetEventTopValues` đếm số lần xuất hiện của các trường trong khoảng `from`..`to`,
  lọc theo `resource_ids`, `request_ids`, và `filters`.
- `GetEgressLatency` nhận `EventSelection` (mặc định 24h gần nhất), gom nhóm
  theo `upstream`, trả về `total`, `avg_micros`, `p50_micros`, `p95_micros`,
  `p99_micros`, và `max_micros` (tính bằng hàm `percentile_cont` của PostgreSQL).
- `EventSelection` là message chuẩn đại diện cho các tiêu chí lọc resource, request,
  filter và khoảng thời gian, dùng chung giữa các RPC analytic.
- Mọi RPC cần token. Việc đọc bị giới hạn theo resource của các tenant mà
  người gọi là member: không truyền `resource_ids` thì chỉ trả event của
  chúng, và người gọi không có resource nào nhận kết quả rỗng. Control plane
  `ADMIN` đọc được tất cả.

#### Pool event

Sink decode vào các message được tái sử dụng
([controlplane/pkg/pools](../../staging/src/github.com/sentinez/controlplane/pkg/pools/)):

| Package | Message | Sink |
|---|---|---|
| `pools/secrule` | `secrulepb.SecRuleEvent` | secrule, ratelimit |
| `pools/secruleset` | `rulesetpb.SecRulesetEvent` | rulesets |
| `pools/httpingress` | `httppb.IngressEvent` | ingress |
| `pools/httpegress` | `httppb.EgressEvent` | egress |

Mỗi package có `Pool()`, `Acquire()`, `Release(obj)`, dựa trên pool mà
vtproto sinh cho message (`<Msg>FromVTPool`, `ReturnToVTPool`, xem
[8.3](08-api-protobuf.md#83-sinh-code)). `ReturnToVTPool` gọi `ResetVT`, hàm
này giữ dung lượng của các trường repeated cho lần `UnmarshalVT` sau. Sink
gom event của một batch vào `pools.Batch` và `defer` `Close` của nó để trả
tất cả về pool.

Cấu hình: `SZ_KAFKA_BROKERS` (không đặt: ingest bị tắt kèm cảnh báo,
gRPC server vẫn chạy) và database của `SZ_POSTGRES_URI`. Các giới hạn
giữ bộ nhớ của consumer nằm ở
[09-core-shared.md](09-core-shared.md#szkafka).

## 7.7 Mã lỗi

[pkg/szerror](../../staging/src/github.com/sentinez/pkg/szerror/errors.go)
cung cấp gRPC status có định dạng `SZ-<code>: <ERRORS_...>` và các hàm
`StatusXxxF(format, ...)` (`NotFound`, `Unauthorized`, `Forbidden`,
`InvalidData`/`InvalidArgument`, `AlreadyExists`, `Internal`,
`Unimplemented`). Sentinel `ErrNotFound`, `ErrInvalidData`, `ErrUnimplemented`
được gateway map sang HTTP. `IsNoRows(err)` nhận diện `pgx.ErrNoRows`.

## 7.8 Kiểm thử

- Mock repository sinh bằng mockery: `iam/v1/repos/{accounts,users}/mock`,
  `tenant/v1/repos/resources/mock`.
- `iam/v1/service/service_test.go`, `iam/v1/handler/iam_test.go`,
  `pkg/crypto/crypto_test.go`.
