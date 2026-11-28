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

// Package security provides the Security service.
package security

import (
	"context"

	securityfac "github.com/sentinez/controlplane/apps/security/v1/factory"
	"github.com/sentinez/core/szgrpc"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"google.golang.org/grpc"
)

// Security serves SecurityServiceServer over gRPC.
type Security struct {
	*szgrpc.Server
	hdl  securitypb.SecurityServiceServer
	conf *settingpb.Config
}

// NewService creates the security gRPC server.
// opts are the options of the gRPC server, at least the authorization
// of interceptor.ServerOptions: without it every method is open.
func NewService(
	ctx context.Context, conf *settingpb.Config,
	opts []grpc.ServerOption,
) (*Security, error) {
	hdl, err := securityfac.NewDefaultHandler(ctx, conf)
	if err != nil {
		return nil, err
	}

	return &Security{
		Server: szgrpc.New(opts...),
		hdl:    hdl,
		conf:   conf,
	}, nil
}

// Start serves security, registered in consul; opts are applied after the
// defaults, e.g. szgrpc.WithAddress.
func (mod *Security) Start(opts ...szgrpc.ServerOption) error {
	securitypb.RegisterSecurityServiceServer(mod.AsServer(), mod.hdl)

	opts = append([]szgrpc.ServerOption{szgrpc.WithDiscorvery(true)}, opts...)

	return mod.Serve(mod.conf, opts...)
}
