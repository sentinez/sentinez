// Copyright 2026 Sentinéz Labs.
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
//go:build linux

package network

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func tcpPair(t *testing.T) (client, server net.Conn) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = lis.Close() })

	client, err = net.Dial("tcp", lis.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	server, err = lis.Accept()
	require.NoError(t, err)
	t.Cleanup(func() { _ = server.Close() })

	return client, server
}

func pipeConn(t *testing.T) net.Conn {
	t.Helper()

	a, b := net.Pipe()
	t.Cleanup(func() { _ = a.Close(); _ = b.Close() })

	return a
}

func TestSockCookie(t *testing.T) {
	client, server := tcpPair(t)

	first, err := sockCookie(server)
	require.NoError(t, err)
	require.NotZero(t, first)

	again, err := sockCookie(server)
	require.NoError(t, err)
	require.Equal(t, first, again)

	other, err := sockCookie(client)
	require.NoError(t, err)
	require.NotEqual(t, first, other)
}

func TestSockCookieNoFd(t *testing.T) {
	a := pipeConn(t)

	_, err := sockCookie(a)
	require.ErrorIs(t, err, errNoSyscallConn)
}

func TestStdConnId(t *testing.T) {
	_, server := tcpPair(t)

	id := stdConnId(server)
	require.Len(t, id, 16)
	require.Equal(t, id, stdConnId(server))
}

func TestStdConnIdFallback(t *testing.T) {
	a := pipeConn(t)

	id := stdConnId(a)
	require.Len(t, id, 16)
	require.NotEqual(t, id, stdConnId(a))
}
