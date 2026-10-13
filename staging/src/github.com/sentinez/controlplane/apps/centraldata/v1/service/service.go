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

package centraldatasvc

import (
	"context"

	"github.com/sentinez/core/storage/dbx/postgres"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var _ pb.CentralDataServiceServer = (*CentralDataService)(nil)

type CentralDataService struct {
	config *settingpb.Config
	tx     *postgres.Tx
}

func New(config *settingpb.Config,
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
