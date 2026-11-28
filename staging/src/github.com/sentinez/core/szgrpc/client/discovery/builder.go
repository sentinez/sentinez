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

package discovery

import (
	"context"
	"time"

	grpcresolver "google.golang.org/grpc/resolver"
)

// Scheme is the scheme of the gRPC targets resolved by a Builder.
const Scheme = "consul"

var (
	_ grpcresolver.Builder = (*Builder)(nil)
	_ Instancer            = (*Discovery)(nil)
)

// Instancer looks up the host:port of the instances of a service, e.g.
// Discovery.
type Instancer interface {
	Instances(serviceKey string) ([]string, error)
}

// Builder is the gRPC resolver of the targets consul:///<service key>
// (see Target): the healthy instances of the service in consul, looked up
// again every interval and when gRPC asks, e.g. after a connection failed.
type Builder struct {
	instancer Instancer
	interval  time.Duration
}

// NewBuilder returns the resolver of the services of instancer, e.g. the
// Discovery of a consul, looked up every interval.
func NewBuilder(instancer Instancer, interval time.Duration) *Builder {
	return &Builder{instancer: instancer, interval: interval}
}

// Target returns the gRPC target of the service of serviceKey.
func Target(serviceKey string) string {
	return Scheme + ":///" + serviceKey
}

// Scheme implements resolver.Builder.
func (b *Builder) Scheme() string {
	return Scheme
}

// Build implements resolver.Builder. Its lookups stop on Close.
func (b *Builder) Build(target grpcresolver.Target,
	cc grpcresolver.ClientConn,
	_ grpcresolver.BuildOptions) (grpcresolver.Resolver, error) {

	ctx, cancel := context.WithCancel(context.Background())
	r := &consulResolver{
		instancer:  b.instancer,
		serviceKey: target.Endpoint(),
		cc:         cc,
		cancel:     cancel,
		refresh:    make(chan struct{}, 1),
		done:       make(chan struct{}),
	}

	go r.watch(ctx, b.interval)

	return r, nil
}

// consulResolver keeps the addresses of a conn to those of the instances
// of serviceKey.
type consulResolver struct {
	instancer  Instancer
	serviceKey string
	cc         grpcresolver.ClientConn

	cancel  context.CancelFunc
	refresh chan struct{}
	done    chan struct{}
}

// ResolveNow implements resolver.Resolver: a lookup soon, at most one
// pending.
func (r *consulResolver) ResolveNow(grpcresolver.ResolveNowOptions) {
	select {
	case r.refresh <- struct{}{}:
	default:
	}
}

// Close implements resolver.Resolver: it stops the lookups and waits for
// the one running.
func (r *consulResolver) Close() {
	r.cancel()
	<-r.done
}

// watch looks the instances up now, then every interval or on ResolveNow,
// until ctx is done.
func (r *consulResolver) watch(ctx context.Context, interval time.Duration) {
	defer close(r.done)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		r.update()

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-r.refresh:
		}
	}
}

// update hands the addresses of the instances to the conn, or the error of
// the lookup: the calls then wait for, or fail without, an instance.
func (r *consulResolver) update() {
	addrs, err := r.instancer.Instances(r.serviceKey)
	if err != nil {
		r.cc.ReportError(err)
		return
	}

	endpoints := make([]grpcresolver.Endpoint, 0, len(addrs))
	for _, addr := range addrs {
		endpoints = append(endpoints, grpcresolver.Endpoint{
			Addresses: []grpcresolver.Address{{Addr: addr}},
		})
	}

	// an error asks for a new lookup (ResolveNow): the next tick does it
	_ = r.cc.UpdateState(grpcresolver.State{Endpoints: endpoints})
}
