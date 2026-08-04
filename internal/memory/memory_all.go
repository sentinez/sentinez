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

package memory

import (
	corehttp "github.com/sentinez/core/http"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
)

func LoadConfiguration(
	server corehttp.Server, st *edgepb.Setting, appConf *settingpb.Config) {

	// save all setting for each tenant
	LoadSetting(st)

	// routing for each tenant
	LoadRouter()

	// load all reverse proxy for target origin
	LoadReverseProxy(server)

	// rule config
	LoadRuleBased()

	// rate limiter rule config
	LoadRateLimiter()

	// waf rulesets config
	LoadWAF(appConf)
}
