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

package validate

import (
	"context"
	"testing"

	validatepb "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	greeterpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/greeter/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMessage(t *testing.T) {
	tests := []struct {
		name     string
		give     any
		wantCode codes.Code
	}{
		{
			name:     "valid",
			give:     &greeterpb.SayHelloRequest{Name: "sentinez"},
			wantCode: codes.OK,
		},
		{
			name:     "violation",
			give:     &greeterpb.SayHelloRequest{},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "not a proto message",
			give:     "raw",
			wantCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := status.Code(Message(tt.give))
			if got != tt.wantCode {
				t.Errorf("Message() code = %v, want %v", got, tt.wantCode)
			}
		})
	}
}

func TestMessageViolationDetails(t *testing.T) {
	st := status.Convert(Message(&greeterpb.SayHelloRequest{}))

	details := st.Details()
	if len(details) != 1 {
		t.Fatalf("details = %d, want 1", len(details))
	}

	violations, ok := details[0].(*validatepb.Violations)
	if !ok {
		t.Fatalf("detail type = %T, want *Violations", details[0])
	}
	if len(violations.GetViolations()) != 1 {
		t.Errorf("violations = %d, want 1", len(violations.GetViolations()))
	}
}

func TestUnaryServerInterceptor(t *testing.T) {
	var called bool
	handler := func(context.Context, any) (any, error) {
		called = true
		return &greeterpb.SayHelloResponse{}, nil
	}

	_, err := UnaryServerInterceptor()(t.Context(),
		&greeterpb.SayHelloRequest{}, &grpc.UnaryServerInfo{}, handler)

	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", status.Code(err))
	}
	if called {
		t.Error("handler called for an invalid request")
	}
}

func TestUnaryClientInterceptor(t *testing.T) {
	var called bool
	invoker := func(context.Context, string, any, any,
		*grpc.ClientConn, ...grpc.CallOption) error {
		called = true
		return nil
	}

	err := UnaryClientInterceptor()(t.Context(), "/svc/Method",
		&greeterpb.SayHelloRequest{}, &greeterpb.SayHelloResponse{},
		nil, invoker)

	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", status.Code(err))
	}
	if called {
		t.Error("invoker called for an invalid request")
	}
}
