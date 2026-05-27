package analytichdl

import (
	"context"

	analyticsvc "github.com/sentinez/modules/analytic/v1/service"
	"github.com/sentinez/modules/pkg/headers"
	pb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/analytic/v1"
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
