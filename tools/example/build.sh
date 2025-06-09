protoc --go_out=paths=source_relative:. \
    --go-grpc_out=paths=source_relative:. \
    --go-sentinez_out=. --go-sentinez_opt=paths=source_relative \
    example.proto