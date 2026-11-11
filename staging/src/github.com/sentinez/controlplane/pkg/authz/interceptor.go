// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package authz authenticates and authorizes the calls to the control plane.
package authz

import (
	"context"
	"strings"

	"github.com/sentinez/controlplane/pkg/headers"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// Unary returns a server interceptor that denies a call by default:
//   - a method marked public is open, and sees the caller when it sent a
//     valid token;
//   - any other method, including one without an x_method option, needs a
//     valid token whose control plane the method allows.
//
// The caller is then available from [headers.GetAuth].
func Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (any, error) {

		ctx, err := authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func authorize(ctx context.Context,
	fullMethod string) (context.Context, error) {

	method := methodOf(fullMethod)

	auth, err := headers.GetAuth(ctx)
	if method.GetPublic() {
		if err == nil {
			ctx = headers.WithAuth(ctx, auth)
		}

		return ctx, nil
	}

	if err != nil {
		return ctx, err
	}

	if err = auth.Check(method); err != nil {
		return ctx, err
	}

	return headers.WithAuth(ctx, auth), nil
}

// methodOf returns the x_method option of the gRPC method fullMethod
// ("/pkg.Service/Method"), or nil when it has none or is unknown.
func methodOf(fullMethod string) *typepb.XMethod {
	name := strings.ReplaceAll(strings.TrimPrefix(fullMethod, "/"), "/", ".")

	desc, err := protoregistry.GlobalFiles.
		FindDescriptorByName(protoreflect.FullName(name))
	if err != nil {
		return nil
	}

	md, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return nil
	}

	ext := proto.GetExtension(md.Options(), typepb.E_XMethod)
	method, _ := ext.(*typepb.XMethod)

	return method
}

// ServerOption returns the gRPC server option that installs [Unary].
func ServerOption() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(Unary())
}
