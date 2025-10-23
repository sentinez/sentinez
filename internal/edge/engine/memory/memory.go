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

// Package memory ...
package memory

import (
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	"github.com/sentinez/sentinez/core"
	"github.com/sentinez/sentinez/internal/edge/engine/memory/routes"
	"github.com/sentinez/sentinez/internal/edge/engine/memory/wafengine"
	"github.com/sentinez/sentinez/pkg/zlog"
)

func Initialized(origin *edgepb.Origin, appConf *common.AppConfig) {

	// routing for each tenant
	routesCache(origin)

	// WAF rulesets config
	wafEngineCache(origin, appConf)
}

func routesCache(origin *edgepb.Origin) {
	router := routes.NewRouter()
	router.Store(origin)
}

func wafEngineCache(origin *edgepb.Origin, appConf *common.AppConfig) {
	flag := core.ReqAppAttackRCE

	engine := wafengine.New()
	err := engine.Store(appConf, origin.Namespace, core.WAF4160, flag)
	if err != nil {
		zlog.Errorf("[edge] init coraza.WAF error: %v", err)
	}
}
