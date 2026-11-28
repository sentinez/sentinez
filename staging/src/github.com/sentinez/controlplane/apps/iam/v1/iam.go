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

// Package iam provides the Identity Access Management service.
package iam

import (
	"context"

	iamfac "github.com/sentinez/controlplane/apps/iam/v1/factory"
	"github.com/sentinez/core/szgrpc"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"google.golang.org/grpc"
)

// IAM serves IdentityAccessManagementServiceServer over gRPC.
type IAM struct {
	*szgrpc.Server
	hdl  iampb.IdentityAccessManagementServiceServer
	conf *settingpb.Config
}

// NewService creates the IAM gRPC server.
// opts are the options of the gRPC server, at least the authorization
// of interceptor.ServerOptions: without it every method is open.
func NewService(ctx context.Context, conf *settingpb.Config,
	opts []grpc.ServerOption) *IAM {
	return &IAM{
		Server: szgrpc.New(opts...),
		hdl:    iamfac.NewDefaultHandler(ctx, conf),
		conf:   conf,
	}
}

// Start serves IAM, registered in consul; opts are applied after the
// defaults, e.g. szgrpc.WithAddress.
func (im *IAM) Start(opts ...szgrpc.ServerOption) error {
	iampb.RegisterIdentityAccessManagementServiceServer(im.AsServer(), im.hdl)

	opts = append([]szgrpc.ServerOption{szgrpc.WithDiscorvery(true)}, opts...)

	return im.Serve(im.conf, opts...)
}
