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

package analytichdl

import (
	"context"

	"github.com/sentinez/controlplane/pkg/authz"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
)

// ListAuditLogs implements pb.AnalyticServiceServer: the caller sees the
// logs of its tenants, all of them and those of no tenant on the admin
// control plane.
func (hdl *AnalyticHandler) ListAuditLogs(ctx context.Context,
	req *pb.ListAuditLogsRequest) (*pb.ListAuditLogsResponse, error) {
	req, ok, err := authz.ScopedTenantless(ctx, hdl.authz, req)
	if err != nil || !ok {
		return &pb.ListAuditLogsResponse{}, err
	}

	return hdl.service.ListAuditLogs(ctx, req)
}

// GetAuditLogHistogram implements pb.AnalyticServiceServer: the caller
// counts the logs of its tenants, all of them and those of no tenant on
// the admin control plane.
func (hdl *AnalyticHandler) GetAuditLogHistogram(ctx context.Context,
	req *pb.GetAuditLogHistogramRequest,
) (*pb.GetAuditLogHistogramResponse, error) {
	req, ok, err := authz.ScopedTenantless(ctx, hdl.authz, req)
	if err != nil || !ok {
		return &pb.GetAuditLogHistogramResponse{}, err
	}

	return hdl.service.GetAuditLogHistogram(ctx, req)
}

// GetAuditLogTopValues implements pb.AnalyticServiceServer: the caller
// counts the logs of its tenants, all of them and those of no tenant on
// the admin control plane.
func (hdl *AnalyticHandler) GetAuditLogTopValues(ctx context.Context,
	req *pb.GetAuditLogTopValuesRequest,
) (*pb.GetAuditLogTopValuesResponse, error) {
	req, ok, err := authz.ScopedTenantless(ctx, hdl.authz, req)
	if err != nil || !ok {
		return &pb.GetAuditLogTopValuesResponse{}, err
	}

	return hdl.service.GetAuditLogTopValues(ctx, req)
}
