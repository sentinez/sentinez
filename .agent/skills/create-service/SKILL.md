---
name: create-service
description: Instructions for creating a base service package following the standard sentinez pattern (e.g., iam service). Use this when tasked with creating a new gRPC service implementation.
---

# Creating a Standard Sentinez Service Package

**CRITICAL PREREQUISITE:** Before generating or implementing the service, you MUST read and apply the rules from the `go-style-guide` skill. All generated code must strictly follow the Uber Go Style Guide conventions.

**DOMAIN DEFINITION:** The protobuf definitions for the domain can be found at `api/proto/sentinez/core/<domain>/v1/<domain>.proto`. Please review it to understand the service interface, endpoints, and models.

When asked to create a new service for a given functional domain, you must follow the standard service pattern established in the `sentinez` project (such as in `github.com/sentinez/modules/iam/v1/service/service.go`).

## 1. File Structure and Package

The service should be placed in an appropriate package under `github.com/sentinez/modules/<domain>/v1/service`.
Use standard imports, especially:
- Protobuf generated code from `github.com/sentinez/sentinez/api/gen/go/sentinez/<domain>/v1` (aliased as `pb`)
- Configuration types from `github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1` (aliased as `confpb`)
- Repository interfaces from `github.com/sentinez/sentinez/github.com/sentinez/modules/<domain>/v1/repos/<model>`
- Standard error handling from `github.com/sentinez/sentinez/pkg/common/errorx`
- Postgres transaction support from `github.com/sentinez/core/storage/dbx/postgres`
- Logging from `github.com/sentinez/shared/zlog`

## 2. Interface Implementation Assertion

Ensure the service struct implements the gRPC server interface generated from Protobuf.

```go
package <domain>svc

import (
	pb "github.com/sentinez/sentinez/api/gen/go/sentinez/<domain>/v1"
	// other imports...
)

var _ pb.<Domain>ServiceServer = (*<Domain>Service)(nil)
```

## 3. Define the Struct and Constructor

The struct implements the interface and holds dependencies like `config`, `postgres.Tx` (if transactional operations are needed), and repository interfaces.

```go
type <Domain>Service struct {
	config  *confpb.Config
	tx      *postgres.Tx
	<model> <domain>repos.I<Model>
	// Other dependencies...
}

func New(config *confpb.Config,
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
	req *pb.<MethodName>Request) (*pb.<MethodName>Response, error) {
	
	// Implementation...

	return &pb.<MethodName>Response{}, nil
}
```

### Utilizing Repositories and Error Handling
When interacting with repositories, correctly handle missing rows or data validation using `errorx`.

```go
func (srv *<Domain>Service) Get<Model>(ctx context.Context, 
	req *pb.Get<Model>Request) (*pb.Get<Model>Response, error) {
	
	model, err := srv.<model>.Get(ctx, req.GetId())
	if err != nil {
		if errorx.NotRowsNotFound(err) {
			return nil, err
		}
		// Custom not found error if necessary
		return nil, errorx.StatusNotFoundF("<Model> not found: id=%s", req.GetId())
	}

	return &pb.Get<Model>Response{<Model>: model}, nil
}
```

### Transaction Support
If multiple repository writes need to be atomic, use the provided `tx` to begin and manage transactions.

```go
func (srv *<Domain>Service) Create<Model>(ctx context.Context, 
	req *pb.Create<Model>Request) (*pb.Create<Model>Response, error) {
	
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
	return &pb.Create<Model>Response{Id: model.Id}, nil
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
		Return(&pb.<Model>{Id: "test-id"}, nil)

	// 2. Setup DB Mock
	pgxMock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer pgxMock.Close(context.Background())

	tx, _ := pgxMock.Begin(ctx)
	txss := postgres.NewTXMock(tx)

	// 3. Init Service
	svc := New(nil, txss, mock<Model>Repo)

	// 4. Assert method
	req := &pb.<MethodName>Request{Id: "test-id"}
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
