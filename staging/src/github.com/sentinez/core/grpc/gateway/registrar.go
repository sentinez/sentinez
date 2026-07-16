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

package grpcgateway

import (
	"context"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sentinez/sentinez/api/client/discovery"
	"github.com/sentinez/sentinez/api/client/options"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
	"github.com/sentinez/shared/cron"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceRegistrar is an interface for registering a gRPC service. Not a server
type ServiceRegistrar interface {
	Accept(context.Context, Server) error
	AcceptFromEndpoint(context.Context, Server, *settingpb.Config) error
}

type (
	RegisterFunc[T any] func(
		ctx context.Context,
		mux *runtime.ServeMux,
		server T) error

	RegisterEndpointFn func(
		ctx context.Context,
		mux *runtime.ServeMux,
		endpoint string,
		opts []grpc.DialOption) error
)

func RegisterServiceHandlerServer[T any](
	ctx context.Context,
	mux *runtime.ServeMux,
	svc T,
	fn RegisterFunc[T]) error {

	return fn(ctx, mux, svc)
}

func RegisterServiceFromEndpoint(
	ctx context.Context,
	appConf *settingpb.Config,
	mux *runtime.ServeMux,
	serviceKey string,
	fn RegisterEndpointFn,
) error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	dcvr := discovery.GetDiscovery(&options.Options{
		ConsulURL: appConf.GetEnv().GetConsulUri(),
	})

	cron.Start(ctx, time.Second*10, func() {
		srv, err := dcvr.Discover(serviceKey)
		if err != nil {
			zlog.Errorf("[httpx] discovery err=%v", err)
			return
		}

		err = fn(ctx, mux, srv.Address, opts)
		if err == nil {
			zlog.Debugf("[%s] service: %s",
				appConf.GetMeta().GetServiceName(), serviceKey)
		}

	})

	return nil
}
