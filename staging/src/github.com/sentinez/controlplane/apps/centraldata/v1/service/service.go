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

// Package centraldatasvc implements the centraldata business logic.
package centraldatasvc

import (
	"context"

	memrepo "github.com/sentinez/controlplane/apps/centraldata/v1/repos/memory"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// CentralDataService holds the centraldata dependencies.
type CentralDataService struct {
	config *settingpb.Config
	memory *memrepo.Memory
}

// New creates a CentralDataService keeping its data in memory.
func New(config *settingpb.Config,
	memory *memrepo.Memory) *CentralDataService {

	return &CentralDataService{
		config: config,
		memory: memory,
	}
}

// Status reports the service health.
func (srv *CentralDataService) Status(_ context.Context,
	_ *centraldatapb.StatusRequest,
) (*centraldatapb.StatusResponse, error) {
	return &centraldatapb.StatusResponse{Msg: "ok"}, nil
}
