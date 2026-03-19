package centraldatasvc

import (
	"context"

	pb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/centraldata/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	"github.com/sentinez/sentinez/pkg/storage/dbx/postgres"
)

var _ pb.CentralDataServiceServer = (*CentralDataService)(nil)

type CentralDataService struct {
	config *confpb.Config
	tx     *postgres.Tx
}

func New(config *confpb.Config,
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
