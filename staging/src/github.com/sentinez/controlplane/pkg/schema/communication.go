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

package schema

import (
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/sztable"
	communicationpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/communication/v1"
)

// CommunicationNotifications the schema of the
// sztable.CommunicationNotifications table: listed per tenant, the newest
// first.
func CommunicationNotifications() dbx.Schema {
	return dbx.Schema{
		Table:     sztable.CommunicationNotifications,
		CreatedAt: communicationpb.Notification_MetadataCreatedAt,
		UpdatedAt: communicationpb.Notification_MetadataUpdatedAt,
		Columns: []dbx.Column{
			{Name: communicationpb.Notification_Id, Type: dbx.Text},
			{Name: communicationpb.Notification_EventTitle, Type: dbx.Text},
			{Name: communicationpb.Notification_EventNotiType, Type: dbx.Int4},
			{Name: communicationpb.Notification_EventMessage, Type: dbx.Text},
			{Name: communicationpb.Notification_TenantId, Type: dbx.Text},
		},
		Indexes: []dbx.Index{{
			Columns: []string{
				communicationpb.Notification_TenantId,
				communicationpb.Notification_MetadataCreatedAt,
			},
			Method: dbx.BTree,
		}},
	}
}
