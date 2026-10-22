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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	greeterpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/greeter/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakeGreeter struct {
	greeterpb.UnimplementedGreeterServiceServer
}

func (fakeGreeter) SayHello(ctx context.Context,
	req *greeterpb.SayHelloRequest) (*greeterpb.SayHelloResponse, error) {

	in, _ := metadata.FromIncomingContext(ctx)
	out, _ := metadata.FromOutgoingContext(ctx)
	_ = grpc.SetHeader(ctx, metadata.Pairs("x-out", "set"))

	msg := req.GetName() + "|" + strings.Join(in.Get("x-in"), ",") +
		"|" + strings.Join(out.Get("x-in"), ",")
	return &greeterpb.SayHelloResponse{Message: msg}, nil
}

func newGreeterClient() greeterpb.GreeterServiceClient {
	conn := NewLocalConn()
	greeterpb.RegisterGreeterServiceServer(conn, fakeGreeter{})

	return greeterpb.NewGreeterServiceClient(conn)
}

func TestLocalConnInvoke(t *testing.T) {
	ctx := metadata.AppendToOutgoingContext(t.Context(), "x-in", "v")

	var header metadata.MD
	resp, err := newGreeterClient().SayHello(ctx,
		&greeterpb.SayHelloRequest{Name: "senz"}, grpc.Header(&header))

	require.NoError(t, err)
	assert.Equal(t, "senz|v|", resp.GetMessage())
	assert.Equal(t, []string{"set"}, header.Get("x-out"))
}

func TestLocalConnValidates(t *testing.T) {
	_, err := newGreeterClient().SayHello(t.Context(),
		&greeterpb.SayHelloRequest{})

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestLocalConnUnknownMethod(t *testing.T) {
	err := NewLocalConn().Invoke(t.Context(), "/svc/Missing",
		&greeterpb.SayHelloRequest{}, &greeterpb.SayHelloResponse{})

	assert.Equal(t, codes.Unimplemented, status.Code(err))
}

func TestLocalConnGateway(t *testing.T) {
	tests := []struct {
		name     string
		giveURL  string
		wantCode int
	}{
		{
			name:     "valid",
			giveURL:  "/greeter/say?name=senz",
			wantCode: http.StatusOK,
		},
		{
			name:     "violation",
			giveURL:  "/greeter/say",
			wantCode: http.StatusBadRequest,
		},
	}

	mux := runtime.NewServeMux()
	require.NoError(t, greeterpb.RegisterGreeterServiceHandlerClient(
		t.Context(), mux, newGreeterClient()))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec,
				httptest.NewRequest(http.MethodGet, tt.giveURL, nil))

			assert.Equal(t, tt.wantCode, rec.Code, rec.Body.String())
		})
	}
}
