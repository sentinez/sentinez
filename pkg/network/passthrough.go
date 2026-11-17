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

package network

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/sentinez/pkg/szlog"
)

const (
	// _helloTimeout bounds the wait for the client hello of a connection.
	_helloTimeout = 10 * time.Second
	// _passthroughDialTimeout bounds the dial of an upstream.
	_passthroughDialTimeout = 10 * time.Second
	// _acceptBackoffMax caps the wait after a failed accept.
	_acceptBackoffMax = time.Second
)

// PassthroughFunc resolves the server name (SNI) of a client hello to the
// addresses ("host:port") the edge forwards the connection to untouched,
// in order of preference; false when the edge terminates its TLS.
type PassthroughFunc func(serverName string) ([]string, bool)

// errHelloRead stops the handshake run only to read the client hello.
var errHelloRead = errors.New("client hello read")

// sniRouter accepts the connections of a listener and reads their client
// hello: those of a passthrough name are piped to their upstream, the
// others are handed to Accept, the hello replayed for the TLS server.
type sniRouter struct {
	lis     net.Listener
	resolve PassthroughFunc
	conns   chan *StdConn

	ctx     context.Context
	cancel  context.CancelFunc
	stopped chan struct{}

	wg     sync.WaitGroup
	mu     sync.Mutex
	active map[net.Conn]struct{}
}

func newSNIRouter(lis net.Listener, resolve PassthroughFunc) *sniRouter {
	ctx, cancel := context.WithCancel(context.Background())
	r := &sniRouter{
		lis:     lis,
		resolve: resolve,
		conns:   make(chan *StdConn),
		ctx:     ctx,
		cancel:  cancel,
		stopped: make(chan struct{}),
		active:  make(map[net.Conn]struct{}),
	}

	r.wg.Add(1)
	go r.run()

	return r
}

// accept the next connection the edge terminates.
func (r *sniRouter) accept() (*StdConn, error) {
	select {
	case conn := <-r.conns:
		return conn, nil
	case <-r.stopped:
		return nil, net.ErrClosed
	}
}

// close stops the router once the listener is closed: it closes the
// connections it still holds and waits for its goroutines.
func (r *sniRouter) close() {
	r.cancel()

	r.mu.Lock()
	for conn := range r.active {
		_ = conn.Close()
	}
	r.mu.Unlock()

	r.wg.Wait()
}

func (r *sniRouter) run() {
	defer r.wg.Done()
	defer close(r.stopped)

	var backoff time.Duration
	for {
		conn, err := r.lis.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			backoff = min(max(2*backoff, 5*time.Millisecond),
				_acceptBackoffMax)
			szlog.Errorf("network: accept: %v, retrying in %v", err, backoff)
			time.Sleep(backoff)

			continue
		}

		backoff = 0
		std := newStdConn(conn)
		r.hold(std)
		r.wg.Add(1)
		go r.route(std)
	}
}

// route pipes conn to its upstream when its server name is passthrough,
// else hands it to accept.
func (r *sniRouter) route(conn *StdConn) {
	defer r.wg.Done()
	defer r.release(conn)

	if name, addrs, ok := r.passthroughOf(conn); ok {
		r.passthrough(conn, name, addrs)
		return
	}

	// not a TLS client hello, or one the edge terminates
	select {
	case r.conns <- conn:
	case <-r.stopped:
		_ = conn.Close()
	}
}

// passthroughOf the server name of the client hello of conn and its
// upstreams, when the edge passes its TLS through.
func (r *sniRouter) passthroughOf(conn *StdConn) (string, []string, bool) {
	name, err := peekServerName(conn, _helloTimeout)
	if err != nil || name == "" {
		return "", nil, false
	}

	addrs, ok := r.resolve(name)

	return name, addrs, ok
}

func (r *sniRouter) passthrough(conn *StdConn, name string, addrs []string) {
	up, err := r.dial(addrs)
	if err != nil {
		szlog.Errorf("network: passthrough %s: %v", name, err)
		_ = conn.Close()

		return
	}

	r.hold(up)
	defer r.release(up)

	szlog.DebugFn(func() string {
		return "network: passthrough " + conn.Id + ": " + name + " -> " +
			up.RemoteAddr().String()
	})

	pipe(conn, up)
}

// dial the first of addrs that answers.
func (r *sniRouter) dial(addrs []string) (net.Conn, error) {
	d := &net.Dialer{Timeout: _passthroughDialTimeout}

	errs := make([]error, 0, len(addrs))
	for _, addr := range addrs {
		up, err := d.DialContext(r.ctx, "tcp", addr)
		if err == nil {
			return up, nil
		}

		errs = append(errs, err)
	}

	return nil, errors.Join(errs...)
}

func (r *sniRouter) hold(conn net.Conn) {
	r.mu.Lock()
	r.active[conn] = struct{}{}
	r.mu.Unlock()
}

func (r *sniRouter) release(conn net.Conn) {
	r.mu.Lock()
	delete(r.active, conn)
	r.mu.Unlock()
}

// pipe copies a and b both ways until both are done, then closes them.
func pipe(a, b net.Conn) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.Copy(b, a)
		_ = closeWrite(b)
	}()

	_, _ = io.Copy(a, b)
	_ = closeWrite(a)
	<-done

	_ = a.Close()
	_ = b.Close()
}

// closeWrite half-closes conn when it can (TCP, or a conn wrapping one):
// the peer reads EOF and may still answer. Else it closes conn.
func closeWrite(conn net.Conn) error {
	if cw, ok := conn.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}

	return conn.Close()
}

// peekServerName reads the client hello of conn and returns its server
// name; what it read is replayed by the next reads of conn.
func peekServerName(conn *StdConn, timeout time.Duration) (string, error) {
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()

	var (
		read bytes.Buffer
		name string
	)

	hello := &helloConn{Conn: conn.Conn, r: io.TeeReader(conn.Conn, &read)}
	err := tls.Server(hello, &tls.Config{
		GetConfigForClient: func(chi *tls.ClientHelloInfo) (*tls.Config,
			error) {
			name = chi.ServerName
			return nil, errHelloRead
		},
	}).Handshake()

	// served again by the next reads of conn, its pending being empty
	conn.pending = read.Bytes()

	if !errors.Is(err, errHelloRead) {
		return "", err
	}

	return name, nil
}

// helloConn a conn the TLS server reads the client hello from and can not
// write to: its alert never reaches the client.
type helloConn struct {
	net.Conn
	r io.Reader
}

func (c *helloConn) Read(b []byte) (int, error) { return c.r.Read(b) }

func (*helloConn) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func (*helloConn) Close() error { return nil }
