package analyticfac

import (
	"context"

	"github.com/sentinez/core/storage/dbx/postgres"
	analytichdl "github.com/sentinez/modules/analytic/v1/handler"
	activitiesrepo "github.com/sentinez/modules/analytic/v1/repos/activities"
	analyticsvc "github.com/sentinez/modules/analytic/v1/service"
	pb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/analytic/v1"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
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
