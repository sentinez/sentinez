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
	"fmt"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sentinez/core/szgrpc/client/connection"
	"github.com/sentinez/core/szgrpc/client/discovery"
	"github.com/sentinez/core/szgrpc/validate"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"google.golang.org/grpc"
)

// _resolveInterval is the interval of the lookups of the nodes of a
// service in consul.
const _resolveInterval = 10 * time.Second

// _roundRobin spreads the calls over the nodes of a service.
const _roundRobin = `{"loadBalancingConfig": [{"round_robin": {}}]}`

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

// RegisterServiceFromEndpoint registers on mux, once, the service of
// serviceKey served by the nodes registered in consul (SZ_CONSUL_URI). Its
// conn resolves them with discovery.Builder, balances the calls over them
// and is closed when ctx is done; a call without a node fails with
// codes.Unavailable. With SZ_TLS_* set the conn is the internal mTLS, the
// nodes verified as serverName (an identity), else plaintext.
func RegisterServiceFromEndpoint(
	ctx context.Context,
	appConf *settingpb.Config,
	mux *runtime.ServeMux,
	serviceKey, serverName string,
	fn RegisterEndpointFn,
) error {
	dcvr, err := discovery.GetDiscovery(
		discovery.WithConsulURL(appConf.Get(settingpb.Sz_SZ_CONSUL_URI)))
	if err != nil {
		return fmt.Errorf("register %s: %w", serviceKey, err)
	}

	opts, err := dialOptions(connection.TLSFromConfig(appConf), serverName,
		discovery.NewBuilder(dcvr, _resolveInterval))
	if err != nil {
		return fmt.Errorf("register %s: %w", serviceKey, err)
	}

	return fn(ctx, mux, discovery.Target(serviceKey), opts)
}

// dialOptions returns the options of a conn to the nodes resolved by
// builder: tls verifying them as serverName, the calls balanced over them
// and their requests validated.
func dialOptions(tls connection.TLS, serverName string,
	builder *discovery.Builder) ([]grpc.DialOption, error) {
	opts, err := tls.DialOptions(serverName)
	if err != nil {
		return nil, err
	}

	return append(opts,
		grpc.WithResolvers(builder),
		grpc.WithDefaultServiceConfig(_roundRobin),
		grpc.WithChainUnaryInterceptor(validate.UnaryClientInterceptor()),
	), nil
}
