package centraldatahdl

import (
	"context"

	centraldatasvc "github.com/sentinez/modules/centraldata/v1/service"
	pb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/centraldata/v1"
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
