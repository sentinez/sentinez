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

	"github.com/sentinez/shared/zlog"
)

func StdListen(addr string, opts ...NetworkOption) (*StdListener, error) {
	listener := &StdListener{}

	for _, opt := range opts {
		opt(&listener.opt)
	}

	if listener.opt.network == "" {
		listener.opt.network = "tcp"
	}

	lis, err := net.Listen(listener.opt.network, addr)
	if err != nil {
		return nil, err
	}

	listener.Listener = lis

	return listener, nil
}

type StdListener struct {
	net.Listener
	opt Option
}

func (l *StdListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	connWrapper := newStdConn(conn)

	if l.opt.onStdAccept != nil {
		if err := l.opt.onStdAccept(connWrapper); err != nil {
			return nil, err
		}
	}

	zlog.Debugf("network: new conn: %s", connWrapper.Id)

	return connWrapper, nil
}
