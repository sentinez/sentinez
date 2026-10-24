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

package centraldatafac

import (
	"context"

	centraldatahdl "github.com/sentinez/controlplane/apps/centraldata/v1/handler"
	centraldatasvc "github.com/sentinez/controlplane/apps/centraldata/v1/service"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/szstorage/dbx/postgres"
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
