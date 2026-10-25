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
	"fmt"
	"sync"

	"github.com/sentinez/core/szrunner"
	realtimepb "github.com/sentinez/sentinez/api/proto/sentinez/gateway/realtime/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/szflag"
	"github.com/spf13/pflag"
)

var _onceWS sync.Once

// ParseFlag flag args for apiserver service
func Parse() (*settingpb.Flag, error) {
	_onceWS.Do(func() {
		szflag.Get().EnvFile = "./cmd/szrealtime/.env"

		pflag.StringVar(&szflag.Get().EnvFile, settingpb.XFlag_EnvFile,
			szflag.Get().GetEnvFile(), "environment variables config file")

		szflag.Parse(szrunner.Banner(realtimepb.GetMetaRealtime()))

	})

	if err := szflag.Validate(szflag.Get()); err != nil {
		return nil, fmt.Errorf("validate flags: %w", err)
	}

	return szflag.Get(), nil
}
