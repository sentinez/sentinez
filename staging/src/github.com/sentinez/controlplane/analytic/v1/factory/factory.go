package analyticfac

import (
	"context"

	analytichdl "github.com/sentinez/controlplane/analytic/v1/handler"
	activitiesrepo "github.com/sentinez/controlplane/analytic/v1/repos/activities"
	analyticsvc "github.com/sentinez/controlplane/analytic/v1/service"
	"github.com/sentinez/core/storage/dbx/postgres"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/zlog"
)

func NewDefaultService(ctx context.Context,
	appConf *settingpb.Config) *analyticsvc.AnalyticService {

	activitiesrepos, err := activitiesrepo.New(ctx, appConf)
	if err != nil {
		zlog.Errorf("analyticfac: init activities repo err=%v", err)
	}

	tx := postgres.NewTX(appConf)

	return analyticsvc.New(appConf, tx, activitiesrepos)
}

func NewDefaultHandler(ctx context.Context, appConf *settingpb.Config,
) pb.AnalyticServiceServer {

	service := NewDefaultService(ctx, appConf)

	return analytichdl.New(service)
}
