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

// Package recovery turns the panic of a gRPC handler into an error.
package recovery

import (
	"context"
	"runtime/debug"

	"github.com/sentinez/pkg/szlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Unary returns a server interceptor that recovers from the panic of the
// handler (and of the interceptors after it), logs it with its stack and
// fails the call with codes.Internal. The panic value is not sent to the
// caller. Put it first in the chain.
func Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {

		defer func() {
			if r := recover(); r != nil {
				szlog.Errorf("panic in %s: %v\n%s",
					info.FullMethod, r, debug.Stack())

				resp, err = nil, status.Error(codes.Internal,
					"internal error")
			}
		}()

		return handler(ctx, req)
	}
}
