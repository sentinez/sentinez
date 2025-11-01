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

	realtimepb "github.com/sentinez/sentinez/api/gen/go/sentinez/realtime/v1"
	flagspb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/flags/v1"
	"github.com/sentinez/sentinez/pkg/x/flagx"
	"github.com/sentinez/sentinez/pkg/zlog"
	"github.com/spf13/pflag"
)

var onceWS sync.Once

// ParseFlag flag args for apiserver service
func Parse() *flagspb.Flag {
	onceWS.Do(func() {
		flagx.Get().EnvFile = "./cmd/realtime/.env"

		pflag.StringVar(&flagx.Get().EnvFile, flagspb.XFlag_EnvFile,
			flagx.Get().GetEnvFile(), "environment variables config file")

		flagx.Parse(realtimepb.GetMetaRealtime())

	})

	if err := flagx.Validate(flagx.Get()); err != nil {
		zlog.Fatal(err)
	}

	return flagx.Get()
}
