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

package communicationhdl

import (
	"context"
	"testing"

	notificationmock "github.com/sentinez/controlplane/apps/communication/v1/repos/notifications/mock"
	membermock "github.com/sentinez/controlplane/apps/tenant/v1/repos/members/mock"
	resourcemock "github.com/sentinez/controlplane/apps/tenant/v1/repos/resources/mock"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/controlplane/pkg/headers/headerstest"
	communicationpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/communication/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// handlerOf returns the handler of a caller member of tenantIds, and its
// notification repo.
func handlerOf(t *testing.T, tenantIds ...string,
) (*Communication, *notificationmock.MockINotification) {
	t.Helper()

	var members []*tenantpb.Member
	for _, id := range tenantIds {
		members = append(members, &tenantpb.Member{TenantId: id})
	}

	memberRepo := membermock.NewMockIMember(t)
	memberRepo.EXPECT().List(mock.Anything, mock.Anything).
		Return(&tenantpb.ListMemberResponse{Members: members}, nil).Maybe()

	notifications := notificationmock.NewMockINotification(t)
	az := authz.New(memberRepo, resourcemock.NewMockIResource(t))

	return New(notifications, az), notifications
}

func portal(t *testing.T) context.Context {
	t.Helper()

	return headerstest.User(t.Context(), "user_1")
}

// tenantsOf matches a list request for tenantIds.
func tenantsOf(tenantIds ...string) any {
	return mock.MatchedBy(
		func(req *communicationpb.ListNotificationsRequest) bool {
			return assert.ObjectsAreEqual(tenantIds, req.GetTenantIds())
		})
}

func TestListNotificationsOfCallerTenants(t *testing.T) {
	h, notifications := handlerOf(t, "tnt_a", "tnt_b")
	want := &communicationpb.ListNotificationsResponse{Total: 2}
	notifications.EXPECT().
		List(mock.Anything, tenantsOf("tnt_a", "tnt_b")).
		Return(want, nil)

	got, err := h.ListNotifications(portal(t),
		&communicationpb.ListNotificationsRequest{})

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestListNotificationsNarrowsRequestedTenants(t *testing.T) {
	h, notifications := handlerOf(t, "tnt_a")
	notifications.EXPECT().
		List(mock.Anything, tenantsOf("tnt_a")).
		Return(&communicationpb.ListNotificationsResponse{}, nil)

	_, err := h.ListNotifications(portal(t),
		&communicationpb.ListNotificationsRequest{
			TenantIds: []string{"tnt_a", "tnt_other"},
		})

	require.NoError(t, err)
}

func TestListNotificationsWithoutTenantSeesNothing(t *testing.T) {
	// the repo mock fails the test if List is called
	h, _ := handlerOf(t)

	got, err := h.ListNotifications(portal(t),
		&communicationpb.ListNotificationsRequest{})

	require.NoError(t, err)
	assert.Empty(t, got.GetNotifications())
}

func TestListNotificationsWithoutCallerFails(t *testing.T) {
	h, _ := handlerOf(t)

	_, err := h.ListNotifications(t.Context(),
		&communicationpb.ListNotificationsRequest{})

	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestNotifyWithoutTenantFails(t *testing.T) {
	// the repo mock fails the test if Create is called
	h, _ := handlerOf(t)

	_, err := h.Notify(t.Context(), &communicationpb.Notification{})

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestNotifyStoresThenPublishes(t *testing.T) {
	h, notifications := handlerOf(t)
	stored := &communicationpb.Notification{Id: "ntf_1", TenantId: "tnt_a"}
	notifications.EXPECT().Create(mock.Anything, mock.Anything).
		Return(stored, nil)
	got, unsubscribe := h.hub.Subscribe(tenantFilter("tnt_a"))
	defer unsubscribe()

	n, err := h.Notify(t.Context(),
		&communicationpb.Notification{TenantId: "tnt_a"})

	require.NoError(t, err)
	assert.Equal(t, stored, n)
	assert.Equal(t, stored, <-got)
}

func TestNotifyNotPublishedWhenNotStored(t *testing.T) {
	h, notifications := handlerOf(t)
	notifications.EXPECT().Create(mock.Anything, mock.Anything).
		Return(nil, assert.AnError)
	got, unsubscribe := h.hub.Subscribe(tenantFilter())
	defer unsubscribe()

	_, err := h.Notify(t.Context(),
		&communicationpb.Notification{TenantId: "tnt_a"})

	require.ErrorIs(t, err, assert.AnError)
	assert.Empty(t, got)
}

type subscribeTest struct {
	name          string
	giveTenants   []string
	giveCtx       func(*testing.T) context.Context
	giveRequested []string
	// wantReceived by tenant id: whether its notification is received
	wantReceived map[string]bool
}

func admin(t *testing.T) context.Context {
	t.Helper()

	return headerstest.Admin(t.Context())
}

var _subscribeTests = []subscribeTest{
	{
		name:         "member",
		giveTenants:  []string{"tnt_a", "tnt_b"},
		giveCtx:      portal,
		wantReceived: map[string]bool{"tnt_a": true, "tnt_c": false},
	},
	{
		name:          "requested",
		giveTenants:   []string{"tnt_a", "tnt_b"},
		giveCtx:       portal,
		giveRequested: []string{"tnt_b", "tnt_other"},
		wantReceived:  map[string]bool{"tnt_a": false, "tnt_b": true},
	},
	{
		name:         "no tenant",
		giveCtx:      portal,
		wantReceived: map[string]bool{"tnt_a": false},
	},
	{
		name:         "admin",
		giveCtx:      admin,
		wantReceived: map[string]bool{"tnt_a": true, "tnt_b": true},
	},
}

func TestSubscribe(t *testing.T) {
	for _, tt := range _subscribeTests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := handlerOf(t, tt.giveTenants...)

			got, unsubscribe, err := h.subscribe(tt.giveCtx(t),
				tt.giveRequested)
			require.NoError(t, err)
			defer unsubscribe()

			for tenantId, want := range tt.wantReceived {
				h.hub.Publish(
					&communicationpb.Notification{TenantId: tenantId})
				assert.Equal(t, want, len(got) == 1, tenantId)
				drain(got)
			}
		})
	}
}

// drain empties a buffered channel, nil or not.
func drain(ch <-chan *communicationpb.Notification) {
	for len(ch) > 0 {
		<-ch
	}
}
