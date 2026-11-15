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
	"github.com/sentinez/pkg/szerror"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
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

	ctx = headers.WithAuth(ctx, auth)

	return ctx, bindCertificate(ctx)
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

// Stream is [Unary] for the streaming calls: the same rules, applied once
// when the stream opens, and the caller is on the context of the stream.
func Stream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo,
		handler grpc.StreamHandler) error {

		ctx, err := authorize(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}

		return handler(srv, &authStream{ServerStream: ss, ctx: ctx})
	}
}

// authStream is a stream whose context carries the caller.
type authStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *authStream) Context() context.Context { return s.ctx }

// ServerOptions returns the gRPC server options that install [Unary] and
// [Stream].
func ServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(Unary()),
		grpc.ChainStreamInterceptor(Stream()),
	}
}

// bindCertificate fails when a service token travels on a connection whose
// verified client certificate is of another service: a token stolen from
// one service is of no use with the certificate of another. The check is
// skipped on a connection without a client certificate (plaintext).
func bindCertificate(ctx context.Context) error {
	auth, err := headers.GetAuth(ctx)
	if err != nil || !auth.IsService() {
		return nil
	}

	cn, ok := peerCommonName(ctx)
	if !ok || cn == auth.GetName() {
		return nil
	}

	return szerror.StatusForbiddenF(
		"service token of %q on the certificate of %q", auth.GetName(), cn)
}

// peerCommonName the common name of the verified client certificate.
func peerCommonName(ctx context.Context) (string, bool) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", false
	}

	info, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok || len(info.State.VerifiedChains) == 0 ||
		len(info.State.VerifiedChains[0]) == 0 {
		return "", false
	}

	return info.State.VerifiedChains[0][0].Subject.CommonName, true
}
