package centraldatahdl

import (
	"context"

	pb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/centraldata/v1"
	centraldatasvc "github.com/sentinez/sentinez/internal/core/centraldata/v1/service"
	"github.com/sentinez/shared/zlog"
)

var _ pb.CentralDataServiceServer = (*CentralDataHandler)(nil)

type CentralDataHandler struct {
	service *centraldatasvc.CentralDataService
}

func New(
	service *centraldatasvc.CentralDataService,
) pb.CentralDataServiceServer {
	return &CentralDataHandler{
		service: service,
	}
}

func (hdl *CentralDataHandler) Status(ctx context.Context,
	req *pb.StatusRequest) (*pb.StatusResponse, error) {

	zlog.Debugf("[CentralDataHandler.Status] req = %v", req)

	resp, err := hdl.service.Status(ctx, req)
	if err != nil {
		zlog.Errorf("[CentralDataHandler.Status] failed: %v", err)
		return nil, err
	}

	return resp, nil
}
