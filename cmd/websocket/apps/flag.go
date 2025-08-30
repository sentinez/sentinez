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

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/ws/v1"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"github.com/spf13/pflag"
)

var onceWS sync.Once

// ParseFlag flag args for apiserver service
func ParseFlag() *common.Flag {
	onceWS.Do(func() {
		flags.Get().EnvFile = "./cmd/websocket/.env"

		pflag.StringVar(&flags.Get().EnvFile, "env-file",
			flags.Get().GetEnvFile(), "environment variables config file")

		flags.Parse(ws.GetMetaWs())

	})

	if err := flags.Validate(flags.Get()); err != nil {
		zlog.Fatal(err)
	}

	return flags.Get()
}
