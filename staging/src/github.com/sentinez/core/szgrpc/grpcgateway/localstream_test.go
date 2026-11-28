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
	"errors"
	"io"
	"strings"
	"testing"

	greeterpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/greeter/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const _echoMethod = "/test.Echo/Echo"

// echo is a bidi stream: it sends the header, then echoes each request
// with the x-in metadata of the caller, then sets the trailer.
func echo(_ any, ss grpc.ServerStream) error {
	md, _ := metadata.FromIncomingContext(ss.Context())
	if err := ss.SendHeader(metadata.Pairs("x-head", "h")); err != nil {
		return err
	}

	for {
		var req greeterpb.SayHelloRequest
		err := ss.RecvMsg(&req)
		if errors.Is(err, io.EOF) {
			ss.SetTrailer(metadata.Pairs("x-tail", "t"))
			return nil
		}
		if err != nil {
			return err
		}

		msg := req.GetName() + "|" + strings.Join(md.Get("x-in"), ",")
		if err := ss.SendMsg(
			&greeterpb.SayHelloResponse{Message: msg}); err != nil {
			return err
		}
	}
}

// fail ends the stream with codes.FailedPrecondition before any header.
func fail(any, grpc.ServerStream) error {
	return status.Error(codes.FailedPrecondition, "fail")
}

// wait blocks until the stream is cancelled.
func wait(_ any, ss grpc.ServerStream) error {
	<-ss.Context().Done()
	return ss.Context().Err()
}

var _echoDesc = grpc.StreamDesc{
	StreamName: "Echo", ServerStreams: true, ClientStreams: true,
}

// newStreamConn returns a LocalConn serving handler at _echoMethod.
func newStreamConn(t *testing.T, handler grpc.StreamHandler,
	opts ...LocalOption) *LocalConn {
	t.Helper()

	desc := _echoDesc
	desc.Handler = handler

	conn := NewLocalConn(opts...)
	conn.RegisterService(&grpc.ServiceDesc{
		ServiceName: "test.Echo",
		HandlerType: (*any)(nil),
		Streams:     []grpc.StreamDesc{desc},
	}, struct{}{})
	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

func TestLocalConnStreamEcho(t *testing.T) {
	ctx := metadata.AppendToOutgoingContext(t.Context(), "x-in", "v")
	cs, err := newStreamConn(t, echo).
		NewStream(ctx, &_echoDesc, _echoMethod)
	require.NoError(t, err)

	header, err := cs.Header()
	require.NoError(t, err)
	assert.Equal(t, []string{"h"}, header.Get("x-head"))

	for _, name := range []string{"a", "b"} {
		require.NoError(t,
			cs.SendMsg(&greeterpb.SayHelloRequest{Name: name}))

		var resp greeterpb.SayHelloResponse
		require.NoError(t, cs.RecvMsg(&resp))
		assert.Equal(t, name+"|v", resp.GetMessage())
	}

	require.NoError(t, cs.CloseSend())
	assert.ErrorIs(t, cs.RecvMsg(&greeterpb.SayHelloResponse{}), io.EOF)
	assert.Equal(t, []string{"t"}, cs.Trailer().Get("x-tail"))
}

func TestLocalConnStreamValidates(t *testing.T) {
	cs, err := newStreamConn(t, echo).
		NewStream(t.Context(), &_echoDesc, _echoMethod)
	require.NoError(t, err)

	require.NoError(t, cs.SendMsg(&greeterpb.SayHelloRequest{}))

	err = cs.RecvMsg(&greeterpb.SayHelloResponse{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestLocalConnStreamNoHeader ends without a header: Header returns nil
// and RecvMsg the status, as on a transport.
func TestLocalConnStreamNoHeader(t *testing.T) {
	desc := grpc.StreamDesc{StreamName: "Echo", ServerStreams: true}
	cs, err := newStreamConn(t, fail).
		NewStream(t.Context(), &desc, _echoMethod)
	require.NoError(t, err)

	header, err := cs.Header()
	require.NoError(t, err)
	assert.Nil(t, header)

	// a call of one message gets its status from RecvMsg only
	require.NoError(t, cs.SendMsg(&greeterpb.SayHelloRequest{Name: "a"}))

	err = cs.RecvMsg(&greeterpb.SayHelloResponse{})
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestLocalConnStreamInterceptor(t *testing.T) {
	deny := func(any, grpc.ServerStream, *grpc.StreamServerInfo,
		grpc.StreamHandler) error {
		return status.Error(codes.Unauthenticated, "deny")
	}

	conn := newStreamConn(t, echo, WithStreamInterceptors(deny))
	cs, err := conn.NewStream(t.Context(), &_echoDesc, _echoMethod)
	require.NoError(t, err)

	err = cs.RecvMsg(&greeterpb.SayHelloResponse{})
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestLocalConnStreamCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cs, err := newStreamConn(t, wait).NewStream(ctx, &_echoDesc, _echoMethod)
	require.NoError(t, err)

	cancel()

	err = cs.RecvMsg(&greeterpb.SayHelloResponse{})
	assert.Equal(t, codes.Canceled, status.Code(err))
}

// TestLocalConnClose ends the open streams, and refuses new ones.
func TestLocalConnClose(t *testing.T) {
	conn := newStreamConn(t, wait)
	cs, err := conn.NewStream(t.Context(), &_echoDesc, _echoMethod)
	require.NoError(t, err)

	require.NoError(t, conn.Close())

	err = cs.RecvMsg(&greeterpb.SayHelloResponse{})
	assert.ErrorIs(t, err, context.Canceled)

	_, err = conn.NewStream(t.Context(), &_echoDesc, _echoMethod)
	assert.Equal(t, codes.Unavailable, status.Code(err))
}

func TestLocalConnStreamUnknownMethod(t *testing.T) {
	_, err := NewLocalConn().
		NewStream(t.Context(), &_echoDesc, "/svc/Missing")

	assert.Equal(t, codes.Unimplemented, status.Code(err))
}
