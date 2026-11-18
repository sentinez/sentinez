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

package analyticsvc

import (
	"context"
	"time"

	auditlogsrepo "github.com/sentinez/controlplane/apps/collector/v1/repos/auditlogs"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListAuditLogs implements pb.AnalyticServiceServer: the logs of the
// tenants of req, all of them without; the handler scopes them.
func (srv *AnalyticService) ListAuditLogs(ctx context.Context,
	req *pb.ListAuditLogsRequest) (*pb.ListAuditLogsResponse, error) {
	filter, err := auditListFilter(req)
	if err != nil {
		return nil, err
	}

	logs, err := srv.events.AuditLog.List(ctx, filter)
	if err != nil {
		return nil, filterError(err)
	}

	return &pb.ListAuditLogsResponse{Logs: logs}, nil
}

// auditListFilter the filter of the list of req, of its tenants.
func auditListFilter(
	req *pb.ListAuditLogsRequest) (*eventfilter.Filter, error) {
	filter, err := listFilter(req)
	if err != nil {
		return nil, err
	}
	filter.TenantIds = req.GetTenantIds()

	return filter, nil
}

// GetAuditLogHistogram implements pb.AnalyticServiceServer: the logs of
// the tenants of req, all of them without; the handler scopes them.
func (srv *AnalyticService) GetAuditLogHistogram(ctx context.Context,
	req *pb.GetAuditLogHistogramRequest,
) (*pb.GetAuditLogHistogramResponse, error) {
	logs, err := histogramEvents(req)
	if err != nil {
		return nil, err
	}

	filter, window, err := keptHistogramFilter(req, time.Now(),
		auditlogsrepo.Retention)
	if err != nil {
		return nil, err
	}
	filter.TenantIds = req.GetTenantIds()

	points, err := srv.events.AuditLog.Histogram(ctx, filter, logs)
	if err != nil {
		return nil, filterError(err)
	}

	return &pb.GetAuditLogHistogramResponse{
		Points:    histogramPoints(points),
		Interval:  window.Interval,
		From:      timestamppb.New(window.From),
		To:        timestamppb.New(window.To),
		Truncated: window.Truncated,
	}, nil
}

// GetAuditLogTopValues implements pb.AnalyticServiceServer: the logs of
// the tenants of req, all of them without; the handler scopes them.
func (srv *AnalyticService) GetAuditLogTopValues(ctx context.Context,
	req *pb.GetAuditLogTopValuesRequest,
) (*pb.GetAuditLogTopValuesResponse, error) {
	filter, err := rangedFilter(req, time.Now(), _topRange)
	if err != nil {
		return nil, err
	}
	filter.TenantIds = req.GetTenantIds()

	result, err := srv.events.AuditLog.TopValues(ctx, filter,
		req.GetFields(), topLimit(req.GetLimit()))
	if err != nil {
		return nil, filterError(err)
	}

	top := topValuesResponse(result)

	return &pb.GetAuditLogTopValuesResponse{
		Total:  top.GetTotal(),
		Fields: top.GetFields(),
	}, nil
}
