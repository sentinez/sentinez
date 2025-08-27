// Copyright 2025 Duc-Hung Ho.
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

// Package apps provides the app setting for apiserver service
package apps

import (
	"sync"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/spf13/pflag"
)

var onceGRPCService sync.Once

var grpcServiceFlags = &common.FlagGRPCService{
	GatewayAddress: "http://0.0.0.0:9000",
	Address:        "127.0.0.1:0",
}

// ParseFlag flag args for grpc service
func ParseFlag() *common.FlagGRPCService {
	onceGRPCService.Do(func() {
		pflag.StringVarP(&grpcServiceFlags.Address, "address", "a",
			grpcServiceFlags.GetAddress(), "host address")

		pflag.StringVar(&grpcServiceFlags.GatewayAddress, "gateway-address",
			grpcServiceFlags.GetGatewayAddress(), "gateway address")
	})

	flags.Parse(greeter.GetMetaGreeter())

	return grpcServiceFlags
}
