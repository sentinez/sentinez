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

package centraldatahdl

import (
	"context"

	centraldatasvc "github.com/sentinez/controlplane/apps/centraldata/v1/service"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	"github.com/sentinez/shared/zlog"
)

var _ pb.CentralDataServiceServer = (*CentralDataHandler)(nil)

type CentralDataHandler struct {
	service *centraldatasvc.CentralDataService
}

func New(
	service *centraldatasvc.CentralDataService,
) pb.CentralDataServiceServer {
	return &CentralDataHandler{
		service: service,
	}
}

func (hdl *CentralDataHandler) Status(ctx context.Context,
	req *pb.StatusRequest) (*pb.StatusResponse, error) {

	zlog.Debugf("[CentralDataHandler.Status] req = %v", req)

	resp, err := hdl.service.Status(ctx, req)
	if err != nil {
		zlog.Errorf("[CentralDataHandler.Status] failed: %v", err)
		return nil, err
	}

	return resp, nil
}
