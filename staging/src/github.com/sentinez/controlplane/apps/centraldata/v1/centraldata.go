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

// Package centraldata implements the CentralData service server.
package centraldata

import (
	"context"
	"fmt"

	centraldatafac "github.com/sentinez/controlplane/apps/centraldata/v1/factory"
	centraldatasvc "github.com/sentinez/controlplane/apps/centraldata/v1/service"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/core/szgrpc"
	"github.com/sentinez/sentinez/api/client/connection"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"google.golang.org/grpc"
)

// CentralData serves CentralDataServiceServer over gRPC.
type CentralData struct {
	*szgrpc.Server
	hdl    centraldatapb.CentralDataServiceServer
	conf   *settingpb.Config
	syncer *centraldatasvc.Syncer
}

// NewService creates the centraldata gRPC server. publisher sends the
// changes of the tables to the edges. The server requires the client
// certificate of the internal CA when SENZ_TLS_* are set.
func NewService(conf *settingpb.Config,
	publisher centraldatasvc.Publisher) (*CentralData, error) {

	service, err := centraldatafac.NewDefaultService(conf)
	if err != nil {
		return nil, err
	}

	syncer, err := centraldatafac.NewSyncer(conf, service, publisher)
	if err != nil {
		return nil, fmt.Errorf("new syncer: %w", err)
	}

	opts, err := serverOptions(conf)
	if err != nil {
		return nil, err
	}

	return &CentralData{
		Server: szgrpc.New(opts...),
		hdl:    centraldatafac.NewDefaultHandler(service),
		conf:   conf,
		syncer: syncer,
	}, nil
}

func serverOptions(conf *settingpb.Config) ([]grpc.ServerOption, error) {
	opts := authz.ServerOptions()

	tls := connection.TLSFromConfig(conf)
	if !tls.Enabled() {
		return opts, nil
	}

	creds, err := tls.ServerCredentials()
	if err != nil {
		return nil, err
	}

	return append(opts, grpc.Creds(creds)), nil
}

// StartSync loads the tables from postgres and keeps them in sync until
// Shutdown. It does nothing without a database.
func (c *CentralData) StartSync(ctx context.Context) error {
	if c.syncer == nil {
		return nil
	}

	return c.syncer.Start(ctx)
}

// Start serves centraldata, registered in consul; opts are applied after
// the defaults, e.g. szgrpc.WithAddress.
func (c *CentralData) Start(opts ...szgrpc.ServerOption) error {
	centraldatapb.RegisterCentralDataServiceServer(c.AsServer(), c.hdl)

	opts = append([]szgrpc.ServerOption{szgrpc.WithDiscorvery(true)}, opts...)

	return c.Serve(c.conf, opts...)
}

// Shutdown stops the sync, then the server.
func (c *CentralData) Shutdown(ctx context.Context) error {
	if c.syncer != nil {
		c.syncer.Close()
	}

	return c.Server.Shutdown(ctx)
}
