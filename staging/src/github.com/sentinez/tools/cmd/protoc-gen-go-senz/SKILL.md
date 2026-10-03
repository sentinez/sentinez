---
name: protoc-gen-go-senz
version: 0.2.0
description: A protoc plugin that generates Go helpers for Sentinez services, including metadata accessors and field name constants.
---

# protoc-gen-go-senz

`protoc-gen-go-senz` is a custom Protocol Buffers plugin for Go, designed to generate helper code for Sentinéz applications. It works alongside `protoc-gen-go` to provide additional type safety and metadata accessors.


## Style
Read `.agent/skills/go-style-guide/SKILL.md` (Uber-based; 80 columns, 35-line functions) before changing the plugin. The generator logic lives in `tools/internal/senz/senz.go`; `main.go` only wires it to `protogen`.

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

## Usage

Install the plugin so `protoc-gen-go-senz` is on `PATH` (`cd staging/src/github.com/sentinez/tools && go install ./cmd/protoc-gen-go-senz`), then use it as a standard `protoc` plugin. Each `api/proto/sentinez/**/v1/generate.sh` already does this:

```bash
protoc --go-senz_out="$SENTINEZ_GEN_OUT" "$(pwd)"/*.proto
```

It creates `<file>_senz.pb.go` next to the other generated files (e.g. `api/proto/sentinez/apps/iam/v1/iam_senz.pb.go`). Never edit it by hand. Generated files import `typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"`, which defines `XMeta`, `XMessage`, `XMethod`, `Zone` and `Console` (`types/v1/options.proto`, `known.proto`).

### 4. Method Metadata Accessors
For methods annotated with `(sentinez.types.v1.x_method)`, the plugin generates a function returning the method metadata.
If `ignore` is false and `consoles` is empty, no getter is generated. `Ignore: true` is emitted only when set, and `Consoles` only when non-empty. Handlers use these getters for permission checks, e.g. `ss.Check(iampb.GetIdentityAccessManagementServiceListUsers())`.

**Format:** `Get<GoServiceName><MethodName>` (`GoServiceName` is the generated Go name of the proto service, e.g. `GreeterService`)

**Example Protobuf:**
```protobuf
service GreeterService {
  rpc SayHello (HelloRequest) returns (HelloReply) {
    option (sentinez.types.v1.x_method) = {
      ignore: false
      consoles: [CONSOLE_PORTAL, CONSOLE_ADMIN] // or `ignore: true`
    };
  }
}
```

**Generated Go:**
```go
func GetGreeterServiceSayHello() *typepb.XMethod {
	return &typepb.XMethod{
		Consoles: []typepb.Console{
			typepb.Console_CONSOLE_PORTAL,
			typepb.Console_CONSOLE_ADMIN,
		},
	}
}
```

## Testing

Run `cd staging/src/github.com/sentinez/tools && go test ./...`. `tools/example/` holds a sample proto with its `_senz.pb.go` output (`example/build.sh` regenerates it); after changing the generator, regenerate it and the affected `api` protos (`cd api && buf generate` or the directory's `generate.sh`) and check that `go build ./...` still passes in the root and `api` modules.
