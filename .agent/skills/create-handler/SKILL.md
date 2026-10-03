---
name: create-handler
description: Instructions for creating a base handler package following the standard sentinez pattern (e.g., iam handler). Use this when tasked with creating a new gRPC handler implementation that exposes a service.
---

# Creating a Standard Sentinez Handler Package

**CRITICAL PREREQUISITE:** Before generating or implementing the handler, you MUST read and apply the rules from the `go-style-guide` skill. All generated code must strictly follow the Uber Go Style Guide conventions.

**PROJECT CONVENTIONS (enforced by `.golangci.yaml`):** every Go file starts with the Apache 2.0 header (`// Copyright 2025 Duc-Hung Ho.`, copy it from any existing file); lines are at most 80 columns (tab = 4) and functions at most 35 lines (`funlen`; use `// nolint:funlen` only when unavoidable, as the iam package does). Verify with `cd staging/src/github.com/sentinez/controlplane && golangci-lint run`.

**DOMAIN DEFINITION:** The protobuf definitions for the domain live in the `api` module at `api/proto/sentinez/apps/<domain>/v1/` (`<domain>.proto` for the service and `model.proto` for models, e.g. `apps/iam/v1/iam.proto`). Review them to understand the service interface, endpoints, and models. Generated code (`*.<domain>pb.go`, `*_grpc.<domain>pb.go`, `*_senz.<domain>pb.go`, `*.<domain>pb.gw.go`, `*.<domain>pb.validate.go`) is produced by that directory's `generate.sh` / `cd api && buf generate`; never edit it by hand. Proto packages are imported as `github.com/sentinez/sentinez/api/proto/sentinez/apps/<domain>/v1` with alias `<domain>pb`. The gRPC service name comes from the proto (e.g. `IdentityAccessManagementService`), so use the real generated names rather than `<Domain>Service` literally.

When asked to create a new handler for a given functional domain, you must follow the standard handler pattern established in the `sentinez` project (such as in `github.com/sentinez/controlplane/iam/v1/handler/iam.go`). Most IAM methods only delegate (`return iam.service.X(ctx, req)`); add logging and the `headers.GetAuth` / `ss.Check` permission gate only for methods that need them. The handler acts as the gRPC server implementation that handles incoming requests, performs authorization checks, logs activities, and delegates business logic to the underlying service package.

## 1. File Structure and Package

The handler should be placed in an appropriate package under `github.com/sentinez/controlplane/<domain>/v1/handler`. The main file should typically be named `<domain>.go`.
The package name should be `<domain>hdl`.

Use standard imports, especially:
- Protobuf generated code from `github.com/sentinez/sentinez/api/proto/sentinez/apps/<domain>/v1` (aliased as `<domain>pb`)
- The service package from `github.com/sentinez/controlplane/<domain>/v1/service` (aliased as `<domain>svc`)
- Standard error handling from `github.com/sentinez/shared/errorx`
- Request context headers and auth from `github.com/sentinez/controlplane/pkg/headers`
- Logging from `github.com/sentinez/shared/zlog`

## 2. Interface Implementation Assertion

Ensure the handler struct implements the gRPC server interface generated from Protobuf.

```go
package <domain>hdl

import (
	<domain>pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/<domain>/v1"
	<domain>svc "github.com/sentinez/controlplane/<domain>/v1/service"
	// other imports...
)

var _ <domain>pb.<Domain>ServiceServer = (*<Domain>Handler)(nil)
```

## 3. Define the Struct and Constructor

The constructor initializes the handler with the business logic service and any other necessary gRPC client dependencies.

```go
type <Domain>Handler struct {
	service *<domain>svc.<Domain>Service
	// Add other gRPC clients here if necessary (e.g., greeterCli)
}

func New(
	service *<domain>svc.<Domain>Service,
	// Add other dependencies here...
) <domain>pb.<Domain>ServiceServer {
	return &<Domain>Handler{
		service: service,
	}
}
```

## 4. Implement RPC Methods

Each gRPC method defined in the proto file must be implemented by the handler struct. The methods should typically perform three steps: Logging, Authorization check (if needed), and Service Delegation.

### Basic Method Signature with Logging, Auth, and Delegation

```go
func (hdl *<Domain>Handler) <MethodName>(ctx context.Context, 
	req *<domain>pb.<MethodName>Request) (*<domain>pb.<MethodName>Response, error) {
	
	// 1. Log the incoming request
	zlog.Debugf("[<Handler>.<MethodName>] req = %v", req)

	// 2. Perform Authorization Check (only if the method requires specific permissions)
	ss, err := headers.GetAuth(ctx)
	if err != nil {
		return nil, err
	}
	
	// The permission getter is generated into the *_senz.pb.go file, e.g.
	// iampb.GetIdentityAccessManagementServiceListUsers().
	err = ss.Check(<domain>pb.Get<Service><MethodName>())
	if err != nil {
		return nil, err
	}

	// 3. Delegate to the service
	resp, err := hdl.service.<MethodName>(ctx, req)
	if err != nil {
		zlog.Errorf("[<Handler>.<MethodName>] error: %v", err)
		return nil, err
	}

	return resp, nil
}
```

### Methods Without Authorization
For public methods (e.g., Login, Status), you might skip the `headers.GetAuth(ctx)` step, but logging and delegation should remain.

```go
func (hdl *<Domain>Handler) Status(ctx context.Context, 
	req *<domain>pb.StatusRequest) (*<domain>pb.StatusResponse, error) {
	
	zlog.Debugf("[<Handler>.Status] req = %v", req)

	resp, err := hdl.service.Status(ctx, req)
	if err != nil {
		zlog.Errorf("[<Handler>.Status] failed: %v", err)
		return nil, err
	}

	return resp, nil
}
```

## 5. Testing Pattern

Create a `<domain>_test.go` file in the same package for unit tests. Tests should mock the service layer or test through it using database mocks (`pgxmock`) and mockery-generated repo mocks, keeping the structure similar to `controlplane/iam/v1/handler/iam_test.go`.

## Rules to Remember:
1. **Follow Go Style Guide:** Always apply the coding conventions from the `go-style-guide` skill.
2. The handler is a thin layer mapping gRPC requests to service methods; avoid putting heavy business logic here.
3. Always prefix debug and error logs with `[<Handler>.<MethodName>]` (`<Handler>` is the handler struct name, e.g. `IdentityAccessManagement`) for traceability.
4. Ensure any requires-authorization endpoint calls `headers.GetAuth(ctx)` and `ss.Check(...)`.
