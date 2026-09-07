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
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/setting/v1"
	"github.com/sentinez/sentinez/internal/cluster"
	"github.com/sentinez/sentinez/internal/memory/cdnrules"
	"github.com/sentinez/sentinez/internal/memory/ratelimiter"
	"github.com/sentinez/sentinez/internal/memory/reverseproxy"
	"github.com/sentinez/sentinez/internal/memory/routes"
	"github.com/sentinez/sentinez/internal/memory/rulebased"
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
			ruleBased:    rulebased.New(),
			rulesets:     rulesets.New(),
			cdnRules:     cdnrules.New(),
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
	ruleBased    *rulebased.RuleBased
	rulesets     *rulesets.Rulesets
	cdnRules     *cdnrules.Rule
}

func (m *MemStore) Start(conf *settingpb.Config) {
	cluster.Start(conf)
}

func (m *MemStore) Shutdown(ctx context.Context) {
	cluster.Shutdown(ctx)
}

func (m *MemStore) Rulesets() *rulesets.Rulesets {
	if m == nil {
		return nil
	}

	return m.rulesets
}

func (m *MemStore) RuleBased() *rulebased.RuleBased {
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

func (m *MemStore) CDNRules() *cdnrules.Rule {
	if m == nil {
		return nil
	}

	return m.cdnRules
}

func (m *MemStore) LoadServer(server corehttp.Server) {
	m.setting.Visit(func(s *edgepb.Setting) bool {
		// routing for each tenant
		_ = m.LoadRouter(s)

		// load all reverse proxy for target origin
		_ = m.LoadReverseProxy(server, s)

		_ = m.LoadCDNRule(s)

		// rule config
		_ = m.LoadRuleBased(s)

		// rate limiter rule config
		_ = m.LoadRateLimiter(s)

		// waf rulesets config
		_ = m.LoadRulesets(s)

		return true
	})
}

func (m *MemStore) LoadCDNRule(st ...*edgepb.Setting) error {
	for _, s := range st {
		for _, cdn := range s.GetController().GetCdn() {
			if !cdn.GetEnable() {
				continue
			}

			m.cdnRules.Store(s.GetServer().GetName(), cdn)
		}
	}

	return nil
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
	for _, limiter := range s.GetSecurity().GetLimiters() {
		if !limiter.GetEnable() {
			zlog.Infof("edge:limiter: ignore '%s'", s.GetServer().GetName())
			return nil
		}

		size, err := time.ParseDuration(limiter.GetTimeWindow())
		if err != nil {
			zlog.Fatalf("edge: rate limiter, parse err: %v", err)
			return err
		}

		timeout, err := time.ParseDuration(limiter.GetTimeout())
		if err != nil {
			zlog.Fatalf("edge: rate limiter, parse err: %v", err)
			return err
		}

		lim := corelimiter.NewRateLimiter(
			timeout,
			size,
			limiter.GetLimit(),
		)
		m.limiter.Store(s.GetServer().GetName(), lim)
	}

	return nil
}

func (m *MemStore) LoadRulesets(s *edgepb.Setting) error {
	var (
		flag = corers.ReqAppAttackRCE
	)

	for _, ruleset := range s.GetSecurity().GetRulesets() {
		if !ruleset.GetEnable() {
			zlog.Infof("edge:waf: ignore '%s'", s.GetServer().GetName())
			continue
		}

		ns := s.GetServer().GetName()

		err := m.rulesets.Store(ns, flag)
		if err != nil {
			zlog.Errorf("edge: init coraza.WAF error: %v", err)
			return err
		}
	}

	return nil
}

func (m *MemStore) LoadRuleBased(s *edgepb.Setting) error {
	for _, rule := range s.GetSecurity().GetRules() {
		if !rule.GetEnable() {
			zlog.Infof("edge:rule: ignore '%s'", s.GetServer().GetName())
			continue
		}

		m.ruleBased.Store(s.GetServer().GetName(), rule.GetIngressCompiled())
	}

	return nil
}
