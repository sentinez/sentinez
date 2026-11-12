// Copyright 2025 Duc-Hung Ho.
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

package connection_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/sentinez/sentinez/api/client/connection"
	"github.com/sentinez/sentinez/api/client/connection/connectiontest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// serve starts a health server requiring client certificates of server's
// CA and returns its address.
func serve(t *testing.T, server connection.TLS) string {
	t.Helper()

	creds, err := server.ServerCredentials()
	require.NoError(t, err)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpc.NewServer(grpc.Creds(creds))
	healthpb.RegisterHealthServer(srv, health.NewServer())
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	return lis.Addr().String()
}

// check calls the health service through client.
func check(addr string, client connection.TLS) error {
	opts, err := client.DialOptions("localhost")
	if err != nil {
		return err
	}

	conn, err := connection.ConnWith(addr, opts...)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = healthpb.NewHealthClient(conn).
		Check(ctx, &healthpb.HealthCheckRequest{})

	return err
}

func TestMutualTLS(t *testing.T) {
	ca := connectiontest.NewCA(t)
	addr := serve(t, ca.Issue(t, "centraldata"))

	require.NoError(t, check(addr, ca.Issue(t, "edge-1")))
}

func TestMutualTLS_ClientOfAnotherCA(t *testing.T) {
	addr := serve(t, connectiontest.NewCA(t).Issue(t, "centraldata"))
	other := connectiontest.NewCA(t).Issue(t, "edge-1")

	// The server certificate is also unknown to the client's CA.
	assert.Error(t, check(addr, other))
}

func TestMutualTLS_ServerOfAnotherCA(t *testing.T) {
	ca := connectiontest.NewCA(t)
	addr := serve(t, connectiontest.NewCA(t).Issue(t, "centraldata"))

	assert.Error(t, check(addr, ca.Issue(t, "edge-1")))
}

func TestMutualTLS_PlaintextClient(t *testing.T) {
	ca := connectiontest.NewCA(t)
	addr := serve(t, ca.Issue(t, "centraldata"))

	assert.Error(t, check(addr, connection.TLS{}))
}

func TestTLS_NeedsAllFiles(t *testing.T) {
	_, err := connection.TLS{CAFile: "ca.pem"}.ServerCredentials()

	require.Error(t, err)
}

func TestTLS_Disabled(t *testing.T) {
	opts, err := connection.TLS{}.DialOptions("x")

	require.NoError(t, err)
	assert.Len(t, opts, 1)
	assert.False(t, connection.TLS{}.Enabled())
}
