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
	"crypto/rand"
	"encoding/hex"
	"net"

	"github.com/sentinez/shared/store/ja4"
)

func newConn(conn net.Conn) *Conn {
	id := make([]byte, 8)
	_, _ = rand.Read(id)

	return &Conn{
		Id:   hex.EncodeToString(id),
		Conn: conn,
	}
}

type Conn struct {
	net.Conn
	Id string
}

func (c *Conn) Close() error {
	if c == nil || c.Conn == nil {
		return nil
	}

	if err := c.Conn.Close(); err != nil {
		return err
	}

	ja4.Delete(c.Id)

	return nil
}
