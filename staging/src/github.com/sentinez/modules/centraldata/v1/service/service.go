package centraldatasvc

import (
	"context"

	"github.com/sentinez/core/storage/dbx/postgres"
	pb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
)

var _ pb.CentralDataServiceServer = (*CentralDataService)(nil)

type CentralDataService struct {
	config *settingpb.Config
	tx     *postgres.Tx
}

func New(config *settingpb.Config,
	tx *postgres.Tx,
) *CentralDataService {

	return &CentralDataService{
		config: config,
		tx:     tx,
	}
}

func (srv *CentralDataService) Status(ctx context.Context,
	req *pb.StatusRequest) (*pb.StatusResponse, error) {

	_ = ctx
	_ = req

	return &pb.StatusResponse{Msg: "OK"}, nil
}
