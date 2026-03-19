---
name: create-handler
description: Instructions for creating a base handler package following the standard sentinez pattern (e.g., iam handler). Use this when tasked with creating a new gRPC handler implementation that exposes a service.
---

# Creating a Standard Sentinez Handler Package

**CRITICAL PREREQUISITE:** Before generating or implementing the handler, you MUST read and apply the rules from the `go-style-guide` skill. All generated code must strictly follow the Uber Go Style Guide conventions.

**DOMAIN DEFINITION:** The protobuf definitions for the domain can be found at `api/proto/sentinez/core/<domain>/v1/<domain>.proto`. Please review it to understand the service interface, endpoints, and models.

When asked to create a new handler for a given functional domain, you must follow the standard handler pattern established in the `sentinez` project (such as in `internal/core/iam/v1/handler/iam.go`). The handler acts as the gRPC server implementation that handles incoming requests, performs authorization checks, logs activities, and delegates business logic to the underlying service package.

## 1. File Structure and Package

The handler should be placed in an appropriate package under `internal/core/<domain>/v1/handler`. The main file should typically be named `<domain>.go`.
The package name should be `<domain>hdl`.

Use standard imports, especially:
- Protobuf generated code from `github.com/sentinez/sentinez/api/gen/go/sentinez/<domain>/v1` (aliased as `pb` or `<domain>pb`)
- The service package from `github.com/sentinez/sentinez/internal/core/<domain>/v1/service` (aliased as `<domain>svc`)
- Standard error handling from `github.com/sentinez/sentinez/pkg/common/errorx`
- Request context headers and auth from `github.com/sentinez/sentinez/pkg/common/headers`
- Logging from `github.com/sentinez/shared/zlog`

## 2. Interface Implementation Assertion

Ensure the handler struct implements the gRPC server interface generated from Protobuf.

```go
package <domain>hdl

import (
	pb "github.com/sentinez/sentinez/api/gen/go/sentinez/<domain>/v1"
	<domain>svc "github.com/sentinez/sentinez/internal/core/<domain>/v1/service"
	// other imports...
)

var _ pb.<Domain>ServiceServer = (*<Domain>Handler)(nil)
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
) pb.<Domain>ServiceServer {
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
	req *pb.<MethodName>Request) (*pb.<MethodName>Response, error) {
	
	// 1. Log the incoming request
	zlog.Debugf("[<Domain>Handler.<MethodName>] req = %v", req)

	// 2. Perform Authorization Check (only if the method requires specific permissions)
	ss, err := headers.GetAuth(ctx)
	if err != nil {
		return nil, err
	}
	
	// Assuming a permission constant is generated in pb for this method wrapper
	err = ss.Check(pb.Get<Domain>Service<MethodName>())
	if err != nil {
		return nil, err
	}

	// 3. Delegate to the service
	resp, err := hdl.service.<MethodName>(ctx, req)
	if err != nil {
		zlog.Errorf("[<Domain>Handler.<MethodName>] error: %v", err)
		return nil, err
	}

	return resp, nil
}
```

### Methods Without Authorization
For public methods (e.g., Login, Status), you might skip the `headers.GetAuth(ctx)` step, but logging and delegation should remain.

```go
func (hdl *<Domain>Handler) Status(ctx context.Context, 
	req *pb.StatusRequest) (*pb.StatusResponse, error) {
	
	zlog.Debugf("[<Domain>Handler.Status] req = %v", req)

	resp, err := hdl.service.Status(ctx, req)
	if err != nil {
		zlog.Errorf("[<Domain>Handler.Status] failed: %v", err)
		return nil, err
	}

	return resp, nil
}
```

## 5. Testing Pattern

Create a `<domain>_test.go` file in the same package for unit tests. Tests should mock the service layer or test through it using database mocks (`pgxmock`), keeping the structure similar to IAM testing.

## Rules to Remember:
1. **Follow Go Style Guide:** Always apply the coding conventions from the `go-style-guide` skill.
2. The handler is a thin layer mapping gRPC requests to service methods; avoid putting heavy business logic here.
3. Always prefix debug and error logs with `[<Domain>Handler.<MethodName>]` for traceability.
4. Ensure any requires-authorization endpoint calls `headers.GetAuth(ctx)` and `ss.Check(...)`.
