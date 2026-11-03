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

// Package collector implements the Collector service server.
package collector

import (
	collectorfac "github.com/sentinez/controlplane/apps/collector/v1/factory"
	"github.com/sentinez/core/szgrpc"
	collectorpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/collector/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// Collector serves CollectorServiceServer over gRPC.
type Collector struct {
	*szgrpc.Server
	handler collectorpb.CollectorServiceServer
	conf    *settingpb.Config
}

// NewService creates the collector gRPC server.
func NewService(conf *settingpb.Config) *Collector {
	return &Collector{
		Server:  szgrpc.New(),
		handler: collectorfac.NewDefaultHandler(conf),
		conf:    conf,
	}
}

// Start serves the collector, registered in consul; opts are applied after
// the defaults, e.g. szgrpc.WithAddress.
func (c *Collector) Start(opts ...szgrpc.ServerOption) error {
	collectorpb.RegisterCollectorServiceServer(c.AsServer(), c.handler)

	opts = append([]szgrpc.ServerOption{szgrpc.WithDiscorvery(true)}, opts...)

	return c.Serve(c.conf, opts...)
}
