package centraldatafac

import (
	"context"

	"github.com/sentinez/core/storage/dbx/postgres"
	centraldatahdl "github.com/sentinez/modules/centraldata/v1/handler"
	centraldatasvc "github.com/sentinez/modules/centraldata/v1/service"
	pb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/centraldata/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
)

func NewDefaultService(_ context.Context,
	appConf *confpb.Config) *centraldatasvc.CentralDataService {

	tx := postgres.NewTX(appConf)

	return centraldatasvc.New(appConf, tx)
}

func NewDefaultHandler(ctx context.Context, appConf *confpb.Config,
) pb.CentralDataServiceServer {

	service := NewDefaultService(ctx, appConf)

	return centraldatahdl.New(service)
}
