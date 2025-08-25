//go:build tools
// +build tools

// Package tools contains tool dependencies.
//
//	go mod tidy
//	go install ./...
package tools

import (
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/codesenberg/bombardier"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	_ "github.com/sentinez/sentinez/tools/cmd/protoc-gen-go-sntzmodels"
	_ "github.com/sentinez/sentinez/tools/cmd/protoc-gen-go-sntzopts"
	_ "github.com/sentinez/sentinez/tools/cmd/ruleparser-sentinez"
	_ "github.com/vektra/mockery/v2"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
