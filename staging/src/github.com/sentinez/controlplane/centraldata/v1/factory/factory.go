package centraldatafac

import (
	"context"

	centraldatahdl "github.com/sentinez/controlplane/centraldata/v1/handler"
	centraldatasvc "github.com/sentinez/controlplane/centraldata/v1/service"
	"github.com/sentinez/core/storage/dbx/postgres"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/mods/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/setting/v1"
)

func NewDefaultService(_ context.Context,
	appConf *settingpb.Config) *centraldatasvc.CentralDataService {

	tx := postgres.NewTX(appConf)

	return centraldatasvc.New(appConf, tx)
}

func NewDefaultHandler(ctx context.Context, appConf *settingpb.Config,
) pb.CentralDataServiceServer {

	service := NewDefaultService(ctx, appConf)

	return centraldatahdl.New(service)
}
