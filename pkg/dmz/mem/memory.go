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
	"time"

	corelimiter "github.com/sentinez/core/ratelimiter"
	corers "github.com/sentinez/core/rulesets"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/sentinez/pkg/dmz/mem/ratelimiter"
	"github.com/sentinez/sentinez/pkg/dmz/mem/reverseproxy"
	"github.com/sentinez/sentinez/pkg/dmz/mem/routes"
	"github.com/sentinez/sentinez/pkg/dmz/mem/ruleengine"
	"github.com/sentinez/sentinez/pkg/dmz/mem/settings"
	"github.com/sentinez/sentinez/pkg/dmz/mem/wafengine"
	stdproxy "github.com/sentinez/sentinez/pkg/network/httpx/std/proxy"
	"github.com/sentinez/shared/zlog"
)

func Initialized(st *edgepb.Setting, appConf *confpb.Config) {
	// save all setting for each tenant
	LoadSetting(st)

	// routing for each tenant
	LoadRouter()

	// load all reverse proxy for target origin
	LoadReverseProxy()

	// rule config
	LoadRuleBased()

	// rate limiter rule config
	LoadRateLimiter()

	// waf rulesets config
	LoadWAF(appConf)
}

func LoadSetting(st *edgepb.Setting) {
	sts := settings.New()

	if err := sts.Store(st); err != nil {
		zlog.Errorf("[edge]%v", err)
	}
}

func LoadRouter() {
	router := routes.NewRouter()

	settings.Get().Visit(func(s *edgepb.Setting) bool {
		router.Store(s.GetOrigin())
		return true
	})
}

func LoadReverseProxy() {
	reverseProxy := reverseproxy.New()
	settings.Get().Visit(func(s *edgepb.Setting) bool {
		for _, routeConfig := range s.GetOrigin().GetRoutes() {
			rproxy, err := stdproxy.NewReverseProxy(routeConfig.Target)
			if err != nil {
				continue
			}

			reverseProxy.Store(routeConfig.Target, rproxy)
		}

		return true
	})
}

func LoadRateLimiter() {
	lim := ratelimiter.New()

	settings.Get().Visit(func(s *edgepb.Setting) bool {
		if !s.GetSecurity().GetIsRateLimitOn() {
			zlog.Infof(
				"[edge][limiter] ignore '%s'", s.GetOrigin().GetNamespace())
			return true
		}

		d, err := time.ParseDuration(s.GetSecurity().GetTimeWindow())
		if err != nil {
			zlog.Fatalf("[edge] load rate limiter to mem err: %v", err)
		}

		limiter := corelimiter.NewSlidingWindow(d, s.GetSecurity().GetLimit())
		lim.Store(s.GetOrigin().GetNamespace(), limiter)
		return true
	})
}

func LoadWAF(appConf *confpb.Config) {
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

func LoadRuleBased() {
	settings.Get().Visit(func(s *edgepb.Setting) bool {
		var (
			engine = ruleengine.New()
			ns     = s.GetOrigin().GetNamespace()
		)

		engine.Store(ns, s.GetSecurity().GetExpr())
		return true
	})
}
