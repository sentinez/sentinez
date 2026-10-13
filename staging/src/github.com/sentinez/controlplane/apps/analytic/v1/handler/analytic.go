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

	analyticsvc "github.com/sentinez/controlplane/apps/analytic/v1/service"
	"github.com/sentinez/controlplane/pkg/headers"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	"github.com/sentinez/shared/zlog"
)

var _ pb.AnalyticServiceServer = (*AnalyticHandler)(nil)

type AnalyticHandler struct {
	service *analyticsvc.AnalyticService
}

func New(
	service *analyticsvc.AnalyticService,
) pb.AnalyticServiceServer {
	return &AnalyticHandler{
		service: service,
	}
}

func (hdl *AnalyticHandler) Status(ctx context.Context,
	req *pb.StatusRequest) (*pb.StatusResponse, error) {

	zlog.Debugf("[AnalyticHandler.Status] req = %v", req)

	resp, err := hdl.service.Status(ctx, req)
	if err != nil {
		zlog.Errorf("[AnalyticHandler.Status] failed: %v", err)
		return nil, err
	}

	return resp, nil
}

func (hdl *AnalyticHandler) ListActivities(ctx context.Context,
	req *pb.ListActivitiesRequest) (*pb.ListActivitiesResponse, error) {

	zlog.Debugf("[AnalyticHandler.ListActivities] req = %v", req)

	ss, err := headers.GetAuth(ctx)
	if err != nil {
		return nil, err
	}

	err = ss.Check(pb.GetAnalyticServiceListActivities())
	if err != nil {
		return nil, err
	}

	resp, err := hdl.service.ListActivities(ctx, req)
	if err != nil {
		zlog.Errorf("[AnalyticHandler.ListActivities] error: %v", err)
		return nil, err
	}

	return resp, nil
}
