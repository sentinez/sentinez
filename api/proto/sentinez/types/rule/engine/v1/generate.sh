#!/bin/bash

# Copyright 2025 Duc-Hung Ho.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

CURRENT_DIR=$(pwd)
GOPATH_DIR=${GOPATH}

if [ -z "$GOPATH_DIR" ]; then
  echo "GOPATH is not set. Please set it before running this script."
  exit 1
fi

if [[ "$CURRENT_DIR" != *"$GOPATH_DIR"* ]]; then
  echo "Current directory ($CURRENT_DIR) does not contain GOPATH ($GOPATH_DIR)."
  exit 1
fi

SENTINEZ_PATH=$GOPATH/src/github.com/sentinez/sentinez
SENTINEZ_GEN_OUT=$GOPATH/src
SENTINEZ_OPENAPI_OUT=$SENTINEZ_PATH/api/docs/v1

protoc \
  -I"$SENTINEZ_PATH"/api/proto \
  -I"$SENTINEZ_PATH"/api/third_party/googleapis \
  -I"$SENTINEZ_PATH"/api/third_party/grpc-gateway \
  -I"$SENTINEZ_PATH"/api/third_party/protovalidate/proto/protovalidate \
  --grpc-gateway_out="$SENTINEZ_GEN_OUT" \
  --go_out="$SENTINEZ_GEN_OUT" \
  --go-grpc_out="$SENTINEZ_GEN_OUT" \
  --validate_out="lang=go,paths=:$SENTINEZ_GEN_OUT" \
  --go-vtproto_out="$SENTINEZ_GEN_OUT" \
  --go-vtproto_opt=features=marshal+unmarshal+size \
  --go-senz-msg_out="$SENTINEZ_GEN_OUT" \
  "$(pwd)"/*.proto || exit 1

protoc-go-inject-tag -input="$SENTINEZ_GEN_OUT"/github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1/*.pb.go