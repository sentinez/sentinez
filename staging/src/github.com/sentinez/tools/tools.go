//go:build tools
// +build tools

// Package tools contains tool dependencies.
//
//	go mod tidy
//	go install ./...
package tools

import (
	_ "github.com/a-h/templ/cmd/templ"
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/cilium/ebpf/cmd/bpf2go"
	_ "github.com/codesenberg/bombardier"
	_ "github.com/favadi/protoc-go-inject-tag"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	_ "github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto"
	_ "github.com/sentinez/tools/cmd/protoc-gen-go-senz"
	_ "github.com/sentinez/tools/cmd/ruleparser-sentinez"
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
	_ "github.com/vektra/mockery/v3"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
