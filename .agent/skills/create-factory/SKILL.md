---
name: create-factory
description: Instructions for creating a base factory package following the standard sentinez pattern (e.g., iam factory). Use this when tasked with creating a new DI factory package that wires up repositories, services, and handlers.
---

# Creating a Standard Sentinez Factory Package

**CRITICAL PREREQUISITE:** Before generating or implementing the factory, you MUST read and apply the rules from the `go-style-guide` skill. All generated code must strictly follow the Uber Go Style Guide conventions.

**PROJECT CONVENTIONS (enforced by `.golangci.yaml`):** every Go file starts with the Apache 2.0 header (`// Copyright 2025 Duc-Hung Ho.`, copy it from any existing file); lines are at most 80 columns (tab = 4) and functions at most 35 lines (`funlen`; use `// nolint:funlen` only when unavoidable, as the iam package does). Verify with `cd staging/src/github.com/sentinez/controlplane && golangci-lint run`.

**DOMAIN DEFINITION:** The protobuf definitions for the domain live in the `api` module at `api/proto/sentinez/apps/<domain>/v1/` (`<domain>.proto` for the service and `model.proto` for models, e.g. `apps/iam/v1/iam.proto`). Review them to understand the service interface, endpoints, and models. Generated code (`*.<domain>pb.go`, `*_grpc.<domain>pb.go`, `*_senz.<domain>pb.go`, `*.<domain>pb.gw.go`, `*.<domain>pb.validate.go`) is produced by that directory's `generate.sh` / `cd api && buf generate`; never edit it by hand. Proto packages are imported as `github.com/sentinez/sentinez/api/proto/sentinez/apps/<domain>/v1` with alias `<domain>pb`. The gRPC service name comes from the proto (e.g. `IdentityAccessManagementService`), so use the real generated names rather than `<Domain>Service` literally.

When asked to create a new factory for a given functional domain, you must follow the standard Dependency Injection (DI) factory pattern established in the `sentinez` project (such as in `github.com/sentinez/controlplane/iam/v1/factory/factory.go`). This factory acts as a composition root that wires up repositories, third-party clients, the business service, and the gRPC handler.

## 1. File Structure and Package

The factory should be placed in an appropriate package under `github.com/sentinez/controlplane/<domain>/v1/factory`. The main file should typically be named `factory.go`.
The package name should be `<domain>fac`.

Use standard imports, especially:
- Context from `context`
- Protobuf generated code from `github.com/sentinez/sentinez/api/proto/sentinez/apps/<domain>/v1` (aliased as `<domain>pb`)
- Configuration types from `github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1` (aliased as `settingpb`)
- Repository interfaces from `github.com/sentinez/controlplane/<domain>/v1/repos/<model>`
- The service package from `github.com/sentinez/controlplane/<domain>/v1/service` (aliased as `<domain>svc`)
- The handler package from `github.com/sentinez/controlplane/<domain>/v1/handler` (aliased as `<domain>hdl`)
- Database context/transactions from `github.com/sentinez/core/storage/dbx/postgres`
- Logging from `github.com/sentinez/shared/zlog`

## 2. Initialize the Service: `NewDefaultService`

Define `NewDefaultService` (mirroring `iamfac.NewDefaultService`) to initialize repositories, database transactions, and inject them into the underlying service.

```go
func NewDefaultService(
	ctx context.Context, appConf *settingpb.Config) *<domain>svc.<Domain>Service {
	
	// 1. Initialize all necessary repositories
	<model>repos, err := <model>repo.New(ctx, appConf)
	if err != nil {
		zlog.Errorf("<domain>fac: init <model> repo err=%v", err)
	}

	// Wait for other initializations if necessary...

	// 2. Initialize transactional database dependency if needed
	tx := postgres.NewTX(appConf)

	// 3. Inject them into the Service and return
	return <domain>svc.New(appConf, tx, <model>repos)
}
```

## 3. Initialize the Handler: `NewDefaultHandler`

Define `NewDefaultHandler` that returns the gRPC server interface (`<domain>pb.<Domain>ServiceServer`). It constructs the service using `NewDefaultService` and initializes any external clients.

```go
func NewDefaultHandler(ctx context.Context, appConf *settingpb.Config,
) <domain>pb.<Domain>ServiceServer {

	// 1. Initialize the Service using the factory method above
	service := NewDefaultService(ctx, appConf)

	// 2. Initialize any external gRPC Clients via their respective factories
	// In-process client over bufconn; `client` is
	// github.com/sentinez/sentinez/api/client (NewLocalGreeter, NewLocalIAM...).
	// greeterCli, err := client.NewLocalGreeter(
	// 	greeterfac.NewDefaultHandler(appConf),
	// )
	// if err != nil {
	// 	zlog.Errorf("<domain>fac: new greeter client err=%v", err)
	// }

	// 3. Mount the service and external clients to the Handler
	return <domain>hdl.New(service /*, greeterCli */)
}
```

## Rules to Remember:
1. **Follow Go Style Guide:** Always apply the coding conventions from the `go-style-guide` skill.
2. Ensure you initialize dependencies safely; prefer logging errors (`zlog.Errorf`) and returning partially instantiated services/handlers than outright panicking unless critically required.
3. Keep the factory focused on wiring dependencies; do NOT put business logic inside `factory.go`.
