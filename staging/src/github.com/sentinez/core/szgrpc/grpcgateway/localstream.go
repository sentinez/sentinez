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
	"io"
	"sync"

	"github.com/sentinez/core/szgrpc/validate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	_ grpc.ClientStream          = clientStream{}
	_ grpc.ServerStream          = serverStream{}
	_ grpc.ServerTransportStream = transportStream{}
)

// pipe is a stream of a LocalConn between a client and a handler. The
// messages are copied over unbuffered channels: a send returns once the
// other side took the message.
type pipe struct {
	desc   *grpc.StreamDesc
	method string

	clientCtx context.Context
	// serverCtx is done when the client's is, on Close of the conn and
	// when the handler returns.
	serverCtx context.Context
	cancel    context.CancelFunc
	stop      func() bool

	toServer  chan proto.Message
	toClient  chan proto.Message
	closeSend chan struct{}
	closeOnce sync.Once

	// headerSent is closed when the handler sends the header.
	headerSent chan struct{}
	// done is closed when the handler returned err.
	done chan struct{}
	err  error

	mu      sync.Mutex
	sent    bool
	header  metadata.MD
	trailer metadata.MD
}

// newPipe returns the pipe of a stream of method opened under ctx, ended
// as well when connCtx is done.
func newPipe(ctx, connCtx context.Context, desc *grpc.StreamDesc,
	method string) *pipe {
	serverCtx, cancel := context.WithCancel(incoming(ctx))

	p := &pipe{
		desc:       desc,
		method:     method,
		clientCtx:  ctx,
		cancel:     cancel,
		stop:       context.AfterFunc(connCtx, cancel),
		toServer:   make(chan proto.Message),
		toClient:   make(chan proto.Message),
		closeSend:  make(chan struct{}),
		headerSent: make(chan struct{}),
		done:       make(chan struct{}),
	}
	p.serverCtx = grpc.NewContextWithServerTransportStream(serverCtx,
		transportStream{p: p})

	return p
}

// finish ends the stream with err, the result of the handler.
func (p *pipe) finish(err error) {
	p.stop()
	p.cancel()
	p.err = err
	close(p.done)
}

// addHeader adds md to the header, until it is sent.
func (p *pipe) addHeader(md metadata.MD) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.sent {
		return status.Error(codes.Internal, "local conn: header already sent")
	}

	p.header = metadata.Join(p.header, md)

	return nil
}

// sendHeader adds md to the header and sends it, once.
func (p *pipe) sendHeader(md metadata.MD) error {
	if err := p.addHeader(md); err != nil {
		return err
	}

	p.flushHeader()

	return nil
}

// flushHeader sends the header unless it was: the first message does.
func (p *pipe) flushHeader() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.sent {
		return
	}

	p.sent = true
	close(p.headerSent)
}

func (p *pipe) addTrailer(md metadata.MD) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.trailer = metadata.Join(p.trailer, md)
}

// asMessage returns a copy of m to send: the sender may reuse m once the
// send returned.
func asMessage(m any) (proto.Message, error) {
	msg, err := message(m)
	if err != nil {
		return nil, err
	}

	return proto.Clone(msg), nil
}

// clientStream is the client side of a pipe.
type clientStream struct {
	p *pipe
}

// Header waits for the header, nil when the stream ended without one: its
// status is then that of RecvMsg.
func (s clientStream) Header() (metadata.MD, error) {
	p := s.p

	select {
	case <-p.headerSent:
	case <-p.done:
	case <-p.clientCtx.Done():
		return nil, status.FromContextError(p.clientCtx.Err()).Err()
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.sent {
		return nil, nil
	}

	return p.header.Copy(), nil
}

// Trailer returns the trailer, set once RecvMsg returned an error.
func (s clientStream) Trailer() metadata.MD {
	s.p.mu.Lock()
	defer s.p.mu.Unlock()

	return s.p.trailer.Copy()
}

func (s clientStream) CloseSend() error {
	s.p.closeOnce.Do(func() { close(s.p.closeSend) })
	return nil
}

func (s clientStream) Context() context.Context {
	return s.p.clientCtx
}

// SendMsg sends m to the handler. Once it returned, a stream of client
// messages gets io.EOF, its status from RecvMsg; a call of one message gets
// nil, as the generated code expects of grpc.
func (s clientStream) SendMsg(m any) error {
	p := s.p

	msg, err := asMessage(m)
	if err != nil {
		return err
	}

	select {
	case p.toServer <- msg:
		return nil
	case <-p.done:
		if p.desc.ClientStreams {
			return io.EOF
		}

		return nil
	case <-p.clientCtx.Done():
		return status.FromContextError(p.clientCtx.Err()).Err()
	}
}

// RecvMsg receives the next message of the handler into m; io.EOF once
// it returned nil, else its error.
func (s clientStream) RecvMsg(m any) error {
	p := s.p

	select {
	case msg := <-p.toClient:
		return copyMessage(m, msg)
	case <-p.done:
		if p.err != nil {
			return p.err
		}

		return io.EOF
	case <-p.clientCtx.Done():
		return status.FromContextError(p.clientCtx.Err()).Err()
	}
}

// serverStream is the handler side of a pipe.
type serverStream struct {
	p *pipe
}

func (s serverStream) SetHeader(md metadata.MD) error {
	return s.p.addHeader(md)
}

func (s serverStream) SendHeader(md metadata.MD) error {
	return s.p.sendHeader(md)
}

func (s serverStream) SetTrailer(md metadata.MD) {
	s.p.addTrailer(md)
}

func (s serverStream) Context() context.Context {
	return s.p.serverCtx
}

// SendMsg sends m to the client, and the header first.
func (s serverStream) SendMsg(m any) error {
	p := s.p

	msg, err := asMessage(m)
	if err != nil {
		return err
	}

	p.flushHeader()

	select {
	case p.toClient <- msg:
		return nil
	case <-p.serverCtx.Done():
		return status.FromContextError(p.serverCtx.Err()).Err()
	}
}

// RecvMsg receives the next message of the client into m and validates it;
// io.EOF once the client closed its side.
func (s serverStream) RecvMsg(m any) error {
	p := s.p

	select {
	case msg := <-p.toServer:
		if err := copyMessage(m, msg); err != nil {
			return err
		}

		return validate.Message(m)
	case <-p.closeSend:
		return io.EOF
	case <-p.serverCtx.Done():
		return status.FromContextError(p.serverCtx.Err()).Err()
	}
}

// transportStream lets the handler set the header and trailer through its
// context, as grpc.SetHeader does.
type transportStream struct {
	p *pipe
}

func (t transportStream) Method() string {
	return t.p.method
}

func (t transportStream) SetHeader(md metadata.MD) error {
	return t.p.addHeader(md)
}

func (t transportStream) SendHeader(md metadata.MD) error {
	return t.p.sendHeader(md)
}

func (t transportStream) SetTrailer(md metadata.MD) error {
	t.p.addTrailer(md)
	return nil
}
