package centraldatafac

import (
	"context"

	pb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/centraldata/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	centraldatahdl "github.com/sentinez/sentinez/internal/core/centraldata/v1/handler"
	centraldatasvc "github.com/sentinez/sentinez/internal/core/centraldata/v1/service"
	"github.com/sentinez/sentinez/pkg/storage/dbx/postgres"
)

func NewDefaultService(
	ctx context.Context, appConf *confpb.Config) *centraldatasvc.CentralDataService {

	tx := postgres.NewTX(appConf)

	return centraldatasvc.New(appConf, tx)
}

func NewDefaultHandler(ctx context.Context, appConf *confpb.Config,
) pb.CentralDataServiceServer {

	service := NewDefaultService(ctx, appConf)

	return centraldatahdl.New(service)
}
