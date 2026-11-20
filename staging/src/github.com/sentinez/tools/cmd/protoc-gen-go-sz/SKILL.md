---
name: protoc-gen-go-sz
version: 0.2.0
description: A protoc plugin that generates Go helpers for Sentinez services, including metadata accessors and field name constants.
---

# protoc-gen-go-sz

`protoc-gen-go-sz` is a custom Protocol Buffers plugin for Go, designed to generate helper code for Sentinéz applications. It works alongside `protoc-gen-go` to provide additional type safety and metadata accessors.


## Style
Read `hack/.agent/skills/go-style-guide/SKILL.md` (Uber-based; 80 columns, 35-line functions) before changing the plugin. The generator logic lives in `tools/internal/sz/sz.go`; `main.go` only wires it to `protogen`.

## Features

### 1. Service Metadata Accessors
If a `.proto` file contains the `(sentinez.types.v1.x_meta)` file option, the plugin generates:
- A private metadata struct.
- Public getter functions to access service details (`ServiceName`, `ServiceZone`, `ServiceKey`).

**Example Protobuf:**
```protobuf
option (sentinez.types.v1.x_meta) = {
  service_name: "SENTINEZ // GREETER"
  service_zone: ZONE_INTERNAL
  service_key:  "sentinez.apps.greeter.v1"
};
```

`service_zone` is the `sentinez.types.v1.Zone` enum: `ZONE_DEMILITARIZED`, `ZONE_INTERNAL`, `ZONE_PRIVATE_API`, `ZONE_PUBLIC_API` (`ZONE_UNSPECIFIED` = 0). The old `service_kind` / `Kind` field no longer exists.

`<Name>` in the getters is the title-cased **proto file base name** (`greeter.proto` -> `Greeter`, `iam.proto` -> `Iam`), not the gRPC service name.

**Generated Go:**
```go
func GetMetaGreeter() *typepb.XMeta { ... } // returns a proto.Clone
func GetMetaGreeterServiceName() string { ... }
func GetMetaGreeterServiceZone() typepb.Zone { ... }
func GetMetaGreeterServiceKey() string { ... }
```

### 2. Exported Field Constants
For messages annotated with `(sentinez.types.v1.x_message).export_field = true`, the plugin generates string constants for each field name. This is useful for reflection or API field filtering.

**Format:** `X<MessageName>_<FieldName>` (PascalCase)

**Example Protobuf:**
```protobuf
message User {
  option (sentinez.types.v1.x_message).export_field = true;
  string user_id = 1;
}
```

**Generated Go:**
```go
const (
	XUser_UserId = "user_id"
)
```

### 3. Database Model Field Constants
For messages annotated with `(sentinez.types.v1.x_message).database_model = true`, the plugin generates string constants for field names, typically used for database queries (ORM-like helpers).

**Format:** `<MessageName>_<FieldName>` (PascalCase)

**Example Protobuf:**
```protobuf
message User {
  option (sentinez.types.v1.x_message).database_model = true;
  string email = 1;
}
```

**Generated Go:**
```go
const (
	User_Email = "email"
)
```

### 4. Flattened Field Constants
A message field annotated with `(sentinez.types.v1.x_field).flatten = true` gets no constant of its own; instead the plugin emits one constant per field of the nested message, named `<field>_<nested field>`. It applies to both `export_field` and `database_model` constants, and nests when the inner field is flattened too. The option is ignored on scalar, `repeated` and `map` fields, and a recursive message is not expanded a second time.

**Example Protobuf:**
```protobuf
message SecRule {
  option (sentinez.types.v1.x_message).database_model = true;
  sentinez.types.v1.Owner owner = 1 [(sentinez.types.v1.x_field).flatten = true];
  string id = 2;
}
```

**Generated Go:**
```go
const (
	SecRule_OwnerCreatedAt  = "owner_created_at"
	SecRule_OwnerUpdatedAt  = "owner_updated_at"
	SecRule_OwnerResourceId = "owner_resource_id"
	SecRule_OwnerTenantId   = "owner_tenant_id"
	SecRule_Id              = "id"
)
```

## Usage

Install the plugin so `protoc-gen-go-sz` is on `PATH` (`cd staging/src/github.com/sentinez/tools && go install ./cmd/protoc-gen-go-sz`), then use it as a standard `protoc` plugin. Each `api/proto/sentinez/**/v1/generate.sh` already does this:

```bash
protoc --go-sz_out="$SENTINEZ_GEN_OUT" "$(pwd)"/*.proto
```

It creates `<file>_sz.pb.go` next to the other generated files (e.g. `api/proto/sentinez/apps/iam/v1/iam_sz.pb.go`). Never edit it by hand. Generated files import `typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"`, which defines `XMeta`, `XMessage`, `XField`, `XMethod`, `Zone` and `ControlPlane` (`types/v1/options.proto`, `known.proto`).

### 5. Method Metadata Accessors
For methods annotated with `(sentinez.types.v1.x_method)`, the plugin generates a function returning the method metadata.
If none of `ignore_control_plane`, `public` and `audit` is set and `control_planes` is empty, no getter is generated. `IgnoreControlPlane`, `Public`, `Audit` and `AuditNoOwner` are emitted only when true, and `ControlPlanes` only when non-empty. Handlers use these getters for permission checks, e.g. `ss.Check(iampb.GetIdentityAccessManagementServiceListUsers())`.

**Format:** `Get<GoServiceName><MethodName>` (`GoServiceName` is the generated Go name of the proto service, e.g. `GreeterService`)

**Example Protobuf:**
```protobuf
service GreeterService {
  rpc SayHello (HelloRequest) returns (HelloReply) {
    option (sentinez.types.v1.x_method) = {
      control_planes: [CONTROL_PLANE_PORTAL, CONTROL_PLANE_ADMIN] // or `ignore_control_plane: true`
    };
  }
}
```

**Generated Go:**
```go
func GetGreeterServiceSayHello() *typepb.XMethod {
	return &typepb.XMethod{
		ControlPlanes: []typepb.ControlPlane{
			typepb.ControlPlane_CONTROL_PLANE_PORTAL,
			typepb.ControlPlane_CONTROL_PLANE_ADMIN,
		},
	}
}
```

## Testing

Run `cd staging/src/github.com/sentinez/tools && go test ./...`. `tools/example/` holds a sample proto with its `_sz.pb.go` output (`example/build.sh` regenerates it); after changing the generator, regenerate it and the affected `api` protos (`cd api && buf generate` or the directory's `generate.sh`) and check that `go build ./...` still passes in the root and `api` modules.
