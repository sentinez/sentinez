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

	"github.com/sentinez/pkg/szlog"
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
	if listener.opt.passthrough != nil {
		listener.sni = newSNIRouter(lis, listener.opt.passthrough)
	}

	return listener, nil
}

type StdListener struct {
	net.Listener
	opt Option
	// sni routes the connections when there is a passthrough, else nil.
	sni *sniRouter
}

func (l *StdListener) Accept() (net.Conn, error) {
	connWrapper, err := l.accept()
	if err != nil {
		return nil, err
	}

	if l.opt.onStdAccept != nil {
		if err := l.opt.onStdAccept(connWrapper); err != nil {
			return nil, err
		}
	}

	szlog.Debugf("network: new conn: %s", connWrapper.Id)

	return connWrapper, nil
}

// Close implements net.Listener: it also stops the passthrough
// connections.
func (l *StdListener) Close() error {
	err := l.Listener.Close()
	if l.sni != nil {
		l.sni.close()
	}

	return err
}

func (l *StdListener) accept() (*StdConn, error) {
	if l.sni != nil {
		return l.sni.accept()
	}

	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return newStdConn(conn), nil
}
