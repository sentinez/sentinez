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

package grpcgateway

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sentinez/sentinez/api/interceptor/validate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	_ grpc.ClientConnInterface = (*LocalConn)(nil)
	_ grpc.ServiceRegistrar    = (*LocalConn)(nil)
)

// LocalConn dispatches unary calls straight to service handlers living in
// the same process, without a transport. Unlike the generated
// Register<Service>HandlerServer it runs unary server interceptors, so the
// gateway validates requests the same way a gRPC server does.
//
// Register services before serving; LocalConn is read-only afterwards.
type LocalConn struct {
	interceptor grpc.UnaryServerInterceptor
	methods     map[string]localMethod
}

type localMethod struct {
	server  any
	handler grpc.MethodHandler
}

// NewLocalConn returns a LocalConn that validates every request and then
// runs interceptors in order.
func NewLocalConn(interceptors ...grpc.UnaryServerInterceptor) *LocalConn {
	chain := append(
		[]grpc.UnaryServerInterceptor{validate.UnaryServerInterceptor()},
		interceptors...,
	)

	return &LocalConn{
		interceptor: chainUnary(chain),
		methods:     make(map[string]localMethod),
	}
}

// RegisterService implements grpc.ServiceRegistrar.
func (c *LocalConn) RegisterService(desc *grpc.ServiceDesc, impl any) {
	for _, m := range desc.Methods {
		name := "/" + desc.ServiceName + "/" + m.MethodName
		c.methods[name] = localMethod{server: impl, handler: m.Handler}
	}
}

// Invoke implements grpc.ClientConnInterface.
func (c *LocalConn) Invoke(ctx context.Context, method string,
	args, reply any, opts ...grpc.CallOption) error {

	m, ok := c.methods[method]
	if !ok {
		return status.Errorf(codes.Unimplemented, "unknown method %s", method)
	}

	var stream runtime.ServerTransportStream
	ctx = grpc.NewContextWithServerTransportStream(incoming(ctx), &stream)

	resp, err := m.handler(m.server, ctx, decoder(args), c.interceptor)
	setCallMetadata(&stream, opts)
	if err != nil {
		return err
	}

	return copyMessage(reply, resp)
}

// NewStream implements grpc.ClientConnInterface. Streaming is unsupported.
func (c *LocalConn) NewStream(context.Context, *grpc.StreamDesc, string,
	...grpc.CallOption) (grpc.ClientStream, error) {

	return nil, status.Error(codes.Unimplemented, "local conn: no streaming")
}

// incoming moves the caller's outgoing metadata to the incoming side, as a
// transport would, so handlers read it and do not forward it downstream.
func incoming(ctx context.Context) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return ctx
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{})
	return metadata.NewIncomingContext(ctx, md)
}

func decoder(args any) func(any) error {
	return func(v any) error {
		return copyMessage(v, args)
	}
}

func copyMessage(dst, src any) error {
	to, ok := dst.(proto.Message)
	if !ok {
		return status.Errorf(
			codes.Internal, "local conn: %T not a message", dst)
	}

	from, ok := src.(proto.Message)
	if !ok {
		return status.Errorf(
			codes.Internal, "local conn: %T not a message", src)
	}

	proto.Merge(to, from)
	return nil
}

// setCallMetadata hands the headers and trailers set by the handler to the
// grpc.Header and grpc.Trailer call options.
func setCallMetadata(stream *runtime.ServerTransportStream,
	opts []grpc.CallOption) {

	for _, opt := range opts {
		switch o := opt.(type) {
		case grpc.HeaderCallOption:
			*o.HeaderAddr = stream.Header()
		case grpc.TrailerCallOption:
			*o.TrailerAddr = stream.Trailer()
		}
	}
}

func chainUnary(
	interceptors []grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (any, error) {

		next := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor, inner := interceptors[i], next
			next = func(ctx context.Context, req any) (any, error) {
				return interceptor(ctx, req, info, inner)
			}
		}

		return next(ctx, req)
	}
}
