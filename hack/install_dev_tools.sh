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

# sql generate tools
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# grpc gateway & protocol buffer
go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc

# mockup test
go install github.com/vektra/mockery/v2@latest

# proto lint
go install github.com/bufbuild/buf/cmd/buf@v1.48.0

# Go 1.16+
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Wire dependency injection
go install github.com/google/wire/cmd/wire@latest

go install github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto@latest