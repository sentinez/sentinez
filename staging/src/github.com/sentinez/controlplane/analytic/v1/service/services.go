package analyticsvc

import (
	"context"

	activitiesrepo "github.com/sentinez/controlplane/analytic/v1/repos/activities"
	"github.com/sentinez/core/storage/dbx/postgres"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/mods/analytic/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/setting/v1"
)

var _ pb.AnalyticServiceServer = (*AnalyticService)(nil)

type AnalyticService struct {
	config     *settingpb.Config
	tx         *postgres.Tx
	activities activitiesrepo.IActivity
}

func New(config *settingpb.Config,
	tx *postgres.Tx,
	activities activitiesrepo.IActivity,
) *AnalyticService {

	return &AnalyticService{
		config:     config,
		tx:         tx,
		activities: activities,
	}
}

func (srv *AnalyticService) Status(ctx context.Context,
	req *pb.StatusRequest) (*pb.StatusResponse, error) {

	_ = ctx
	_ = req

	return &pb.StatusResponse{Msg: "OK"}, nil
}

func (srv *AnalyticService) ListActivities(ctx context.Context,
	req *pb.ListActivitiesRequest) (*pb.ListActivitiesResponse, error) {

	return srv.activities.List(ctx, req)
}
