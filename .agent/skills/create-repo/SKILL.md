---
name: create-repo
description: Instructions for creating a database repository following the standard sentinez pattern (e.g., users.go). Use this when tasked with creating a new repository for a protobuf message model in the database layer.
---

# Creating a Standard Sentinez Database Repository

**CRITICAL PREREQUISITE:** Before generating or implementing the repository, you MUST read and apply the rules from the `go-style-guide` skill. All generated code must strictly follow the Uber Go Style Guide conventions.

**DOMAIN DEFINITION:** The protobuf definitions for the domain can be found at `api/proto/sentinez/core/<domain>/v1/<domain>.proto`. Please review it to understand the service interface, endpoints, and models.

When asked to create a new repository for a given protobuf model, you must follow the standard repository pattern established in the `sentinez` project (such as in `github.com/sentinez/modules/iam/v1/repos/users/users.go`).

## 1. File Structure and Package

The repository should be placed in an appropriate package under `repos/<model_plural>`.
Use standard imports, especially:
- `github.com/Masterminds/squirrel` for query building
- Protobuf models from `api/gen/go/sentinez/...`
- `github.com/sentinez/sentinez/internal/shared/tables` for table definitions
- `github.com/sentinez/core/storage/dbx` and `github.com/sentinez/core/storage/dbx/postgres` for database interactions
- `github.com/sentinez/core/storage/utils/table`
- `github.com/sentinez/shared/ids` for ID generation

## 2. Define the Interface

Define an interface named `I<ModelName>`, e.g., `IUser`. It must include standard CRUD methods, `WithTX` for transaction support, and `List`.

```go
type I<ModelName> interface {
	Create(ctx context.Context, model *pb.<ModelName>) (*pb.<ModelName>, error)
	Update(ctx context.Context, model *pb.<ModelName>) error
	Get(ctx context.Context, id string) (*pb.<ModelName>, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) I<ModelName>

	List(ctx context.Context, req *pb.List<ModelNamePlural>Request) (*pb.List<ModelNamePlural>Response, error)
}
```

## 3. Define the Struct and Constructor

The struct implements the interface and holds a `dbx.Database` generic over the protobuf model.

```go
type <ModelNamePlural> struct {
	storage dbx.Database[pb.<ModelName>]
}

func New(ctx context.Context, appConf *confpb.Config) (I<ModelName>, error) {
	storage, err := postgres.New[pb.<ModelName>](ctx, appConf,
		dbx.WithTable(tables.<ModelNamePlural>),
		dbx.WithColumns(dbx.ColumnM{
			pb.<ModelName>_Id:          postgres.String,
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
func buildListQuery(builder sq.SelectBuilder, req *pb.List<ModelNamePlural>Request) sq.SelectBuilder {
	for _, id := range req.GetIds() {
		builder = builder.Where(sq.Eq{pb.<ModelName>_Id: id})
	}
	// Handle other filters...
	return builder
}

func (r *<ModelNamePlural>) List(ctx context.Context, req *pb.List<ModelNamePlural>Request) (*pb.List<ModelNamePlural>Response, error) {
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

	return &pb.List<ModelNamePlural>Response{<ModelNamePlural>: models, Total: total}, nil
}
```

### Create
Generate a new ID and insert.
```go
func (r *<ModelNamePlural>) Create(ctx context.Context, model *pb.<ModelName>) (*pb.<ModelName>, error) {
	model.Id = ids.NewID(table.NewPrimaryKey(tables.<ModelNamePlural>))
	query := postgres.InsertBuilder(r.storage, postgres.M{
		pb.<ModelName>_Id: model.GetId(),
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

func (r *<ModelNamePlural>) Get(ctx context.Context, id string) (*pb.<ModelName>, error) {
	builder := r.selectQuery(nil).Where(sq.Eq{pb.<ModelName>_Id: id})
	return r.storage.Select(ctx, builder, scanOne)
}

func (r *<ModelNamePlural>) Update(ctx context.Context, model *pb.<ModelName>) error {
	query := postgres.UpdateBuilder(r.storage, model.GetId())

	if model.GetSomeField() != "" {
		query = query.Set(pb.<ModelName>_SomeField, model.GetSomeField())
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
		pb.<ModelName>_Id,
		// Other fields...
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*pb.<ModelName>, error) {
	var models []*pb.<ModelName>
	for rows.Next() {
		model, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, rows.Err()
}

func scanOne(row dbx.Row) (*pb.<ModelName>, error) {
	var (
		createdAt, updatedAt time.Time
		model                 pb.<ModelName>
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

## Rules to Remember:
1. **Follow Go Style Guide:** Always apply the coding conventions from the `go-style-guide` skill when writing Go code.
2. Ensure all model specific `pb.<ModelName>_*` constants match the column names expected by `dbx`.
3. Do not use generic names; adapt all placeholders (e.g., `<ModelName>`) to the actual domain terminology.
4. Don't forget `dbx.FieldCreatedAt` and `dbx.FieldUpdatedAt` when selecting queries and scanning mapping into `typepb.Metadata`!
