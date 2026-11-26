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

// Package validate provides gRPC interceptors that check every request
// message against its protovalidate (buf.validate) rules.
package validate

import (
	"context"
	"errors"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// UnaryServerInterceptor rejects requests that break their buf.validate
// rules before they reach the handler.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if err := Message(req); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// UnaryClientInterceptor rejects requests that break their buf.validate
// rules before they are sent.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if err := Message(req); err != nil {
			return err
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// Message validates req and reports rule violations as an InvalidArgument
// status carrying the buf.validate.Violations detail. Values that are not
// proto messages are accepted.
func Message(req any) error {
	msg, ok := req.(proto.Message)
	if !ok {
		return nil
	}

	err := protovalidate.Validate(msg)
	if err == nil {
		return nil
	}

	var verr *protovalidate.ValidationError
	if !errors.As(err, &verr) {
		return status.Error(codes.Internal, err.Error())
	}

	st := status.New(codes.InvalidArgument, verr.Error())
	if detailed, derr := st.WithDetails(verr.ToProto()); derr == nil {
		st = detailed
	}

	return st.Err()
}
