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
- ID bản ghi: `rand.NewId(table.NewPrimaryKey(name))` =
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
| `passkey` | `NewWebAuthn(conf)` (RP = `core.Name`, RPID = `SENZ_HOSTNAME`, origin = `SENZ_CLIENT_ORIGIN`); `Store` in-memory (session WebAuthn + `Registration` đang chờ, xem 7.3) |

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

Proto: [apps/iam/v1](../../api/proto/sentinez/apps/iam/v1/). Code:
[iam/v1](../../staging/src/github.com/sentinez/controlplane/iam/v1/) — logic
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

"Không tìm thấy" là `errorx.ErrNotFound`; helper `found` của service đổi nó
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
