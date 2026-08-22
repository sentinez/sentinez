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

// Package memory ...
package memory

import (
	"context"
	"sync"
	"time"

	corehttp "github.com/sentinez/core/http"
	corelimiter "github.com/sentinez/core/limiter"
	corers "github.com/sentinez/core/rulesets"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
	"github.com/sentinez/sentinez/internal/cluster"
	"github.com/sentinez/sentinez/internal/memory/ratelimiter"
	"github.com/sentinez/sentinez/internal/memory/reverseproxy"
	"github.com/sentinez/sentinez/internal/memory/routes"
	"github.com/sentinez/sentinez/internal/memory/rules"
	"github.com/sentinez/sentinez/internal/memory/rulesets"
	"github.com/sentinez/sentinez/internal/memory/settings"
	"github.com/sentinez/sentinez/pkg/protocol"
	"github.com/sentinez/shared/zlog"
)

var store *MemStore
var once sync.Once

func NewMemStore(st *edgepb.Setting) *MemStore {
	once.Do(func() {
		store = &MemStore{
			setting:      settings.New(),
			limiter:      ratelimiter.New(),
			reverseProxy: reverseproxy.New(),
			route:        routes.New(),
			ruleBased:    rules.New(),
			rulesets:     rulesets.New(),
		}

		if err := store.setting.Store(st); err != nil {
			zlog.Errorf("memory: save distribute setting err: %v", err)
		}
	})

	return store
}

type MemStore struct {
	setting      *settings.Setting
	limiter      *ratelimiter.Limiter
	reverseProxy *reverseproxy.ReverseProxy
	route        *routes.Router
	ruleBased    *rules.RuleBased
	rulesets     *rulesets.RuleSets
}

func (m *MemStore) Start(conf *settingpb.Config) {
	cluster.Start(conf)
}

func (m *MemStore) Shutdown(ctx context.Context) {
	cluster.Shutdown(ctx)
}

func (m *MemStore) WAFRulesets() *rulesets.RuleSets {
	if m == nil {
		return nil
	}

	return m.rulesets
}

func (m *MemStore) RuleBased() *rules.RuleBased {
	if m == nil {
		return nil
	}

	return m.ruleBased
}

func (m *MemStore) Route() *routes.Router {
	if m == nil {
		return nil
	}

	return m.route
}

func (m *MemStore) ReverseProxy() *reverseproxy.ReverseProxy {
	if m == nil {
		return nil
	}

	return m.reverseProxy
}

func (m *MemStore) Limiter() *ratelimiter.Limiter {
	if m == nil {
		return nil
	}

	return m.limiter
}

func (m *MemStore) Setting() *settings.Setting {
	if m == nil {
		return nil
	}

	return m.setting
}

func (m *MemStore) LoadServer(server corehttp.Server) {
	m.setting.Visit(func(s *edgepb.Setting) bool {
		// routing for each tenant
		_ = m.LoadRouter(s)

		// load all reverse proxy for target origin
		_ = m.LoadReverseProxy(server, s)

		// rule config
		_ = m.LoadRuleBased(s)

		// rate limiter rule config
		_ = m.LoadRateLimiter(s)

		// waf rulesets config
		_ = m.LoadRulesWAF(s)

		return true
	})
}

func (m *MemStore) LoadSetting(st ...*edgepb.Setting) {
	for _, s := range st {
		if err := m.setting.Store(s); err != nil {
			zlog.Errorf("edge: %v", err)
		}
	}
}

func (m *MemStore) LoadRouter(s *edgepb.Setting) error {
	m.route.Store(s.GetServer())
	return nil
}

func (m *MemStore) LoadReverseProxy(
	server corehttp.Server, s *edgepb.Setting) error {

	for _, routeConfig := range s.GetServer().GetLocations() {
		for _, upstream := range routeConfig.GetProxyPass() {
			if _, ok := m.reverseProxy.Load(upstream.GetServer()); ok {
				continue
			}

			target, err := protocol.Upstream2Target(upstream)
			if err != nil {
				zlog.Warnf("reverse proxy: warn: %v", err)
				return err
			}

			rproxy, err := server.AcceptReverse(target)
			if err != nil {
				zlog.Errorf("reverse proxy create err: %s", err)
				return err
			}

			m.reverseProxy.Store(upstream.GetServer(), rproxy)
		}
	}

	return nil
}

func (m *MemStore) LoadRateLimiter(s *edgepb.Setting) error {
	if !s.GetSecurity().GetIsRateLimitOn() {
		zlog.Infof("edge:limiter: ignore '%s'", s.GetServer().GetName())
		return nil
	}

	size, err := time.ParseDuration(s.GetSecurity().GetTimeWindow())
	if err != nil {
		zlog.Fatalf("edge: rate limiter, parse err: %v", err)
		return err
	}

	timeout, err := time.ParseDuration(s.GetSecurity().GetTimeout())
	if err != nil {
		zlog.Fatalf("edge: rate limiter, parse err: %v", err)
		return err
	}

	lim := corelimiter.NewRateLimiter(
		timeout,
		size,
		s.GetSecurity().GetLimit(),
	)
	m.limiter.Store(s.GetServer().GetName(), lim)

	return nil
}

func (m *MemStore) LoadRulesWAF(s *edgepb.Setting) error {
	var (
		flag = corers.ReqAppAttackRCE
	)

	if !s.GetSecurity().GetIsWafEngineOn() {
		zlog.Infof("edge:waf: ignore '%s'", s.GetServer().GetName())
		return nil
	}

	ns := s.GetServer().GetName()

	err := m.rulesets.Store(ns, flag)
	if err != nil {
		zlog.Errorf("edge: init coraza.WAF error: %v", err)
		return err
	}

	return nil
}

func (m *MemStore) LoadRuleBased(s *edgepb.Setting) error {
	m.ruleBased.Store(s.GetServer().GetName(),
		s.GetSecurity().GetRuleBasedCompiled())
	return nil
}
