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

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sentinez/controlplane/pkg/headers"
	communicationpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/communication/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// _heartbeat is the interval of the comments sent on an idle stream, so
// that proxies do not close it.
const _heartbeat = 15 * time.Second

type (
	notificationResponse = communicationpb.ServerStreamNotificationResponse
	notificationStream   = grpc.ServerStreamingClient[notificationResponse]
)

// RegisterServerSentEvents serves at defaults.SSEPath the notification
// stream of cli as server-sent events: one "notification" event per
// notification, its data the JSON of communicationpb.Notification. The
// caller authenticates with its Authorization header, as on the API: an
// EventSource cannot send it, the console reads the stream with fetch.
func RegisterServerSentEvents(mux *http.ServeMux,
	cli communicationpb.CommunicationServiceClient) {
	mux.Handle("GET "+defaults.SSEPath, &sse{cli: cli})
}

type sse struct {
	cli communicationpb.CommunicationServiceClient
}

func (s *sse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// cancelled on return: it ends the gRPC stream, then its receiver
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	stream, err := s.open(ctx, r.Header.Get(headers.AuthHeader),
		r.URL.Query()["tenantIds"])
	if err != nil {
		st := status.Convert(err)
		http.Error(w, st.Message(), runtime.HTTPStatusFromCode(st.Code()))

		return
	}

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	// no-transform: a compressing proxy (the console's) would hold the
	// events in its gzip buffer
	h.Set("Cache-Control", "no-cache, no-transform")
	// nginx would buffer the events otherwise
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	relay(ctx, cancel, w, stream)
}

// open opens the notification stream of tenantIds (all the tenants of the
// caller when empty) as the caller of token, and waits for its headers: the
// call is then authorized, or its error returned.
func (s *sse) open(ctx context.Context, token string,
	tenantIds []string) (notificationStream, error) {
	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, headers.AuthHeader, token)
	}

	stream, err := s.cli.ServerStreamNotification(ctx,
		&communicationpb.ServerStreamNotificationRequest{
			TenantIds: tenantIds,
		})
	if err != nil {
		return nil, err
	}

	if md, err := stream.Header(); err == nil && md != nil {
		return stream, nil
	}

	// ended without headers: its status comes with the first message
	if _, err = stream.Recv(); err == nil {
		err = status.Error(codes.Internal, "notification stream: no header")
	}

	return nil, err
}

// relay writes the notifications of stream to w until either ends, and a
// heartbeat while it is idle. cancel ends the stream, opened under ctx.
func relay(ctx context.Context, cancel context.CancelFunc,
	w http.ResponseWriter, stream notificationStream) {
	rc := http.NewResponseController(w)
	if rc.Flush() != nil {
		return
	}

	notifications := make(chan *communicationpb.Notification)
	done := make(chan struct{})

	go func() {
		defer close(done)
		receive(ctx, stream, notifications)
	}()
	// the receiver blocks in Recv until the stream is cancelled
	defer func() {
		cancel()
		<-done
	}()

	write(w, rc, notifications)
}

// write writes each notification as an event, and a heartbeat while none
// comes, until notifications is closed (the stream ended or was cancelled)
// or a write fails.
func write(w http.ResponseWriter, rc *http.ResponseController,
	notifications <-chan *communicationpb.Notification) {
	ticker := time.NewTicker(_heartbeat)
	defer ticker.Stop()

	for {
		var err error

		select {
		case n, ok := <-notifications:
			if !ok {
				return
			}
			err = writeEvent(w, n)
		case <-ticker.C:
			_, err = fmt.Fprint(w, ": ping\n\n")
		}

		if err != nil || rc.Flush() != nil {
			return
		}
	}
}

// receive sends the notifications of stream to out, closed when the stream
// ends or ctx is done.
func receive(ctx context.Context, stream notificationStream,
	out chan<- *communicationpb.Notification) {
	defer close(out)

	for {
		resp, err := stream.Recv()
		if err != nil {
			return
		}

		select {
		case out <- resp.GetNotification():
		case <-ctx.Done():
			return
		}
	}
}

// writeEvent writes n as an event; protojson writes it on one line, as the
// data field of an event must be.
func writeEvent(w http.ResponseWriter, n *communicationpb.Notification) error {
	data, err := protojson.Marshal(n)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "id: %s\nevent: notification\ndata: %s\n\n",
		n.GetId(), data)

	return err
}
