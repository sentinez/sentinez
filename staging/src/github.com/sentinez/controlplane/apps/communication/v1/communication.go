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

// Package communication implements the Communication service server.
package communication

import (
	"context"

	communicationfac "github.com/sentinez/controlplane/apps/communication/v1/factory"
	"github.com/sentinez/core/szgrpc"
	communicationpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/communication/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"google.golang.org/grpc"
)

// Communication implements CommunicationServiceServer.
type Communication struct {
	*szgrpc.Server
	handler communicationpb.CommunicationServiceServer
	conf    *settingpb.Config
}

// NewService builds the communication server; opts are the options of the
// gRPC server, at least the authorization of interceptor.ServerOptions:
// without it every method is open.
func NewService(ctx context.Context, conf *settingpb.Config,
	opts []grpc.ServerOption) (*Communication, error) {
	handler, err := communicationfac.NewDefaultHandler(ctx, conf)
	if err != nil {
		return nil, err
	}

	return &Communication{
		Server:  szgrpc.New(opts...),
		handler: handler,
		conf:    conf,
	}, nil
}

// Start serves the communication service, registered in consul; opts are
// applied after the defaults, e.g. szgrpc.WithAddress.
func (c *Communication) Start(opts ...szgrpc.ServerOption) error {
	communicationpb.RegisterCommunicationServiceServer(
		c.AsServer(), c.handler,
	)

	opts = append(
		[]szgrpc.ServerOption{szgrpc.WithDiscorvery(true)}, opts...,
	)

	return c.Serve(c.conf, opts...)
}
