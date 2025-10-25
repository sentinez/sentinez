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
	"github.com/sentinez/sentinez/internal/edge/pkgs/memory/routes"
	"github.com/sentinez/sentinez/internal/edge/pkgs/memory/settings"
	"github.com/sentinez/sentinez/internal/edge/pkgs/memory/wafengine"
	"github.com/sentinez/sentinez/pkg/zlog"
)

func Initialized(st *edgepb.Setting, appConf *common.AppConfig) {
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

func loadWAF(appConf *common.AppConfig) {
	settings.Get().Visit(func(s *edgepb.Setting) bool {
		var (
			ns     = s.GetOrigin().GetNamespace()
			engine = wafengine.New()
			flag   = core.ReqAppAttackRCE
		)

		if err := engine.Store(appConf, ns, core.WAF4160, flag); err != nil {
			zlog.Errorf("[edge] init coraza.WAF error: %v", err)
		}
		return true
	})
}
