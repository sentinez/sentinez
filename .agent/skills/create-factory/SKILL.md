---
name: create-factory
description: Instructions for creating a base factory package following the standard sentinez pattern (e.g., iam factory). Use this when tasked with creating a new DI factory package that wires up repositories, services, and handlers.
---

# Creating a Standard Sentinez Factory Package

**CRITICAL PREREQUISITE:** Before generating or implementing the factory, you MUST read and apply the rules from the `go-style-guide` skill. All generated code must strictly follow the Uber Go Style Guide conventions.

**DOMAIN DEFINITION:** The protobuf definitions for the domain can be found at `api/proto/sentinez/core/<domain>/v1/<domain>.proto`. Please review it to understand the service interface, endpoints, and models.

When asked to create a new factory for a given functional domain, you must follow the standard Dependency Injection (DI) factory pattern established in the `sentinez` project (such as in `internal/core/iam/v1/factory/factory.go`). This factory acts as a composition root that wires up repositories, third-party clients, the business service, and the gRPC handler.

## 1. File Structure and Package

The factory should be placed in an appropriate package under `internal/core/<domain>/v1/factory`. The main file should typically be named `factory.go`.
The package name should be `<domain>fac`.

Use standard imports, especially:
- Context from `context`
- Protobuf generated code from `github.com/sentinez/sentinez/api/gen/go/sentinez/<domain>/v1` (aliased as `pb` or `<domain>pb`)
- Configuration types from `github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1` (aliased as `confpb`)
- Repository interfaces from `github.com/sentinez/sentinez/internal/core/<domain>/v1/repos/<model>`
- The service package from `github.com/sentinez/sentinez/internal/core/<domain>/v1/service` (aliased as `<domain>svc`)
- The handler package from `github.com/sentinez/sentinez/internal/core/<domain>/v1/handler` (aliased as `<domain>hdl`)
- Database context/transactions from `github.com/sentinez/sentinez/pkg/storage/dbx/postgres`
- Logging from `github.com/sentinez/shared/zlog`

## 2. Initialize the Service: `NewDefaultService`

Define `NewDefaultService` to initialize repositories, database transactions, and inject them into the underlying service.

```go
func NewDefaultService(
	ctx context.Context, appConf *confpb.Config) *<domain>svc.<Domain>Service {
	
	// 1. Initialize all necessary repositories
	<model>repos, err := <model>repo.New(ctx, appConf)
	if err != nil {
		zlog.Errorf("<domain>factory: init <model> repo err=%v", err)
	}

	// Wait for other initializations if necessary...

	// 2. Initialize transactional database dependency if needed
	tx := postgres.NewTX(appConf)

	// 3. Inject them into the Service and return
	return <domain>svc.New(appConf, tx, <model>repos)
}
```

## 3. Initialize the Handler: `NewDefaultHandler`

Define `NewDefaultHandler` that returns the gRPC server interface (`pb.<Domain>ServiceServer`). It constructs the service using `NewDefaultService` and initializes any external clients.

```go
func NewDefaultHandler(ctx context.Context, appConf *confpb.Config,
) pb.<Domain>ServiceServer {

	// 1. Initialize the Service using the factory method above
	service := NewDefaultService(ctx, appConf)

	// 2. Initialize any external gRPC Clients via their respective factories
	// greeterCli, err := client.NewLocalGreeter(greeterfac.NewDefaultHandler(appConf))
	// if err != nil {
	// 	zlog.Errorf("<domain>factory: new greeter client err=%v", err)
	// }

	// 3. Mount the service and external clients to the Handler
	return <domain>hdl.New(service /*, greeterCli */)
}
```

## Rules to Remember:
1. **Follow Go Style Guide:** Always apply the coding conventions from the `go-style-guide` skill.
2. Ensure you initialize dependencies safely; prefer logging errors (`zlog.Errorf`) and returning partially instantiated services/handlers than outright panicking unless critically required.
3. Keep the factory focused on wiring dependencies; do NOT put business logic inside `factory.go`.
