---
name: create-repo
description: Instructions for creating a database repository following the standard sentinez pattern (e.g., users.go). Use this when tasked with creating a new repository for a protobuf message model in the database layer.
---

# Creating a Standard Sentinez Database Repository

**CRITICAL PREREQUISITE:** Before generating or implementing the repository, you MUST read and apply the rules from the `go-style-guide` skill. All generated code must strictly follow the Uber Go Style Guide conventions.

**PROJECT CONVENTIONS (enforced by `.golangci.yaml`):** every Go file starts with the Apache 2.0 header (`// Copyright 2026 Duc-Hung Ho.`, copy it from any existing file); lines are at most 80 columns (tab = 4) and functions at most 35 lines (`funlen`; use `// nolint:funlen` only when unavoidable, as the iam package does). Verify with `cd staging/src/github.com/sentinez/controlplane && golangci-lint run`.

**DOMAIN DEFINITION:** The protobuf definitions for the domain live in the `api` module at `api/proto/sentinez/apps/<domain>/v1/` (`<domain>.proto` for the service and `model.proto` for models, e.g. `apps/iam/v1/iam.proto`). Review them to understand the service interface, endpoints, and models. Generated code (`*.<domain>pb.go`, `*_grpc.<domain>pb.go`, `*_senz.<domain>pb.go`, `*.<domain>pb.gw.go`, `*.<domain>pb.validate.go`) is produced by that directory's `generate.sh` / `cd api && buf generate`; never edit it by hand. Proto packages are imported as `github.com/sentinez/sentinez/api/proto/sentinez/apps/<domain>/v1` with alias `<domain>pb`. The gRPC service name comes from the proto (e.g. `IdentityAccessManagementService`), so use the real generated names rather than `<Domain>Service` literally.

When asked to create a new repository for a given protobuf model, you must follow the standard repository pattern established in the `sentinez` project (such as in `github.com/sentinez/controlplane/iam/v1/repos/users/users.go`).

## 1. File Structure and Package

The repository should be placed in an appropriate package under `repos/<model_plural>`.
Use standard imports, especially:
- `github.com/Masterminds/squirrel` for query building
- Protobuf models from `github.com/sentinez/sentinez/api/proto/sentinez/apps/<domain>/v1`
- `github.com/sentinez/core/common/tables` for table definitions (add the new table constant there)
- `github.com/sentinez/core/storage/dbx` and `github.com/sentinez/core/storage/dbx/postgres` for database interactions
- `github.com/sentinez/core/storage/utils/table`
- `github.com/sentinez/shared/rand` for ID generation (`rand.NewID`)

## 2. Define the Interface

Define an interface named `I<ModelName>`, e.g., `IUser`. It must include standard CRUD methods, `WithTX` for transaction support, and `List`.

```go
type I<ModelName> interface {
	Create(ctx context.Context, model *<domain>pb.<ModelName>) (*<domain>pb.<ModelName>, error)
	Update(ctx context.Context, model *<domain>pb.<ModelName>) error
	Get(ctx context.Context, id string) (*<domain>pb.<ModelName>, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) I<ModelName>

	List(ctx context.Context, req *<domain>pb.List<ModelNamePlural>Request) (*<domain>pb.List<ModelNamePlural>Response, error)
}
```

## 3. Define the Struct and Constructor

The struct implements the interface and holds a `dbx.Database` generic over the protobuf model.

```go
type <ModelNamePlural> struct {
	storage dbx.Database[<domain>pb.<ModelName>]
}

func New(ctx context.Context, appConf *settingpb.Config) (I<ModelName>, error) {
	storage, err := postgres.New[<domain>pb.<ModelName>](ctx, appConf,
		dbx.WithTable(tables.<ModelNamePlural>),
		dbx.WithColumns(dbx.ColumnM{
			<domain>pb.<ModelName>_Id:          postgres.String,
			// Map other protobuf fields to postgres types...
		}),
	)
	if err != nil {
		return nil, err
	}

	return &<ModelNamePlural>{
		storage: storage,
	}, nil
}
```

## 4. Implement Methods

### WithTX
```go
func (r *<ModelNamePlural>) WithTX(tx *postgres.TxSession) I<ModelName> {
	return &<ModelNamePlural>{
		storage: postgres.WithTx(tx, r.storage),
	}
}
```

