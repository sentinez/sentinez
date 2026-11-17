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

package network

import (
	"net"

	"github.com/sentinez/pkg/szstore/ja4"
)

func newStdConn(conn net.Conn) *StdConn {
	return &StdConn{
		Id:   stdConnId(conn),
		Conn: conn,
	}
}

type StdConn struct {
	net.Conn
	Id string

	// pending what was read off Conn before (the client hello the
	// passthrough router peeked), served first by Read.
	pending []byte
}

// Read implements net.Conn: the pending bytes, then the conn.
func (c *StdConn) Read(b []byte) (int, error) {
	if len(c.pending) != 0 {
		n := copy(b, c.pending)
		c.pending = c.pending[n:]

		return n, nil
	}

	return c.Conn.Read(b)
}

// CloseWrite half-closes the conn when it is TCP, else closes it.
func (c *StdConn) CloseWrite() error {
	return closeWrite(c.Conn)
}

func (c *StdConn) Close() error {
	if c == nil || c.Conn == nil {
		return nil
	}

	if err := c.Conn.Close(); err != nil {
		return err
	}

	ja4.Delete(c.Id)

	return nil
}
