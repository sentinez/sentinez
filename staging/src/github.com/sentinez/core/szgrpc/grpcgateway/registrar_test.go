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
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sentinez/core/szgrpc/client/connection"
	"github.com/sentinez/core/szgrpc/client/discovery"
	greeterpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/greeter/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// staticInstancer returns addrs, or err when set.
type staticInstancer struct {
	addrs []string
	err   error
}

func (s staticInstancer) Instances(string) ([]string, error) {
	return s.addrs, s.err
}

// serveGreeter serves fakeGreeter on a loopback port, returned.
func serveGreeter(t *testing.T) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	gs := grpc.NewServer()
	greeterpb.RegisterGreeterServiceServer(gs, fakeGreeter{})

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = gs.Serve(lis)
	}()
	t.Cleanup(func() {
		gs.Stop()
		<-done
	})

	return lis.Addr().String()
}

// endpointMux returns a gateway of the greeter nodes of in.
func endpointMux(t *testing.T, in discovery.Instancer) *runtime.ServeMux {
	t.Helper()

	opts, err := dialOptions(connection.TLS{}, "",
		discovery.NewBuilder(in, time.Hour))
	require.NoError(t, err)

	mux := runtime.NewServeMux()
	require.NoError(t, greeterpb.RegisterGreeterServiceHandlerFromEndpoint(
		t.Context(), mux,
		discovery.Target(greeterpb.GetMetaGreeterServiceKey()), opts))

	return mux
}

func TestEndpointCallsNode(t *testing.T) {
	mux := endpointMux(t,
		staticInstancer{addrs: []string{serveGreeter(t)}})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec,
		httptest.NewRequest(http.MethodGet, "/greeter/say?name=sz", nil))

	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

func TestEndpointWithoutNode(t *testing.T) {
	mux := endpointMux(t,
		staticInstancer{err: errors.New("no service available")})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec,
		httptest.NewRequest(http.MethodGet, "/greeter/say?name=sz", nil))

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code,
		rec.Body.String())
}
