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

// Package cache ...
package cache

import (
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	edgeyaml "github.com/sentinez/sentinez/cmd/edge/v1/apps/yaml"
	"github.com/sentinez/sentinez/internal/edge/v1/cache/routes"
	wafcache "github.com/sentinez/sentinez/internal/edge/v1/cache/waf"
	"github.com/sentinez/sentinez/pkg/zlog"
	"github.com/sentinez/sentinez/rules"
)

func Initialized(edgeYaml *edgeyaml.Config, appConf *common.AppConfig) {

	routesCache(edgeYaml)

	wafCache(edgeYaml, appConf)
}

func wafCache(edgeYaml *edgeyaml.Config, appConf *common.AppConfig) {
	flag := rules.ReqAppAttackRCE

	wafCached := wafcache.New()
	for _, proxy := range edgeYaml.ReverseProxies {

		err := wafCached.Store(appConf, proxy.Namespace, rules.Ver4_16_0, flag)
		if err != nil {
			zlog.Errorf("[edge] init coraza.WAF error: %v", err)
		}
	}
}

func routesCache(edgeYaml *edgeyaml.Config) {
	router := routes.NewRouter()
	router.Store(edgeYaml)
}
