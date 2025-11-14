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

// Package mem ...
package mem

import (
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	corers "github.com/sentinez/sentinez/core/rulesets"
	"github.com/sentinez/sentinez/pkg/dmz/mem/routes"
	"github.com/sentinez/sentinez/pkg/dmz/mem/settings"
	"github.com/sentinez/sentinez/pkg/dmz/mem/wafengine"
	"github.com/sentinez/sentinez/shared/zlog"
)

func Initialized(st *edgepb.Setting, appConf *confpb.Config) {
	// save all setting for each tenant
	loadSetting(st)

	// routing for each tenant
	loadRouting()

	// WAF rulesets config
	loadWAF(appConf)
}

func loadSetting(st *edgepb.Setting) {
	sts := settings.New()

	if err := sts.Store(st); err != nil {
		zlog.Errorf("[edge]%v", err)
	}
}

func loadRouting() {
	router := routes.NewRouter()

	settings.Get().Visit(func(s *edgepb.Setting) bool {
		router.Store(s.GetOrigin())
		return true
	})
}

func loadWAF(appConf *confpb.Config) {
	settings.Get().Visit(func(s *edgepb.Setting) bool {
		if !s.GetSecurity().GetIsWafEngineOn() {
			zlog.Infof("[edge][waf] ignore '%s'", s.GetOrigin().GetNamespace())
			return true
		}

		var (
			ns     = s.GetOrigin().GetNamespace()
			engine = wafengine.New()
			flag   = corers.ReqAppAttackRCE
		)

		err := engine.Store(appConf, ns, corers.WAF4160, flag)
		if err != nil {
			zlog.Errorf("[edge] init coraza.WAF error: %v", err)
		}
		return true
	})
}
