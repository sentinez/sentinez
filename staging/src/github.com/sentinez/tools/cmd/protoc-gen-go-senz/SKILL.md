---
name: protoc-gen-go-senz
version: 0.1.0
description: A protoc plugin that generates Go helpers for Sentinez services, including metadata accessors and field name constants.
---

# protoc-gen-go-senz

`protoc-gen-go-senz` is a custom Protocol Buffers plugin for Go, designed to generate helper code for Sentinéz applications. It works alongside `protoc-gen-go` to provide additional type safety and metadata accessors.


## Style
Read `github.com/sentinez/docs/guide/uber-guide.md` for style guide.

## Features

### 1. Service Metadata Accessors
If a `.proto` file contains the `(sentinez.types.v1.x_meta)` file option, the plugin generates:
- A private metadata struct.
- Public getter functions to access service details (`ServiceName`, `ServiceKind`, `ServiceKey`).

**Example Protobuf:**
```protobuf
option (sentinez.types.v1.x_meta) = {
  service_name: "Greeter"
  service_kind: KIND_SERVICE
  service_key:  "greeter"
};
```

**Generated Go:**
```go
func GetMetaGreeter() *typepb.XMeta { ... }
func GetMetaGreeterServiceName() string { return "Greeter" }
// ... other getters
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

Use it as a standard `protoc` plugin:

```bash
protoc --go-senz_out=. --go-senz_opt=paths=source_relative <proto_files>
```

It creates files with the suffix `_senz.pb.go`.

### 4. Method Requirement Accessors
For methods annotated with `(sentinez.types.v1.x_method).require`, the plugin gets the list of requirements for that method.
If the list is not empty, the plugin generates a function returning the list of requirements for that method.

**Format:** `GetReq<ServiceName><MethodName>`

**Example Protobuf:**
```protobuf
service Greeter {
  rpc SayHello (HelloRequest) returns (HelloReply) {
    option (sentinez.types.v1.x_method).require = {
      role: ROLE_MEMBER
      permission: PERMISSION_CREATE_ANY
    };
  }
}
```

**Generated Go:**
```go
func GetReqGreeterSayHello() []*typepb.XRequire {
	return []*typepb.XRequire{
		{
			Role:       typepb.Role_ROLE_MEMBER,
			Permission: typepb.Permission_PERMISSION_CREATE_ANY,
		},
	}
}
```
