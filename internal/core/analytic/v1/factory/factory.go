package analyticfac

import (
	"context"

	pb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/analytic/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	analytichdl "github.com/sentinez/sentinez/internal/core/analytic/v1/handler"
	activitiesrepo "github.com/sentinez/sentinez/internal/core/analytic/v1/repos/activities"
	analyticsvc "github.com/sentinez/sentinez/internal/core/analytic/v1/service"
	"github.com/sentinez/sentinez/pkg/storage/dbx/postgres"
	"github.com/sentinez/shared/zlog"
)

func NewDefaultService(
	ctx context.Context, appConf *confpb.Config) *analyticsvc.AnalyticService {

	activitiesrepos, err := activitiesrepo.New(ctx, appConf)
	if err != nil {
		zlog.Errorf("analyticfac: init activities repo err=%v", err)
	}

	tx := postgres.NewTX(appConf)

	return analyticsvc.New(appConf, tx, activitiesrepos)
}

func NewDefaultHandler(ctx context.Context, appConf *confpb.Config,
) pb.AnalyticServiceServer {

	service := NewDefaultService(ctx, appConf)

	return analytichdl.New(service)
}
