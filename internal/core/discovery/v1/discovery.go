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

// Package discovery implements the discovery service
package discovery

import (
	"context"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/discovery/v1"
	dischdl "github.com/sentinez/sentinez/internal/core/discovery/v1/handler"
	"github.com/sentinez/sentinez/pkg/common/protobuf"
	grpcgw "github.com/sentinez/sentinez/pkg/core/gateway/grpc"
	"github.com/sentinez/sentinez/pkg/core/sentinez/v1"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

// make sure Discovery implement sentinez.Server
// it will start by sentinez.runner through sentinez.Server
var _ sentinez.Server = (*Discovery)(nil)

var _ = sentinez.Inject(dischdl.New)

// New creates a new discovery module.
func New(srv discovery.DiscoveryServiceServer, conf *common.Config,
	flag *common.FlagGRPCService) sentinez.Server {

	return &Discovery{
		Server: grpcgw.NewDefault(),
		srv:    srv,
		config: conf,
		flag:   flag,
	}
}

// Discovery implements DiscoveryServiceServer.
type Discovery struct {
	*grpcgw.Server // inherit grpc.Server
	config         *common.Config
	flag           *common.FlagGRPCService
	srv            discovery.DiscoveryServiceServer
}

func (g *Discovery) Start(_ context.Context) error {
	discovery.PrintASCII()
	if err := protobuf.Validate(g.config); err != nil {
		return err
	}

	discovery.RegisterDiscoveryServiceServer(g.AsServer(), g.srv)
	zlog.Debugf("discovery service started on %s", g.flag.GetAddress())

	return g.Serve(g.flag.GetAddress())
}
