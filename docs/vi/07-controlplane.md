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

Scaffold mẫu cho từng lớp nằm trong `.agent/skills/` (`create-service`,
`create-handler`, `create-repo`, `create-factory`, `create-module`).

## 7.1 Lớp lưu trữ (`core/storage/dbx`)

[core/storage/dbx](../../staging/src/github.com/sentinez/core/storage/dbx/):

- `dbx.Database[T]`: `Insert`, `Select`, `Delete`, `Total`, `Table`, cùng
  `Exec`, `Query`, `CollectRows`, `CollectOneRow` nhận `query.Query`
  (interface `ToSql()` của squirrel).
- `postgres.New[T](ctx, conf, WithTable(name), WithColumns(cols))`:
  - Tên bảng thật = `<mode>.sentinez.<name>` thay `.` bằng `_`, ví dụ
    `dev_sentinez_iam_users`; phải khớp regex
    `^(dev|sandbox|prod)_sentinez_[a-z]+(_[a-z]+)*$`.
  - Tự tạo bảng (`id TEXT PK, created_at, updated_at`) và thêm cột còn thiếu
    — **không xoá/đổi kiểu cột**.
  - Dùng một `pgxpool.Pool` toàn cục từ `SENZ_POSTGRES_URI`.
- Helper: `SelectBuilder` (+ phân trang `LIMIT/OFFSET ORDER BY created_at DESC`
  khi `page.index` và `page.size` > 0), `InsertBuilder` (`RETURNING id`),
  `UpdateBuilder` (tự set `updated_at`).
- Transaction: `postgres.NewTX(conf).Begin(ctx)` → `*TxSession`;
  `postgres.WithTx(ss, db)` trả về `Database[T]` chạy trong transaction; repo
  cung cấp `WithTX(ss)`.
- ID bản ghi: `rand.NewID(table.NewPrimaryKey(name))` =
  `senz.<name>.<uuid>`.

Tên bảng ([core/common/tables](../../staging/src/github.com/sentinez/core/common/tables/tables.go)):
`iam.users`, `iam.accounts`, `tenant.resources`, `analytic.activities`,
`security.sec_rules`.

Hằng tên cột (`User_Email`, `SecRule_Expr`, ...) được sinh bởi
`protoc-gen-go-senz` cho message có option `x_message.database_model = true`.

## 7.2 Xác thực & phân quyền

[controlplane/pkg](../../staging/src/github.com/sentinez/controlplane/pkg/):

| Gói | Chức năng |
|---|---|
| `crypto` | `TokenGenerator(*typepb.Context)` ký JWT HS256 với secret = base64-decode(`SENZ_SECRET_KEY`); claim `auth` = `Context.String()` (prototext), `exp`. `BearerTokenVerifier` bỏ tiền tố `Bearer `, kiểm tra HMAC, parse lại `Context`. `HashPassword`/`CheckPasswordHash` dùng bcrypt |
| `headers` | `GetAuth(ctx)` đọc metadata `Authorization`, verify → `Auth{*typepb.Context}`; `Auth.Check(method)` gọi `perms.Allow` |
| `passkey` | `NewWebAuthn(conf)` (RP = `core.Name`, RPID = `SENZ_HOSTNAME`, origin = `SENZ_CLIENT_ORIGIN`); `Store` in-memory (session WebAuthn + account tạm) |

`typepb.Context` = `{name, expire_at, user_id, control_plane}`; JWT hết hạn
sau 1 giờ.

Phân quyền dựa trên **control plane** (`PORTAL` / `ADMIN`):

- Mỗi RPC có thể khai báo `option (sentinez.types.v1.x_method) = {ignore |
  control_planes: [...]}`; codegen sinh hàm `Get<Service><Method>()`.
- `perms.Allow(method, cp)`: cho phép nếu method nil, `ignore`, danh sách rỗng
  hoặc `cp` nằm trong danh sách.
- Handler phải **tự gọi** `headers.GetAuth` + `Check` — hiện chỉ
  `IAM.ListAccounts` và `Analytic.ListActivities` làm việc này. Không có
  interceptor gRPC chung.

## 7.3 IAM

Proto: [apps/iam/v1](../../api/proto/sentinez/apps/iam/v1/). Bảng: `iam.users`
(`id, email, full_name, control_plane`) và `iam.accounts`
(`id, email, username, password_hash, credentials TEXT[], user_id, provider,
provider_user_id`). Một `User` có thể có nhiều `Account`.

| RPC | REST | Hành vi |
|---|---|---|
| `Login` | `PUT /iam/login` | Nếu username = `SENZ_ADMIN_USERNAME`: so mật khẩu với `SENZ_ADMIN_PASSWORD`, cấp token ADMIN (`user_id = sentinez.admin`). Ngược lại: tìm account theo username/email, bcrypt, lấy user, cấp token PORTAL |
| `CreateAccount` | `POST /iam/account` | Kiểm tra trùng username/email, trong 1 transaction tạo `User` (PORTAL) + `Account` (bcrypt) |
| `CreateUser` | — (chỉ gRPC) | Tạo `User` nếu email chưa có |
| `GetUser` | `GET /iam/user` | Theo `id` hoặc `email` |
| `ListUsers` | `GET /iam/users` | Lọc `ids`, `emails`, phân trang |
| `UpdateUser` | — | Cập nhật `full_name`, `email` |
| `DeleteUser` | `DELETE /iam/user/{id}` | Xoá user |
| `ListAccounts` | `GET /iam/accounts` | Yêu cầu token; lọc `ids/emails/user_ids/usernames` |
| `Status` | `GET /iam/status` | Gọi Greeter.Status qua bufconn, trả về `Context` từ token (nếu có) |
| `PasskeyRegisterChallenge` | `GET /iam/passkey/register/challenge` | Báo lỗi nếu username/email đã tồn tại; tạo account tạm trong bộ nhớ, `BeginRegistration`, lưu session, trả `options` + `session_id` |
| `PasskeyRegisterVerify` | `POST /iam/passkey/register/verify` | Lấy session + account tạm, tạo account thật (mật khẩu ngẫu nhiên) nếu chưa có, `CreateCredential`, lưu credential JSON vào `credentials` |
| `PasskeyLoginChallenge` | `GET /iam/passkey/login/challenge` | `BeginLogin` cho account, lưu session |
| `PasskeyLoginVerify` | `PUT /iam/passkey/login/verify` | `ValidateLogin`, cập nhật credential (sign count), cấp token (ADMIN nếu username `admin`) |

`AccountX` bọc `iampb.Account` để cài `webauthn.User` (WebAuthn ID = email).

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
| Analytic | `ListActivities` (`GET /analytic/activities`, cần token) trên bảng `analytic.activities` (`id, resource_id, unique_visitor INTEGER[]`). Chưa đăng ký vào apiserver |
| Greeter | Service mẫu: `SayHello` (`GET /greeter/say?name=`) trả lời kèm thời gian; `Status`. Có binary riêng `szgreeter` và cũng chạy in-process trong apiserver |
| CentralData | Chỉ có `Status`; binary `szcentraldata` chưa có `main` |

## 7.7 Mã lỗi

[shared/errorx](../../staging/src/github.com/sentinez/shared/errorx/errors.go)
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
