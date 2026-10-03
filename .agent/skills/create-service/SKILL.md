---
name: create-service
description: Instructions for creating a base service package following the standard sentinez pattern (e.g., iam service). Use this when tasked with creating a new gRPC service implementation.
---

# Creating a Standard Sentinez Service Package

**CRITICAL PREREQUISITE:** Before generating or implementing the service, you MUST read and apply the rules from the `go-style-guide` skill. All generated code must strictly follow the Uber Go Style Guide conventions.

**PROJECT CONVENTIONS (enforced by `.golangci.yaml`):** every Go file starts with the Apache 2.0 header (`// Copyright 2025 Duc-Hung Ho.`, copy it from any existing file); lines are at most 80 columns (tab = 4) and functions at most 35 lines (`funlen`; use `// nolint:funlen` only when unavoidable, as the iam package does). Verify with `cd staging/src/github.com/sentinez/controlplane && golangci-lint run`.

**DOMAIN DEFINITION:** The protobuf definitions for the domain live in the `api` module at `api/proto/sentinez/apps/<domain>/v1/` (`<domain>.proto` for the service and `model.proto` for models, e.g. `apps/iam/v1/iam.proto`). Review them to understand the service interface, endpoints, and models. Generated code (`*.<domain>pb.go`, `*_grpc.<domain>pb.go`, `*_senz.<domain>pb.go`, `*.<domain>pb.gw.go`, `*.<domain>pb.validate.go`) is produced by that directory's `generate.sh` / `cd api && buf generate`; never edit it by hand. Proto packages are imported as `github.com/sentinez/sentinez/api/proto/sentinez/apps/<domain>/v1` with alias `<domain>pb`. The gRPC service name comes from the proto (e.g. `IdentityAccessManagementService`), so use the real generated names rather than `<Domain>Service` literally.

When asked to create a new service for a given functional domain, you must follow the standard service pattern established in the `sentinez` project (such as in `github.com/sentinez/controlplane/iam/v1/service/service.go`).

## 1. File Structure and Package

The service should be placed in an appropriate package under `github.com/sentinez/controlplane/<domain>/v1/service`.
Use standard imports, especially:
- Protobuf generated code from `github.com/sentinez/sentinez/api/proto/sentinez/apps/<domain>/v1` (aliased as `<domain>pb`)
- Configuration types from `github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1` (aliased as `settingpb`)
- Repository interfaces from `github.com/sentinez/controlplane/<domain>/v1/repos/<model>`
- Standard error handling from `github.com/sentinez/shared/errorx`
- Postgres transaction support from `github.com/sentinez/core/storage/dbx/postgres`
- Logging from `github.com/sentinez/shared/zlog`

## 2. Interface Implementation Assertion

Ensure the service struct implements the gRPC server interface generated from Protobuf.

```go
package <domain>svc

import (
	<domain>pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/<domain>/v1"
	// other imports...
)

var _ <domain>pb.<Domain>ServiceServer = (*<Domain>Service)(nil)
```

## 3. Define the Struct and Constructor

The struct implements the interface and holds dependencies like `config`, `postgres.Tx` (if transactional operations are needed), and repository interfaces.

```go
type <Domain>Service struct {
	config  *settingpb.Config
	tx      *postgres.Tx
	<model> <domain>repos.I<Model>
	// Other dependencies...
}

func New(config *settingpb.Config,
	tx *postgres.Tx,
	<model> <domain>repos.I<Model>,
) *<Domain>Service {

	return &<Domain>Service{
		config:  config,
		tx:      tx,
		<model>: <model>,
	}
}
```

## 4. Implement RPC Methods

Each gRPC method defined in the proto file must be implemented by the service struct. Use the `errorx` package for unified gRPC error handling.

### Basic Method Signature

```go
func (srv *<Domain>Service) <MethodName>(ctx context.Context, 
	req *<domain>pb.<MethodName>Request) (*<domain>pb.<MethodName>Response, error) {
	
	// Implementation...

	return &<domain>pb.<MethodName>Response{}, nil
}
```

### Utilizing Repositories and Error Handling
When interacting with repositories, correctly handle missing rows or data validation using `errorx`.

```go
func (srv *<Domain>Service) Get<Model>(ctx context.Context, 
	req *<domain>pb.Get<Model>Request) (*<domain>pb.Get<Model>Response, error) {
	
	model, err := srv.<model>.Get(ctx, req.GetId())
	if err != nil {
		if errorx.IsNoRows(err) {
			return nil, err
		}
		// Custom not found error if necessary
		return nil, errorx.StatusNotFoundF("<Model> not found: id=%s", req.GetId())
	}

	return &<domain>pb.Get<Model>Response{<Model>: model}, nil
}
```

### Transaction Support
If multiple repository writes need to be atomic, use the provided `tx` to begin and manage transactions.

```go
func (srv *<Domain>Service) Create<Model>(ctx context.Context, 
	req *<domain>pb.Create<Model>Request) (*<domain>pb.Create<Model>Response, error) {
	
	txss, err := srv.tx.Begin(ctx)
	if err != nil {
		return nil, errorx.StatusInternalErrorF("failed to begin tx: %v", err)
	}
	// Explicit commit/rollback based on success/failure is handled

	// Call repositories bound with tx
	model, err := srv.<model>.WithTX(txss).Create(ctx, req.Get<Model>())
	if err != nil {
		_ = txss.Rollback(ctx)
		return nil, err
	}
	
	_ = txss.Commit(ctx)
	return &<domain>pb.Create<Model>Response{Id: model.Id}, nil
}
```

## 5. Testing Pattern

Create a `service_test.go` file in the same package to contain unit tests using `testify` and mocked repositories/database connections.

```go
package <domain>svc

import (
	"context"
	"testing"
	
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	// Mock repo imports...
)

func Test<MethodName>(t *testing.T) {
	ctx := context.Background()

	// 1. Initialize mocks
	mock<Model>Repo := <model>mock.NewMockI<Model>(t)
	mock<Model>Repo.On("Get", mock.Anything, "test-id").
		Return(&<domain>pb.<Model>{Id: "test-id"}, nil)

	// 2. Setup DB Mock
	pgxMock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer pgxMock.Close(context.Background())

	tx, _ := pgxMock.Begin(ctx)
	txss := postgres.NewTXMock(tx)

	// 3. Init Service
	svc := New(nil, txss, mock<Model>Repo)

	// 4. Assert method
	req := &<domain>pb.<MethodName>Request{Id: "test-id"}
	resp, err := svc.<MethodName>(ctx, req)
	
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-id", resp.<Model>.Id)
}
```

## Rules to Remember:
1. **Follow Go Style Guide:** Always apply the coding conventions from the `go-style-guide` skill.
2. Ensure you initialize models/responses safely to avoid `nil` pointer panics.
3. Handle transactions manually by checking for errors, rolling back on failure, and committing on success.
4. Pass the appropriate request fields through to underlying repos as defined.
