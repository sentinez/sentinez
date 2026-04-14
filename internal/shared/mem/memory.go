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

	corelimiter "github.com/sentinez/core/limiter"
	corers "github.com/sentinez/core/rulesets"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	"github.com/sentinez/sentinez/internal/shared/mem/ratelimiter"
	"github.com/sentinez/sentinez/internal/shared/mem/reverseproxy"
	"github.com/sentinez/sentinez/internal/shared/mem/routes"
	"github.com/sentinez/sentinez/internal/shared/mem/ruleengine"
	"github.com/sentinez/sentinez/internal/shared/mem/settings"
	"github.com/sentinez/sentinez/internal/shared/mem/wafengine"
	stdproxy "github.com/sentinez/sentinez/pkg/network/httpx/std/proxy"
	"github.com/sentinez/shared/zlog"
)

func LoadSetting(st *edgepb.Setting) {
	if err := settings.Store(st); err != nil {
		zlog.Errorf("[edge] %v", err)
	}
}

func LoadRouter() {
	settings.Visit(func(s *edgepb.Setting) bool {
		routes.Store(s.GetOrigin())
		return true
	})
}

func LoadReverseProxy() {
	settings.Visit(func(s *edgepb.Setting) bool {
		for _, routeConfig := range s.GetOrigin().GetRoutes() {
			rproxy, err := stdproxy.NewReverseProxy(routeConfig.ProxyPass)
			if err != nil {
				continue
			}

			reverseproxy.Store(routeConfig.ProxyPass, rproxy)
		}

		return true
	})
}

func LoadRateLimiter() {
	settings.Visit(func(s *edgepb.Setting) bool {
		if !s.GetSecurity().GetIsRateLimitOn() {
			zlog.Infof("[edge][limiter] ignore '%s'",
				s.GetOrigin().GetNamespace())
			return true
		}

		size, err := time.ParseDuration(s.GetSecurity().GetTimeWindow())
		if err != nil {
			zlog.Fatalf("[edge] rate limiter, parse err: %v", err)
			return true
		}

		timeout, err := time.ParseDuration(s.GetSecurity().GetTimeout())
		if err != nil {
			zlog.Fatalf("[edge] rate limiter, parse err: %v", err)
			return true
		}

		lim := corelimiter.NewRateLimiter(
			timeout,
			size,
			s.GetSecurity().GetLimit(),
		)
		ratelimiter.Store(s.GetOrigin().GetNamespace(), lim)

		return true
	})
}

func LoadWAF(appConf *confpb.Config) {
	var (
		flag = corers.ReqAppAttackRCE
	)

	settings.Visit(func(s *edgepb.Setting) bool {
		if !s.GetSecurity().GetIsWafEngineOn() {
			zlog.Infof("[edge][waf] ignore '%s'", s.GetOrigin().GetNamespace())
			return true
		}

		ns := s.GetOrigin().GetNamespace()

		rulePath := appConf.GetFlag().GetRulePath()
		err := wafengine.Store(rulePath, ns, corers.WAF4160, flag)
		if err != nil {
			zlog.Errorf("[edge] init coraza.WAF error: %v", err)
		}

		return true
	})
}

func LoadRuleBased() {
	settings.Visit(func(s *edgepb.Setting) bool {
		ruleengine.Store(
			s.GetOrigin().GetNamespace(),
			s.GetSecurity().GetExpr())

		return true
	})
}
