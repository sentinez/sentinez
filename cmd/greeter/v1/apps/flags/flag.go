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

// Package flags provides the app setting for apiserver service
package flags

import (
	"sync"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/stdcmn/zflag"
	"github.com/sentinez/sentinez/pkg/stdcmn/zlog"
	"github.com/spf13/pflag"
)

var onceGRPCService sync.Once

// ParseFlag flag args for grpc service
func Parse() *common.Flag {
	onceGRPCService.Do(func() {
		zflag.Get().EnvFile = "./cmd/greeter/v1/.env"

		pflag.StringVar(&zflag.Get().EnvFile, "env-file",
			zflag.Get().GetEnvFile(), "environment variables config file")

		zflag.Parse(greeter.GetMetaGreeter())

	})

	if err := zflag.Validate(zflag.Get()); err != nil {
		zlog.Fatal(err)
	}

	return zflag.Get()
}
