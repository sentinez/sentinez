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

// Package centraldatafac builds centraldata services and handlers.
package centraldatafac

import (
	centraldatahdl "github.com/sentinez/controlplane/apps/centraldata/v1/handler"
	memrepo "github.com/sentinez/controlplane/apps/centraldata/v1/repos/memory"
	centraldatasvc "github.com/sentinez/controlplane/apps/centraldata/v1/service"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// NewDefaultService returns the centraldata service backed by memory.
func NewDefaultService(
	conf *settingpb.Config,
) (*centraldatasvc.CentralDataService, error) {
	memory, err := memrepo.New()
	if err != nil {
		return nil, err
	}

	return centraldatasvc.New(conf, memory), nil
}

// NewDefaultHandler returns the default centraldata handler.
func NewDefaultHandler(
	conf *settingpb.Config,
) (centraldatapb.CentralDataServiceServer, error) {
	service, err := NewDefaultService(conf)
	if err != nil {
		return nil, err
	}

	return centraldatahdl.New(service), nil
}
