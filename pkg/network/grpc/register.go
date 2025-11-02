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

package grpc

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/sentinez/sentinez/api/client/discovery"
	"github.com/sentinez/sentinez/api/client/options"
	configspb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/configs/v1"
	"github.com/sentinez/sentinez/pkg/common/cron"
	"github.com/sentinez/sentinez/pkg/zlog"
)

func Register(name string, conf *configspb.EnvConfig) {
	addr, port, err := net.SplitHostPort(conf.GetGrpcAddress())
	if err != nil {
		zlog.Errorf("failed to split address: %v", err)
		return
	}

	dcvr := discovery.GetDiscovery(&options.Options{
		ConsulURL: conf.GetConsulUri(),
	})
	portInt, _ := strconv.Atoi(port)
	serviceID := ""

	for {
		serviceID, err = dcvr.Register(&discovery.RegisterRequest{
			Name:    name,
			Address: addr,
			Port:    portInt,
			TTL:     time.Second * 15,
		})
		if err != nil {
			zlog.Errorf("failed to register service: %v, retrying...", err)
			time.Sleep(time.Second * 5)
			continue
		}

		break
	}

	cron.Start(context.Background(), 10*time.Second, func() {
		_ = dcvr.Heartbeat(serviceID)
	})
}
