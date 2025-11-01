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

package runner

import (
	"context"

	configspb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/configs/v1"
)

type runnerAppConfKey string

const runnerAppConf runnerAppConfKey = "RunnerContextValue"

func newContext(appConf *configspb.AppConfig) context.Context {
	return context.WithValue(context.Background(), runnerAppConf, appConf)
}

func GetAppConfig(ctx context.Context) *configspb.AppConfig {
	val := ctx.Value(runnerAppConf)
	if val == nil {
		return nil
	}

	return val.(*configspb.AppConfig)
}
