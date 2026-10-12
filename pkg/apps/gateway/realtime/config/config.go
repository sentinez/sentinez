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

package config

import (
	"sync"

	realtimepb "github.com/sentinez/sentinez/api/proto/sentinez/gateway/realtime/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/pkg/apps/gateway/realtime/flags"
	"github.com/sentinez/shared/config"
)

var (
	_once    sync.Once
	_appConf *settingpb.Config
	_loadErr error
)

func Config() (*settingpb.Config, error) {
	_once.Do(func() { _appConf, _loadErr = load() })

	return _appConf, _loadErr
}

func load() (*settingpb.Config, error) {
	flag, err := flags.Parse()
	if err != nil {
		return nil, err
	}

	env, err := config.LoadEnv(flag.GetEnvFile(),
		settingpb.Senz_SENZ_HOSTNAME,
		settingpb.Senz_SENZ_ADDRESS,
		settingpb.Senz_SENZ_SECRET_KEY,
	)
	if err != nil {
		return nil, err
	}

	return &settingpb.Config{
		Meta: realtimepb.GetMetaRealtime(),
		Flag: flag,
		Env:  env,
	}, nil
}
