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

	coregrpc "github.com/sentinez/core/grpc"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/sentinez/internal/dmz/dataplane/driver"
	"github.com/sentinez/sentinez/pkg/network"
	"github.com/sentinez/shared/zlog"
)

func New(conf *confpb.Config) *Server {
	return &Server{
		conf: conf,
		grpc: coregrpc.New(coregrpc.WithXMeta(conf.GetMeta())),
	}
}

type Server struct {
	grpc *coregrpc.Server
	conf *confpb.Config
}

const VETH0 = "veth0"

func (s *Server) Start(networkInterface string) error {
	ctx := driver.NewContext()

	iface, err := network.GetInterface(networkInterface)
	if err != nil {
		zlog.Errorf("stream: getting interface: %v", err)
		return err
	}

	if err := ctx.AttachXDP(iface.Index); err != nil {
		zlog.Errorf("stream: attaching xdp: %v", err)
		return err
	}

	zlog.Infof("stream: attaching xdp successfully")

	return s.grpc.Serve(s.conf)
}

func (s *Server) Stop(ctx context.Context) error {
	_ = driver.CloseContext()
	return s.grpc.Shutdown(ctx)
}
