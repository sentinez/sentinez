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

// Package communicationhdl provides the controller of the communication
// service.
package communicationhdl

import (
	"context"
	"fmt"
	"time"

	notificationrepo "github.com/sentinez/controlplane/apps/communication/v1/repos/notifications"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szlog"
	communicationpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/communication/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	notificationResponse = communicationpb.ServerStreamNotificationResponse
	notificationStream   = grpc.ServerStreamingServer[notificationResponse]
)

// Communication is the handler of CommunicationService.
type Communication struct {
	communicationpb.UnimplementedCommunicationServiceServer
	hub           *Hub
	notifications notificationrepo.INotification
	authz         *authz.Authorizer
}

var _ communicationpb.CommunicationServiceServer = (*Communication)(nil)

// New creates the communication handler of the stored notifications; az
// limits what a caller lists to its tenants.
func New(notifications notificationrepo.INotification,
	az *authz.Authorizer) *Communication {
	return &Communication{
		hub:           NewHub(),
		notifications: notifications,
		authz:         az,
	}
}

// Notify stores n, its id generated, then sends it to the open streams of
// the callers of its tenant. It is not sent when it cannot be stored.
func (c *Communication) Notify(ctx context.Context,
	n *communicationpb.Notification) (*communicationpb.Notification, error) {
	if n.GetTenantId() == "" {
		return nil, szerror.StatusInvalidArgumentF(
			"notification without tenant")
	}

	n, err := c.notifications.Create(ctx, n)
	if err != nil {
		return nil, err
	}

	c.hub.Publish(n)

	return n, nil
}

// SayHello implements CommunicationServiceServer.
func (c *Communication) SayHello(_ context.Context,
	msg *communicationpb.SayHelloRequest,
) (*communicationpb.SayHelloResponse, error) {
	szlog.Debugf("communication.SayHello: req = %v", msg)

	resp := fmt.Sprintf("Hello %s! Current time is %s",
		msg.GetName(), time.Now().Format(time.DateTime),
	)

	return &communicationpb.SayHelloResponse{Message: resp}, nil
}

// Status is the health check of consul.
func (c *Communication) Status(_ context.Context,
	msg *communicationpb.StatusRequest,
) (*communicationpb.StatusResponse, error) {
	szlog.Debugf("communication.Status: req = %v", msg)

	return &communicationpb.StatusResponse{Message: "ok"}, nil
}

// ServerStreamNotification streams the notifications of the tenants of the
// caller (of any tenant for the admin control plane), of those of
// tenant_ids among them when set, published from now until the caller
// leaves; the tenants are those of when it opens. The headers go first,
// once subscribed: the caller knows the stream is authorized, and misses
// nothing published after them.
func (c *Communication) ServerStreamNotification(
	req *communicationpb.ServerStreamNotificationRequest,
	stream notificationStream,
) error {
	ctx := stream.Context()

	notifications, unsubscribe, err := c.subscribe(ctx, req.GetTenantIds())
	if err != nil {
		return err
	}
	defer unsubscribe()

	if err := stream.SendHeader(metadata.MD{}); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case n := <-notifications:
			err := stream.Send(&notificationResponse{Notification: n})
			if err != nil {
				return err
			}
		}
	}
}

// ListNotifications lists the notifications of the tenants of the caller,
// of those of tenant_ids among them when set; the admin control plane
// lists those of any tenant.
func (c *Communication) ListNotifications(ctx context.Context,
	req *communicationpb.ListNotificationsRequest,
) (*communicationpb.ListNotificationsResponse, error) {
	req, ok, err := authz.ScopedTenants(ctx, c.authz, req)
	if err != nil || !ok {
		return &communicationpb.ListNotificationsResponse{}, err
	}

	return c.notifications.List(ctx, req)
}

// subscribe subscribes the caller of ctx to the notifications of its
// tenants, narrowed to tenantIds when set. A caller of no tenant gets a nil
// channel, which never receives: its stream stays open, idle.
func (c *Communication) subscribe(ctx context.Context, tenantIds []string,
) (<-chan *communicationpb.Notification, func(), error) {
	scope, err := c.authz.TenantIds(ctx, tenantIds)
	if err != nil {
		return nil, nil, err
	}

	if scope.None {
		return nil, func() {}, nil
	}

	notifications, unsubscribe := c.hub.Subscribe(tenantFilter(scope.Ids...))

	return notifications, unsubscribe, nil
}
