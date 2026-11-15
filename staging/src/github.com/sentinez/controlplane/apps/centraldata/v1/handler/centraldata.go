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

// Package centraldatahdl provides the controller for the centraldata
// service.
package centraldatahdl

import (
	"context"

	centraldatasvc "github.com/sentinez/controlplane/apps/centraldata/v1/service"
	"github.com/sentinez/pkg/szlog"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	"google.golang.org/grpc"
)

var _ centraldatapb.CentralDataServiceServer = (*CentralDataHandler)(nil)

// CentralDataHandler handles CentralDataService RPCs.
type CentralDataHandler struct {
	centraldatapb.UnimplementedCentralDataServiceServer
	service *centraldatasvc.CentralDataService
}

// New creates a new CentralData handler.
func New(
	service *centraldatasvc.CentralDataService,
) centraldatapb.CentralDataServiceServer {
	return &CentralDataHandler{service: service}
}

// Status is the consul health check.
func (hdl *CentralDataHandler) Status(ctx context.Context,
	req *centraldatapb.StatusRequest,
) (*centraldatapb.StatusResponse, error) {
	szlog.Debugf("centraldata.Status: req = %v", req)

	return hdl.service.Status(ctx, req)
}

// Sync streams the synced tables, see CentralDataService.Sync.
func (hdl *CentralDataHandler) Sync(req *centraldatapb.SyncRequest,
	stream grpc.ServerStreamingServer[centraldatapb.SyncResponse],
) error {
	szlog.Debugf("centraldata.Sync: tables = %v", req.GetTables())

	return hdl.service.Sync(stream.Context(), req, stream.Send)
}