### List and Query Building
Implement `buildListQuery` to handle filter conditions defined in the List request.
```go
func buildListQuery(builder sq.SelectBuilder, req *<domain>pb.List<ModelNamePlural>Request) sq.SelectBuilder {
	for _, id := range req.GetIds() {
		builder = builder.Where(sq.Eq{<domain>pb.<ModelName>_Id: id})
	}
	// Handle other filters...
	return builder
}

func (r *<ModelNamePlural>) List(ctx context.Context, req *<domain>pb.List<ModelNamePlural>Request) (*<domain>pb.List<ModelNamePlural>Response, error) {
	builder := postgres.SelectBuilder(r.storage, req.GetPage())
	builder = buildListQuery(builder, req)

	var total int64
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	zlog.Debug("[<model_name_plural>] query: ", query, " args: ", args)

	models, err := r.storage.CollectRows(ctx, builder, scan)
	if err != nil {
		return nil, err
	}

	if req.GetPage().GetTotal() {
		total, err = r.storage.Total(ctx)
		if err != nil {
			return nil, err
		}
	}

	return &<domain>pb.List<ModelNamePlural>Response{<ModelNamePlural>: models, Total: total}, nil
}
```

### Create
Generate a new ID and insert.
```go
func (r *<ModelNamePlural>) Create(ctx context.Context, model *<domain>pb.<ModelName>) (*<domain>pb.<ModelName>, error) {
	model.Id = rand.NewID(table.NewPrimaryKey(tables.<ModelNamePlural>))
	query := postgres.InsertBuilder(r.storage, postgres.M{
		<domain>pb.<ModelName>_Id: model.GetId(),
		// Set other fields...
	})

	if _, err := r.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return model, nil
}
```

### Get, Update, Delete
```go
func (r *<ModelNamePlural>) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

func (r *<ModelNamePlural>) Get(ctx context.Context, id string) (*<domain>pb.<ModelName>, error) {
	builder := r.selectQuery(nil).Where(sq.Eq{<domain>pb.<ModelName>_Id: id})
	return r.storage.Select(ctx, builder, scanOne)
}

func (r *<ModelNamePlural>) Update(ctx context.Context, model *<domain>pb.<ModelName>) error {
	query := postgres.UpdateBuilder(r.storage, model.GetId())

	if model.GetSomeField() != "" {
		query = query.Set(<domain>pb.<ModelName>_SomeField, model.GetSomeField())
	}
	// Check and set other modifiable fields...

	_, err := r.storage.Exec(ctx, query)
	return err
}
```

## 5. Scanning and Helpers

Implement `selectQuery`, `scan`, and `scanOne`. Be sure to attach `CreatedAt` and `UpdatedAt` from the row's base meta mapping to the protobuf `Metadata`.

```go
func (r *<ModelNamePlural>) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(r.storage, page,
		<domain>pb.<ModelName>_Id,
		// Other fields...
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*<domain>pb.<ModelName>, error) {
	var models []*<domain>pb.<ModelName>
	for rows.Next() {
		model, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, rows.Err()
}

func scanOne(row dbx.Row) (*<domain>pb.<ModelName>, error) {
	var (
		createdAt, updatedAt time.Time
		model                 <domain>pb.<ModelName>
	)

	err := row.Scan(
		&model.Id,
		// Other fields corresponding to selectQuery order
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	model.Metadata = &typepb.Metadata{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	return &model, nil
}
```

## 6. Mocks

Repo interfaces are mocked with `mockery` (testify template) into a `mock` sub-package next to the repo, e.g. `repos/users/mock/users_mocks.go` exposing `NewMockIUser(t)`. Regenerate the mock after changing the interface and never edit the generated file; service tests import it (see `iam/v1/service/service_test.go`).

## Rules to Remember:
1. **Follow Go Style Guide:** Always apply the coding conventions from the `go-style-guide` skill when writing Go code.
2. Ensure all model specific `<domain>pb.<ModelName>_*` constants match the column names expected by `dbx`.
3. Do not use generic names; adapt all placeholders (e.g., `<ModelName>`) to the actual domain terminology.
4. Don't forget `dbx.FieldCreatedAt` and `dbx.FieldUpdatedAt` when selecting queries and scanning mapping into `typepb.Metadata`!
