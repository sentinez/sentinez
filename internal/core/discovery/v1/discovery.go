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
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

// make sure Discovery implement runner.Server
// it will start by runner.runner through runner.Server
var _ runner.Server = (*Discovery)(nil)

type Service struct {
	*grpcgw.Server
	handler discovery.DiscoveryServiceServer
}

func NewService() *Service {
	return &Service{
		Server:  grpcgw.NewDefault(),
		handler: dischdl.New(),
	}
}

// New creates a new discovery module.
func New(srv *Service, conf *common.Config,
	flag *common.FlagGRPCService) runner.Server {

	return &Discovery{
		Service: srv,
		config:  conf,
		flag:    flag,
	}
}

// Discovery implements DiscoveryServiceServer.
type Discovery struct {
	*Service
	config *common.Config
	flag   *common.FlagGRPCService
}

func (g *Discovery) Start(_ context.Context) error {
	discovery.PrintASCII()
	if err := protobuf.Validate(g.config); err != nil {
		return err
	}

	discovery.RegisterDiscoveryServiceServer(g.AsServer(), g.handler)
	zlog.Debugf("discovery service started on %s", g.flag.GetAddress())

	return g.Serve(g.flag.GetAddress())
}
