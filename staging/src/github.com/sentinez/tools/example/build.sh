SENTINEZ_PATH=$GOPATH/src/github.com/sentinez/sentinez
SENTINEZ_GEN_OUT=$GOPATH/src

protoc \
  -I. \
  -I"$SENTINEZ_PATH"/api/proto \
  -I"$SENTINEZ_PATH"/api/third_party/googleapis \
  -I"$SENTINEZ_PATH"/api/third_party/grpc-gateway \
  -I"$SENTINEZ_PATH"/api/third_party/protovalidate/proto/protovalidate \
    --go_out="$SENTINEZ_GEN_OUT" \
    --go-grpc_out="$SENTINEZ_GEN_OUT" \
    --go-senz_out="$SENTINEZ_GEN_OUT" \
    example.proto