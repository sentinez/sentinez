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
	"sync"
	"time"

	corehttp "github.com/sentinez/core/http"
	corelimiter "github.com/sentinez/core/limiter"
	corers "github.com/sentinez/core/rulesets"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/sentinez/internal/utils"
	"github.com/sentinez/sentinez/pkg/dmz/memory/ratelimiter"
	"github.com/sentinez/sentinez/pkg/dmz/memory/reverseproxy"
	"github.com/sentinez/sentinez/pkg/dmz/memory/routes"
	"github.com/sentinez/sentinez/pkg/dmz/memory/ruleengine"
	"github.com/sentinez/sentinez/pkg/dmz/memory/settings"
	"github.com/sentinez/sentinez/pkg/dmz/memory/wafengine"
	"github.com/sentinez/shared/zlog"
)

var (
	reverseProxyConstructor func(string) (corehttp.ReverseProxy, error)
	lock                    sync.Mutex
)

func SetReverseProxyConstructor(
	fn func(string) (corehttp.ReverseProxy, error)) {
	if fn == nil {
		return
	}

	lock.Lock()
	defer lock.Unlock()

	if reverseProxyConstructor != nil {
		return
	}

	reverseProxyConstructor = fn
}

func LoadSetting(st *edgepb.Setting) {
	if err := settings.Store(st); err != nil {
		zlog.Errorf("[edge] %v", err)
	}
}

func LoadRouter() {
	settings.Visit(func(s *edgepb.Setting) bool {
		routes.Store(s.GetServer())
		return true
	})
}

func LoadReverseProxy() {
	settings.Visit(func(s *edgepb.Setting) bool {
		for _, routeConfig := range s.GetServer().GetLocations() {
			for _, upstream := range routeConfig.GetProxyPass() {
				if _, ok := reverseproxy.Load(upstream.GetServer()); ok {
					continue
				}

				target, err := utils.Upstream2Target(upstream)
				if err != nil {
					zlog.Warnf("reverse proxy: warn: %v", err)
					continue
				}

				if reverseProxyConstructor == nil {
					zlog.Errorf("reverse proxy: constructor is not set")
					continue
				}

				rproxy, err := reverseProxyConstructor(target)
				if err != nil {
					continue
				}

				reverseproxy.Store(upstream.GetServer(), rproxy)
			}
		}

		return true
	})
}

func LoadRateLimiter() {
	settings.Visit(func(s *edgepb.Setting) bool {
		if !s.GetSecurity().GetIsRateLimitOn() {
			zlog.Infof("[edge][limiter] ignore '%s'",
				s.GetServer().GetName())
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
		ratelimiter.Store(s.GetServer().GetName(), lim)

		return true
	})
}

func LoadWAF(appConf *confpb.Config) {
	var (
		flag = corers.ReqAppAttackRCE
	)

	settings.Visit(func(s *edgepb.Setting) bool {
		if !s.GetSecurity().GetIsWafEngineOn() {
			zlog.Infof("[edge][waf] ignore '%s'", s.GetServer().GetName())
			return true
		}

		ns := s.GetServer().GetName()

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
			s.GetServer().GetName(),
			s.GetSecurity().GetRuleBasedCompiled())

		return true
	})
}
