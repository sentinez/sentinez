// Copyright 2025 Duc-Hung Ho.
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
	"context"
	"time"

	centraldatahdl "github.com/sentinez/controlplane/apps/centraldata/v1/handler"
	memrepo "github.com/sentinez/controlplane/apps/centraldata/v1/repos/memory"
	pgrepo "github.com/sentinez/controlplane/apps/centraldata/v1/repos/postgres"
	centraldatasvc "github.com/sentinez/controlplane/apps/centraldata/v1/service"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

const (
	// Channel the postgres channel the synced tables notify.
	Channel = "sentinez_centraldata"

	_openTimeout = 30 * time.Second
)

// NewDefaultService returns the centraldata service backed by memory,
// streaming the tables named by SENZ_SYNC_TABLES (all those synced by
// default).
func NewDefaultService(
	conf *settingpb.Config,
) (*centraldatasvc.CentralDataService, error) {
	tables, err := pgrepo.Parse(conf.Get(settingpb.Senz_SENZ_SYNC_TABLES))
	if err != nil {
		return nil, err
	}

	memory, err := memrepo.New()
	if err != nil {
		return nil, err
	}

	return centraldatasvc.New(conf, memory, tables...), nil
}

// NewDefaultHandler returns the handler of service.
func NewDefaultHandler(
	service *centraldatasvc.CentralDataService,
) centraldatapb.CentralDataServiceServer {
	return centraldatahdl.New(service)
}

// NewSyncer returns the Syncer that keeps the memory of service equal to
// its tables in postgres, creating them and their triggers. It returns nil
// when no database is configured: the service then streams nothing.
func NewSyncer(conf *settingpb.Config,
	service *centraldatasvc.CentralDataService,
	publisher centraldatasvc.Publisher,
) (*centraldatasvc.Syncer, error) {
	if conf.Get(settingpb.Senz_SENZ_POSTGRES_URI) == "" {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), _openTimeout)
	defer cancel()

	tables, err := pgrepo.Open(ctx, conf, service.Tables(), Channel)
	if err != nil {
		return nil, err
	}

	listen := func(ctx context.Context) (centraldatasvc.Subscription, error) {
		return postgres.Listen(ctx, conf, Channel)
	}

	return centraldatasvc.NewSyncer(service.Memory(), tables, listen,
		publisher, centraldatasvc.DefaultReconcile), nil
}
