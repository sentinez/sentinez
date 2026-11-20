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
	"errors"
	"fmt"
	"sync"

	"github.com/sentinez/core/szgrpc"
	"github.com/sentinez/pkg/szlog"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/sentinez/internal/dmz/dataplane/driver"
	"github.com/sentinez/sentinez/pkg/network"
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

	mu  sync.Mutex
	xdp *driver.Context // nil until Start attached it
}

const VETH0 = "veth0"

// Start attaches the XDP program to networkInterface, then serves gRPC.
func (s *Server) Start(networkInterface string) error {
	if err := s.attach(networkInterface); err != nil {
		return err
	}

	return s.grpc.Serve(s.conf,
		szgrpc.WithDiscorvery(false),
		szgrpc.WithAddress(s.conf.GetDefault(
			settingpb.Sz_SZ_ADDRESS, defaults.DataplaneAddress)),
	)
}

// Stop shuts gRPC down, and detaches and unloads the XDP program.
func (s *Server) Stop(ctx context.Context) error {
	err := s.grpc.Shutdown(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.xdp != nil {
		err = errors.Join(err, s.xdp.Close())
		s.xdp = nil
	}

	return err
}

func (s *Server) attach(networkInterface string) error {
	iface, err := network.GetInterface(networkInterface)
	if err != nil {
		return fmt.Errorf("dataplane: interface: %w", err)
	}

	xdp, err := driver.NewContext()
	if err != nil {
		return err
	}

	if err := xdp.AttachXDP(iface.Index); err != nil {
		return errors.Join(err, xdp.Close())
	}

	s.mu.Lock()
	s.xdp = xdp
	s.mu.Unlock()

	szlog.Infof("dataplane: xdp attached to %s", networkInterface)

	return nil
}
