---
name: create-module
description: Instructions for creating a base core module entrypoint file following the standard sentinez pattern (e.g., iam.go). Use this when tasked with creating the root orchestrating package for a new core domain.
---

# Creating a Standard Sentinez Core Module Entrypoint

**CRITICAL PREREQUISITE:** Before generating or implementing the module entrypoint, you MUST read and apply the rules from the `go-style-guide` skill. All generated code must strictly follow the Uber Go Style Guide conventions.

**PROJECT CONVENTIONS (enforced by `.golangci.yaml`):** every Go file starts with the Apache 2.0 header (`// Copyright 2025 Duc-Hung Ho.`, copy it from any existing file); lines are at most 80 columns (tab = 4) and functions at most 35 lines (`funlen`; use `// nolint:funlen` only when unavoidable, as the iam package does). Verify with `cd staging/src/github.com/sentinez/controlplane && golangci-lint run`.

**DOMAIN DEFINITION:** The protobuf definitions for the domain live in the `api` module at `api/proto/sentinez/apps/<domain>/v1/` (`<domain>.proto` for the service and `model.proto` for models, e.g. `apps/iam/v1/iam.proto`). Review them to understand the service interface, endpoints, and models. Generated code (`*.<domain>pb.go`, `*_grpc.<domain>pb.go`, `*_senz.<domain>pb.go`, `*.<domain>pb.gw.go`, `*.<domain>pb.validate.go`) is produced by that directory's `generate.sh` / `cd api && buf generate`; never edit it by hand. Proto packages are imported as `github.com/sentinez/sentinez/api/proto/sentinez/apps/<domain>/v1` with alias `<domain>pb`. The gRPC service name comes from the proto (e.g. `IdentityAccessManagementService`), so use the real generated names rather than `<Domain>Service` literally.

When asked to create the base core service/module entrypoint for a given functional domain, you must follow the standard orchestrating pattern established in the `sentinez` project (such as in `github.com/sentinez/controlplane/iam/v1/iam.go`). This file binds the gRPC handlers to a local buffer listener (used for internal/gateway communication).

## 1. File Structure and Package

The module entrypoint should be placed at the root of the domain version package: `github.com/sentinez/controlplane/<domain>/v1/<domain>.go`.
The package name should be `<domain>`.

Use standard imports, especially:
- Context from `context`
- Local client buffer configurations from `github.com/sentinez/sentinez/api/client/local`
- Protobuf generated code from `github.com/sentinez/sentinez/api/proto/sentinez/apps/<domain>/v1` (aliased as `<domain>pb`)
- Configuration types from `github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1` (aliased as `settingpb`)
- The factory package from `github.com/sentinez/controlplane/<domain>/v1/factory` (aliased as `<domain>fac`)
- Default gRPC server utilities from `github.com/sentinez/core/grpc` (aliased as `coregrpc`)
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

Define the main module struct (usually named `<Domain>`, e.g. `IAM`, `Analytic`). It embeds `*coregrpc.Server` allowing it to function as a server, and it holds the gRPC handler interface.

```go
type <Domain> struct {
	*coregrpc.Server
	hdl <domain>pb.<Domain>ServiceServer
}
```

## 4. Constructor (`NewService`)

Define the constructor `NewService` (this is what `cmd/` / `pkg/apps` inject into the runner) which orchestrates setting up the network server and the domain handlers using the factory.

```go
func NewService(ctx context.Context, appConf *settingpb.Config) *<Domain> {
	return &<Domain>{
		Server: coregrpc.New(coregrpc.WithXMeta(appConf.GetMeta())),
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
	<domain>pb.Register<Domain>ServiceServer(mod.AsServer(), mod.hdl)

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
