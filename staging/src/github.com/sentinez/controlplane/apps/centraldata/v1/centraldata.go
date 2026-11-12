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
	centraldatafac "github.com/sentinez/controlplane/apps/centraldata/v1/factory"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/core/szgrpc"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// CentralData serves CentralDataServiceServer over gRPC.
type CentralData struct {
	*szgrpc.Server
	hdl  centraldatapb.CentralDataServiceServer
	conf *settingpb.Config
}

// NewService creates the centraldata gRPC server.
func NewService(conf *settingpb.Config) (*CentralData, error) {
	hdl, err := centraldatafac.NewDefaultHandler(conf)
	if err != nil {
		return nil, err
	}

	return &CentralData{
		Server: szgrpc.New(authz.ServerOptions()...),
		hdl:    hdl,
		conf:   conf,
	}, nil
}

// Start serves centraldata, registered in consul; opts are applied after
// the defaults, e.g. szgrpc.WithAddress.
func (c *CentralData) Start(opts ...szgrpc.ServerOption) error {
	centraldatapb.RegisterCentralDataServiceServer(c.AsServer(), c.hdl)

	opts = append([]szgrpc.ServerOption{szgrpc.WithDiscorvery(true)}, opts...)

	return c.Serve(c.conf, opts...)
}
