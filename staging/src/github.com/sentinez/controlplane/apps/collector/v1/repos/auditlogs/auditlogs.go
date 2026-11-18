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

// Package auditlogsrepo stores the auditpb.AuditLog records consumed from
// kafka (sztopic.EventActivityAuditLog).
package auditlogsrepo

import (
	"context"
	"time"

	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventreader"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/topvalues"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/szstorage/dbx/timescale"
	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
	auditpb "github.com/sentinez/sentinez/api/proto/sentinez/types/audit/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// Retention how long the audit logs, and their histograms, are kept.
const Retention = 90 * 24 * time.Hour

var _ IAuditLog = (*AuditLogs)(nil)

// _spec the reader of the logs: their histogram is split by grpc code.
var _spec = eventreader.Spec[auditpb.AuditLog]{
	Hypertable: hypertable(),
	Key:        auditpb.AuditLog_Code,
	Columns: eventfilter.Columns{
		RequestId:   auditpb.AuditLog_Id,
		ResourceId:  auditpb.AuditLog_ResourceId,
		TenantId:    auditpb.AuditLog_TenantId,
		MatchedTime: auditpb.AuditLog_Timestamp,
		Fields:      _fields,
	},
	Select:   auditpb.AuditLogColumns(),
	Scan:     auditpb.ScanAuditLog[dbx.Rows],
	ScanOne:  auditpb.ScanOneAuditLog[dbx.Row],
	Values:   auditpb.AuditLogValues,
	IdColumn: auditpb.AuditLog_Id,
}

// _fields the fields of the logs filters match and top values count, in
// this order.
var _fields = []eventfilter.Field{
	eventfilter.Column(auditpb.AuditLog_Method),
	eventfilter.Column(auditpb.AuditLog_UserId),
	eventfilter.TypedColumn(auditpb.AuditLog_Code, dbx.Int4),
	eventfilter.TypedColumn(auditpb.AuditLog_ControlPlane, dbx.Int4),
	eventfilter.Column(auditpb.AuditLog_TenantId),
	eventfilter.Column(auditpb.AuditLog_RemoteAddr),
	eventfilter.Column(auditpb.AuditLog_UserAgent),
}

type IAuditLog interface {
	// Create inserts the logs, which have ids, in one statement; a log
	// whose id and timestamp are stored already is skipped, so inserting
	// a batch again is a no-op.
	Create(ctx context.Context, logs ...*auditpb.AuditLog) error
	Get(ctx context.Context, id string) (*auditpb.AuditLog, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IAuditLog

	List(ctx context.Context,
		filter *eventfilter.Filter) ([]*auditpb.AuditLog, error)
	// Histogram counts the logs per bucket, resource and grpc code
	// (histogram.Point.Key).
	// The logs of logs only when set (see eventreader.Reader).
	Histogram(ctx context.Context, filter *histogram.Filter,
		logs *eventfilter.Filter) ([]*histogram.Point, error)
	topvalues.TopValuer
}

// AuditLogs the repo of the sztable.AnalyticAuditLogsEvent hypertable.
type AuditLogs struct {
	*eventreader.Reader[auditpb.AuditLog]
}

// New returns the repo of the sztable.AnalyticAuditLogsEvent hypertable, and
// of its histograms.
func New(ctx context.Context,
	appConf *settingpb.Config) (IAuditLog, error) {
	reader, err := eventreader.New(ctx, appConf, _spec)
	if err != nil {
		return nil, err
	}

	return &AuditLogs{Reader: reader}, nil
}

// WithTX implements IAuditLog.
func (r *AuditLogs) WithTX(tx *postgres.TxSession) IAuditLog {
	return &AuditLogs{Reader: r.WithTx(tx)}
}

// hypertable the event hypertable of the logs, kept for Retention.
func hypertable() timescale.Hypertable {
	ht := schema.EventHypertable(schema.AnalyticAuditLogsEvent(),
		auditpb.AuditLog_Timestamp)
	ht.Retention = Retention

	return ht
}
