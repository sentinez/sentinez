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
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	edgeyaml "github.com/sentinez/sentinez/cmd/edge/v1/apps/yaml"
	"github.com/sentinez/sentinez/core"
	"github.com/sentinez/sentinez/internal/shared/memory/routes"
	wafcache "github.com/sentinez/sentinez/internal/shared/memory/waf"
	"github.com/sentinez/sentinez/pkg/zlog"
)

func Initialized(edgeYaml *edgeyaml.Config, appConf *common.AppConfig) {

	// routing for each tenant
	routesCache(edgeYaml)

	// WAF rulesets config
	wafRuleCache(edgeYaml, appConf)
}

func routesCache(edgeYaml *edgeyaml.Config) {
	router := routes.NewRouter()
	router.Store(edgeYaml)
}

func wafRuleCache(edgeYaml *edgeyaml.Config, appConf *common.AppConfig) {
	flag := core.ReqAppAttackRCE

	cache := wafcache.New()
	for _, proxy := range edgeYaml.ReverseProxies {

		err := cache.Store(appConf, proxy.Namespace, core.WAF4160, flag)
		if err != nil {
			zlog.Errorf("[edge] init coraza.WAF error: %v", err)
		}
	}
}
