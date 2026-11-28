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
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	commhdl "github.com/sentinez/controlplane/apps/communication/v1/handler"
	notificationmock "github.com/sentinez/controlplane/apps/communication/v1/repos/notifications/mock"
	membermock "github.com/sentinez/controlplane/apps/tenant/v1/repos/members/mock"
	resourcemock "github.com/sentinez/controlplane/apps/tenant/v1/repos/resources/mock"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/controlplane/pkg/headers/headerstest"
	"github.com/sentinez/core/szgrpc/client"
	"github.com/sentinez/core/szgrpc/grpcgateway"
	communicationpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/communication/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	notifypb "github.com/sentinez/sentinez/api/proto/sentinez/types/notify/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/sentinez/pkg/interceptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

// serve serves hdl over a local conn with opts, and the SSE route on an
// HTTP test server, whose URL it returns.
func serve(t *testing.T, hdl *commhdl.Communication,
	opts ...grpcgateway.LocalOption) string {
	t.Helper()

	cli, conn := client.NewLocalCommunication(hdl, opts...)
	t.Cleanup(func() { _ = conn.Close() })

	mux := http.NewServeMux()
	RegisterServerSentEvents(mux, cli)

	hs := httptest.NewServer(mux)
	// cleanups run last first: the server closes before the conn
	t.Cleanup(hs.Close)

	return hs.URL + defaults.SSEPath
}

// handler returns the communication handler of a caller member of
// tenantIds, its stored notifications created by the repo mock as given.
func handler(t *testing.T, tenantIds ...string) *commhdl.Communication {
	t.Helper()

	var members []*tenantpb.Member
	for _, id := range tenantIds {
		members = append(members, &tenantpb.Member{TenantId: id})
	}

	memberRepo := membermock.NewMockIMember(t)
	memberRepo.EXPECT().List(mock.Anything, mock.Anything).
		Return(&tenantpb.ListMemberResponse{Members: members}, nil).Maybe()

	notifications := notificationmock.NewMockINotification(t)
	notifications.EXPECT().Create(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context,
			n *communicationpb.Notification,
		) (*communicationpb.Notification, error) {
			return n, nil
		}).Maybe()

	az := authz.New(memberRepo, resourcemock.NewMockIResource(t))

	return commhdl.New(notifications, az)
}

// signedIn signs every stream in as user_1, as the authorization
// interceptor would for its token.
func signedIn() grpcgateway.LocalOption {
	return grpcgateway.WithStreamInterceptors(func(srv any,
		ss grpc.ServerStream, _ *grpc.StreamServerInfo,
		h grpc.StreamHandler) error {
		ctx := headerstest.User(ss.Context(), "user_1")
		return h(srv, &ctxStream{ServerStream: ss, ctx: ctx})
	})
}

type ctxStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *ctxStream) Context() context.Context { return s.ctx }

func TestServerSentEventsUnauthorized(t *testing.T) {
	url := serve(t, handler(t), interceptor.LocalOptions()...)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestServerSentEventsRelay sends the notifications of the tenants of the
// caller only.
func TestServerSentEventsRelay(t *testing.T) {
	hdl := handler(t, "tnt_a")
	url := serve(t, hdl, signedIn()) + "?tenantIds=tnt_a"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	// the stream is subscribed once its headers are sent
	notify(t, hdl, "tnt_other", "not yours")
	n := notify(t, hdl, "tnt_a", "hi")

	lines := readEvent(t, bufio.NewReader(resp.Body))
	require.Len(t, lines, 3)
	assert.Equal(t, "id: "+n.GetId(), lines[0])
	assert.Equal(t, "event: notification", lines[1])

	var got communicationpb.Notification
	data, ok := strings.CutPrefix(lines[2], "data: ")
	require.True(t, ok)
	require.NoError(t, protojson.Unmarshal([]byte(data), &got))
	assert.Equal(t, "hi", got.GetEvent().GetTitle())
}

func notify(t *testing.T, hdl *commhdl.Communication,
	tenantId, title string) *communicationpb.Notification {
	t.Helper()

	n, err := hdl.Notify(t.Context(), &communicationpb.Notification{
		Id:       "ntf_" + title,
		TenantId: tenantId,
		Event:    &notifypb.Notification{Title: title},
	})
	require.NoError(t, err)

	return n
}

// readEvent reads the lines of the next event, without the blank line
// ending it.
func readEvent(t *testing.T, r *bufio.Reader) []string {
	t.Helper()

	var lines []string
	for {
		line, err := r.ReadString('\n')
		require.NoError(t, err)

		line = strings.TrimSuffix(line, "\n")
		if line == "" {
			return lines
		}

		lines = append(lines, line)
	}
}
