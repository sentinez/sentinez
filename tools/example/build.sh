protoc --go_out=paths=source_relative:. \
    --go-grpc_out=paths=source_relative:. \
    --go-sntzmodels_out=. --go-sntzmodels_opt=paths=source_relative \
    example.proto