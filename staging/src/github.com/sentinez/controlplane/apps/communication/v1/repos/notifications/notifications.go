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

// Package notificationrepo stores the notifications of the tenants.
package notificationrepo

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szprefix"
	"github.com/sentinez/pkg/szrand"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	communicationpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/communication/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ INotification = (*Notifications)(nil)

// INotification is the store of the notifications.
//
// nolint
type INotification interface {
	Create(ctx context.Context, notification *communicationpb.Notification) (*communicationpb.Notification, error)
	Update(ctx context.Context, notification *communicationpb.Notification) error
	Get(ctx context.Context, id string) (*communicationpb.Notification, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) INotification

	List(ctx context.Context, req *communicationpb.ListNotificationsRequest) (*communicationpb.ListNotificationsResponse, error)
}

// Notifications is the INotification of the
// sztable.CommunicationNotifications table.
type Notifications struct {
	storage dbx.Database[communicationpb.Notification]
}

// New opens the notification table, created when missing.
func New(ctx context.Context,
	appConf *settingpb.Config) (INotification, error) {
	storage, err := postgres.New[communicationpb.Notification](ctx, appConf,
		schema.CommunicationNotifications())
	if err != nil {
		return nil, err
	}

	return &Notifications{storage: storage}, nil
}

// WithTX implements INotification.
func (n *Notifications) WithTX(tx *postgres.TxSession) INotification {
	return &Notifications{storage: postgres.WithTx(tx, n.storage)}
}

// listCondition is the WHERE condition of the filters of the request, all
// of them must match.
func listCondition(req *communicationpb.ListNotificationsRequest) sq.And {
	var conds sq.And

	// a slice becomes id IN (...)
	if ids := req.GetIds(); len(ids) != 0 {
		conds = append(conds, sq.Eq{communicationpb.Notification_Id: ids})
	}

	if tenantIds := req.GetTenantIds(); len(tenantIds) != 0 {
		conds = append(conds,
			sq.Eq{communicationpb.Notification_TenantId: tenantIds})
	}

	return conds
}

// List implements INotification: the newest first, paged or not.
func (n *Notifications) List(ctx context.Context,
	req *communicationpb.ListNotificationsRequest,
) (*communicationpb.ListNotificationsResponse, error) {
	cond := listCondition(req)

	builder := postgres.LimitOffset(n.selectQuery(), req.GetPage()).
		Where(cond).
		OrderBy(n.storage.CreatedAt() + " DESC")

	notifications, err := n.storage.CollectRows(ctx, builder,
		communicationpb.ScanNotification)
	if err != nil {
		return nil, err
	}

	var total int64
	if req.GetPage().GetTotal() {
		if total, err = n.total(ctx, cond); err != nil {
			return nil, err
		}
	}

	return &communicationpb.ListNotificationsResponse{
		Notifications: notifications,
		Total:         total,
	}, nil
}

// total counts the rows matching the list condition.
func (n *Notifications) total(ctx context.Context,
	cond sq.Sqlizer) (int64, error) {
	builder := postgres.CountBuilder(n.storage).Where(cond)

	var total int64
	if err := n.storage.Query(ctx, builder, &total); err != nil {
		return 0, err
	}

	return total, nil
}

// Create implements INotification; the id and the times are set.
func (n *Notifications) Create(ctx context.Context,
	notification *communicationpb.Notification,
) (*communicationpb.Notification, error) {
	// the times stored are those of the returned notification
	now := time.Now().UTC()
	notification.Id = szrand.NewUUID(szprefix.Notification)
	notification.Metadata = &typepb.Metadata{
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}

	id, tenantId := notification.GetId(), notification.GetTenantId()
	event := notification.GetEvent()
	query := postgres.InsertBuilder(n.storage, postgres.M{
		communicationpb.Notification_MetadataCreatedAt: now,
		communicationpb.Notification_MetadataUpdatedAt: now,
		communicationpb.Notification_Id:                id,
		communicationpb.Notification_EventTitle:        event.GetTitle(),
		communicationpb.Notification_EventNotiType:     event.GetNotiType(),
		communicationpb.Notification_EventMessage:      event.GetMessage(),
		communicationpb.Notification_TenantId:          tenantId,
	})

	if _, err := n.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return notification, nil
}

// Delete implements INotification.
func (n *Notifications) Delete(ctx context.Context, id string) error {
	return n.storage.Delete(ctx, id)
}

// Get implements INotification.
func (n *Notifications) Get(ctx context.Context,
	id string) (*communicationpb.Notification, error) {
	builder := n.selectQuery().
		Where(sq.Eq{communicationpb.Notification_Id: id})

	return n.storage.Select(ctx, builder,
		communicationpb.ScanOneNotification)
}

// Update implements INotification: the set fields of the event.
func (n *Notifications) Update(ctx context.Context,
	notification *communicationpb.Notification) error {
	query := postgres.UpdateBuilder(n.storage, notification.GetId())

	event := notification.GetEvent()
	if event.GetTitle() != "" {
		query = query.Set(communicationpb.Notification_EventTitle,
			event.GetTitle())
	}

	if event.GetNotiType() != 0 {
		query = query.Set(communicationpb.Notification_EventNotiType,
			event.GetNotiType())
	}

	if event.GetMessage() != "" {
		query = query.Set(communicationpb.Notification_EventMessage,
			event.GetMessage())
	}

	_, err := n.storage.Exec(ctx, query)

	return err
}

func (n *Notifications) selectQuery() sq.SelectBuilder {
	return postgres.SelectBuilder(n.storage, nil,
		communicationpb.NotificationColumns()...)
}
