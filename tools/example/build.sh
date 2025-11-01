protoc --go_out=paths=source_relative:. \
    --go-grpc_out=paths=source_relative:. \
    --go-senz-msg_out=. --go-senz-msg_opt=paths=source_relative \
    example.proto