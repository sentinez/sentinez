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

package szgrpc

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/sentinez/sentinez/api/client/discovery"
	"github.com/sentinez/sentinez/api/client/options"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/szcron"
	"github.com/sentinez/shared/szlog"
	"github.com/sony/gobreaker"
)

// nolint:funlen
// Register registers the service in consul and sends heartbeats until ctx
// is done.
func Register(ctx context.Context, name string, conf *settingpb.Config) {
	addr, port, err := net.SplitHostPort(conf.Get(settingpb.Senz_SENZ_ADDRESS))
	if err != nil {
		szlog.Errorf("failed to split address: %v", err)
		return
	}

	portInt, _ := strconv.Atoi(port)
	dcvr := discovery.GetDiscovery(
		&options.Options{ConsulURL: conf.Get(settingpb.Senz_SENZ_CONSUL_URI)})

	var (
		serviceId = ""
		timeout   = 10 * time.Second
		ttl       = 15 * time.Second
	)

	cb := circuitBreaker(timeout)
	for {
		if state := cb.State(); state == gobreaker.StateOpen {
			szlog.Warnf("breaker is OPEN - waiting %v before retry...", timeout)
			if !sleep(ctx, timeout) {
				return
			}

			continue
		}

		if _, err := cb.Execute(func() (any, error) {
			id, err := dcvr.Register(&discovery.RegisterRequest{
				Name: name, Address: addr, Port: portInt, TTL: ttl})
			if err == nil {
				serviceId = id
			}

			return nil, err

		}); err != nil {
			szlog.Errorf("failed to register service: %v, retrying...", err)
			if !sleep(ctx, time.Second) {
				return
			}

			continue
		}
		break
	}

	<-startCron(ctx, cb, dcvr, serviceId, timeout)
}

// sleep waits for d, it reports false when ctx is done first.
func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func startCron(
	ctx context.Context,
	cb *gobreaker.CircuitBreaker,
	d *discovery.Discovery,
	serviceId string,
	timeout time.Duration,
) <-chan struct{} {
	// Periodic heartbeat, also protected by breaker
	return szcron.Start(ctx, timeout, func() {
		if _, err := cb.Execute(func() (any, error) {
			return nil, d.Heartbeat(serviceId)

		}); err != nil {
			szlog.Warnf(
				"heartbeat failed for %s: %v (breaker: %s)",
				serviceId,
				err,
				cb.State(),
			)
		}
	})
}

func circuitBreaker(timeout time.Duration) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "grpcRegister",
		MaxRequests: 1,
		Interval:    30 * time.Second, // reset counts every 30s
		Timeout:     timeout,          // time before trying again after open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Open circuit if more than 3 consecutive failures
			return counts.ConsecutiveFailures >= 1
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			szlog.Infof(
				"CircuitBreaker[%s] state changed: %s > %s", name, from, to)
		},
	})
}
