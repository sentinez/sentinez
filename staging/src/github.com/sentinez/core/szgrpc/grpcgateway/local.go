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
	"slices"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sentinez/core/szgrpc/recovery"
	"github.com/sentinez/core/szgrpc/validate"
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

// LocalConn dispatches calls straight to service handlers living in the
// same process, without a transport. Unlike the generated
// Register<Service>HandlerServer it runs server interceptors, so the
// gateway validates and authorizes requests the same way a gRPC server
// does. A stream runs its handler in a goroutine and passes the messages
// over channels; Close ends the open streams and waits for them.
//
// Register services before serving; LocalConn is read-only afterwards.
type LocalConn struct {
	unary   grpc.UnaryServerInterceptor
	stream  grpc.StreamServerInterceptor
	methods map[string]localMethod
	streams map[string]streamMethod

	// ctx is done on Close, which ends the open streams.
	ctx    context.Context
	cancel context.CancelFunc

	mu sync.Mutex
	wg sync.WaitGroup
}

// LocalOption configures a LocalConn.
type LocalOption interface {
	apply(*localOptions)
}

type localOptions struct {
	unary  []grpc.UnaryServerInterceptor
	stream []grpc.StreamServerInterceptor
}

type localMethod struct {
	server  any
	handler grpc.MethodHandler
}

type streamMethod struct {
	server  any
	handler grpc.StreamHandler
}

// NewLocalConn returns a LocalConn that recovers from panics, validates every
// request and then runs the interceptors of opts in order.
func NewLocalConn(opts ...LocalOption) *LocalConn {
	var o localOptions
	for _, opt := range opts {
		opt.apply(&o)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &LocalConn{
		unary: chainUnary(slices.Concat(
			[]grpc.UnaryServerInterceptor{
				recovery.Unary(), validate.UnaryServerInterceptor(),
			},
			o.unary,
		)),
		stream: chainStream(slices.Concat(
			[]grpc.StreamServerInterceptor{recovery.Stream()}, o.stream)),
		methods: make(map[string]localMethod),
		streams: make(map[string]streamMethod),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// WithUnaryInterceptors runs interceptors, in order, on the unary calls.
func WithUnaryInterceptors(
	interceptors ...grpc.UnaryServerInterceptor) LocalOption {
	return unaryOption(interceptors)
}

// WithStreamInterceptors runs interceptors, in order, on the streams.
func WithStreamInterceptors(
	interceptors ...grpc.StreamServerInterceptor) LocalOption {
	return streamOption(interceptors)
}

// RegisterService implements grpc.ServiceRegistrar.
func (c *LocalConn) RegisterService(desc *grpc.ServiceDesc, impl any) {
	for _, m := range desc.Methods {
		name := "/" + desc.ServiceName + "/" + m.MethodName
		c.methods[name] = localMethod{server: impl, handler: m.Handler}
	}

	for _, s := range desc.Streams {
		name := "/" + desc.ServiceName + "/" + s.StreamName
		c.streams[name] = streamMethod{server: impl, handler: s.Handler}
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

	resp, err := m.handler(m.server, ctx, decoder(args), c.unary)
	setCallMetadata(&stream, opts)
	if err != nil {
		return err
	}

	return copyMessage(reply, resp)
}

// NewStream implements grpc.ClientConnInterface. The handler runs until it
// returns; it is cancelled with ctx, or by Close. The call options are not
// applied: read the header and trailer from the returned stream.
func (c *LocalConn) NewStream(ctx context.Context, desc *grpc.StreamDesc,
	method string, _ ...grpc.CallOption) (grpc.ClientStream, error) {

	m, ok := c.streams[method]
	if !ok {
		return nil, status.Errorf(codes.Unimplemented,
			"unknown method %s", method)
	}

	if !c.track() {
		return nil, status.Error(codes.Unavailable, "local conn: closed")
	}

	p := newPipe(ctx, c.ctx, desc, method)
	info := &grpc.StreamServerInfo{
		FullMethod:     method,
		IsClientStream: desc.ClientStreams,
		IsServerStream: desc.ServerStreams,
	}

	go func() {
		defer c.wg.Done()
		p.finish(c.stream(m.server, serverStream{p: p}, info, m.handler))
	}()

	return clientStream{p: p}, nil
}

// Close ends the open streams and waits for their handlers to return.
// Later streams fail with codes.Unavailable; unary calls still run. A nil
// LocalConn closes as one without streams.
func (c *LocalConn) Close() error {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	c.cancel()
	c.mu.Unlock()

	c.wg.Wait()

	return nil
}

// track counts a new stream for Close to wait for, false once closed.
func (c *LocalConn) track() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ctx.Err() != nil {
		return false
	}

	c.wg.Add(1)

	return true
}

type unaryOption []grpc.UnaryServerInterceptor

func (o unaryOption) apply(opts *localOptions) {
	opts.unary = append(opts.unary, o...)
}

type streamOption []grpc.StreamServerInterceptor

func (o streamOption) apply(opts *localOptions) {
	opts.stream = append(opts.stream, o...)
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
	to, err := message(dst)
	if err != nil {
		return err
	}

	from, err := message(src)
	if err != nil {
		return err
	}

	proto.Merge(to, from)
	return nil
}

// message asserts v is a proto message.
func message(v any) (proto.Message, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, status.Errorf(
			codes.Internal, "local conn: %T not a message", v)
	}

	return msg, nil
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

func chainStream(
	interceptors []grpc.StreamServerInterceptor,
) grpc.StreamServerInterceptor {

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo,
		handler grpc.StreamHandler) error {

		next := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor, inner := interceptors[i], next
			next = func(srv any, ss grpc.ServerStream) error {
				return interceptor(srv, ss, info, inner)
			}
		}

		return next(srv, ss)
	}
}
