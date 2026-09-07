---
name: create-module
description: Instructions for creating a base core module entrypoint file following the standard sentinez pattern (e.g., iam.go). Use this when tasked with creating the root orchestrating package for a new core domain.
---

# Creating a Standard Sentinez Core Module Entrypoint

**CRITICAL PREREQUISITE:** Before generating or implementing the module entrypoint, you MUST read and apply the rules from the `go-style-guide` skill. All generated code must strictly follow the Uber Go Style Guide conventions.

**DOMAIN DEFINITION:** The protobuf definitions for the domain can be found at `api/proto/sentinez/core/<domain>/v1/<domain>.proto`. Please review it to understand the service interface, endpoints, and models.

When asked to create the base core service/module entrypoint for a given functional domain, you must follow the standard orchestrating pattern established in the `sentinez` project (such as in `github.com/sentinez/modules/iam/v1/iam.go`). This file binds the gRPC handlers to a local buffer listener (used for internal/gateway communication).

## 1. File Structure and Package

The module entrypoint should be placed at the root of the domain version package: `github.com/sentinez/modules/<domain>/v1/<domain>.go`.
The package name should be `<domain>`.

Use standard imports, especially:
- Context from `context`
- Local client buffer configurations from `github.com/sentinez/sentinez/api/client/local`
- Protobuf generated code from `github.com/sentinez/sentinez/api/modules/<domain>/v1` (aliased as `pb` or `<domain>pb`)
- Configuration types from `github.com/sentinez/sentinez/api/types/conf/v1` (aliased as `confpb`)
- The factory package from `github.com/sentinez/sentinez/github.com/sentinez/modules/<domain>/v1/factory` (aliased as `<domain>fac`)
- Default gRPC server utilities from `github.com/sentinez/sentinez/pkg/network/grpc` (aliased as `netgrpc`)
- Buffer connection from `google.golang.org/grpc/test/bufconn`

## 2. Buffer Listener Setup

Declare a package-level global variable for the `bufconn.Listener` and a getter function `GetListener()` so the API Gateway or local test clients can connect to it.

```go
package <domain>

import (
    "context"
    // Other imports...
)

var bufLis *bufconn.Listener

func GetListener() *bufconn.Listener {
	return bufLis
}
```

## 3. Module Struct Definition

Define the main module struct (usually named `<Domain>`, e.g. `IAM`, `Analytic`). It embeds `*netgrpc.Server` allowing it to function as a server, and it holds the gRPC handler interface.

```go
type <Domain> struct {
	*netgrpc.Server
	hdl pb.<Domain>ServiceServer
}
```

## 4. Constructor (`NewService`)

Define the constructor `NewService` which orchestrates setting up the network server and the domain handlers using the factory.

```go
func NewService(ctx context.Context, appConf *confpb.Config) *<Domain> {
	return &<Domain>{
		Server: netgrpc.NewDefault(appConf.GetMeta()),
		hdl:    <domain>fac.NewDefaultHandler(ctx, appConf),
	}
}
```

## 5. Startup Routine (`Start`)

Implement the `Start(ctx context.Context) error` method for the struct. This method must:
1. Register the handler onto the built-in gRPC server using the generated protobuf registration method.
2. Initialize the buffer listener with sizes defined in `local.BufSize`.
3. Serve traffic on that listener.

```go
func (mod *<Domain>) Start(_ context.Context) error {
	// Register the handler. Replace 'Register<Domain>ServiceServer' with the actual method name.
	pb.Register<Domain>ServiceServer(mod.AsServer(), mod.hdl)

	// Begin listening on the buffer
	bufLis = bufconn.Listen(local.BufSize)
	
    // Serve requests over the buffer connections
	return mod.BufServe(bufLis)
}
```

## Rules to Remember:
1. **Follow Go Style Guide:** Always apply the coding conventions from the `go-style-guide` skill.
2. Keep the module file clean and devoid of business logic; it strictly orchestrates startup requirements.
3. Don't forget `mod.AsServer()` when registering the generated gRPC server.
