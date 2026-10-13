// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package analyticfac

import (
	"context"

	analytichdl "github.com/sentinez/controlplane/apps/analytic/v1/handler"
	activitiesrepo "github.com/sentinez/controlplane/apps/analytic/v1/repos/activities"
	analyticsvc "github.com/sentinez/controlplane/apps/analytic/v1/service"
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
