// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dataplane

import (
	"context"
	"fmt"

	"github.com/sentinez/core/szgrpc"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/sentinez/internal/dmz/dataplane/driver"
	"github.com/sentinez/sentinez/pkg/network"
	"github.com/sentinez/shared/szlog"
)

func New(conf *settingpb.Config) *Server {
	return &Server{
		conf: conf,
		grpc: szgrpc.New(),
	}
}

type Server struct {
	grpc *szgrpc.Server
	conf *settingpb.Config
}

const VETH0 = "veth0"

func (s *Server) Start(networkInterface string) error {
	ctx := driver.NewContext()

	iface, err := network.GetInterface(networkInterface)
	if err != nil {
		return fmt.Errorf("stream: getting interface: %w", err)
	}

	if err := ctx.AttachXDP(iface.Index); err != nil {
		return fmt.Errorf("stream: attaching xdp: %w", err)
	}

	szlog.Infof("stream: attaching xdp successfully")

	return s.grpc.Serve(s.conf,
		szgrpc.WithDiscorvery(false),
		szgrpc.WithAddress(s.conf.GetDefault(
			settingpb.Senz_SENZ_ADDRESS, defaults.DataplaneAddress)),
	)
}

func (s *Server) Stop(ctx context.Context) error {
	_ = driver.CloseContext()
	return s.grpc.Shutdown(ctx)
}
