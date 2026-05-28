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

package coregrpc

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/sentinez/sentinez/api/client/discovery"
	"github.com/sentinez/sentinez/api/client/options"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/shared/cron"
	"github.com/sentinez/shared/zlog"
	"github.com/sony/gobreaker"
)

// nolint:funlen
func Register(name string, conf *confpb.EnvConfig) {
	addr, port, err := net.SplitHostPort(conf.GetGrpcAddress())
	if err != nil {
		zlog.Errorf("failed to split address: %v", err)
		return
	}

	portInt, _ := strconv.Atoi(port)
	dcvr := discovery.GetDiscovery(
		&options.Options{ConsulURL: conf.GetConsulUri()})

	var (
		serviceId = ""
		timeout   = 10 * time.Second
		ttl       = 15 * time.Second
	)

	cb := circuitBreaker(timeout)
	for {
		if state := cb.State(); state == gobreaker.StateOpen {
			zlog.Warnf("breaker is OPEN - waiting %v before retry...", timeout)
			time.Sleep(timeout)

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
			zlog.Errorf("failed to register service: %v, retrying...", err)
			time.Sleep(time.Second * 1)

			continue
		}
		break
	}

	startCron(cb, dcvr, serviceId, timeout)
}

func startCron(
	cb *gobreaker.CircuitBreaker,
	d *discovery.Discovery,
	serviceId string,
	timeout time.Duration,
) {
	// Periodic heartbeat, also protected by breaker
	cron.Start(context.Background(), timeout, func() {
		if _, err := cb.Execute(func() (any, error) {
			return nil, d.Heartbeat(serviceId)

		}); err != nil {
			zlog.Warnf(
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
			zlog.Infof(
				"CircuitBreaker[%s] state changed: %s > %s", name, from, to)
		},
	})
}
