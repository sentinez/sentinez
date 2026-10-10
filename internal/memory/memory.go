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
	"fmt"
	"sync"
	"time"

	corehttp "github.com/sentinez/core/http"
	corelimiter "github.com/sentinez/core/limiter"
	corerule "github.com/sentinez/core/rules"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/cluster"
	"github.com/sentinez/sentinez/internal/memory/cdnrules"
	"github.com/sentinez/sentinez/internal/memory/ratelimiter"
	"github.com/sentinez/sentinez/internal/memory/reverseproxy"
	"github.com/sentinez/sentinez/internal/memory/routes"
	"github.com/sentinez/sentinez/internal/memory/rulesets"
	"github.com/sentinez/sentinez/internal/memory/secrules"
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
			secRule:      secrules.New(),
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
	secRule      *secrules.SecRule
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

func (m *MemStore) SecRule() *secrules.SecRule {
	if m == nil {
		return nil
	}

	return m.secRule
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
		_ = m.LoadSecRule(s)

		// rate limiter rule config
		_ = m.LoadRateLimiter(s)

		// waf rulesets config
		_ = m.LoadRulesets(s)

		return true
	})
}

func (m *MemStore) LoadCDNRule(st ...*edgepb.Setting) error {
	for _, s := range st {
		for _, cdn := range s.GetDelivery().GetCdn() {
			if cdn.GetRuleRuntime().GetStatus() != typepb.Status_STATUS_ACTIVE {
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
	limiters := s.GetSecurity().GetLimiters()
	entries := make([]ratelimiter.Entry, 0, len(limiters))

	for _, l := range limiters {
		if l.GetIngressRuntime().GetStatus() != typepb.Status_STATUS_ACTIVE {
			zlog.Infof("edge:limiter: ignore '%s'", s.GetServer().GetName())
			continue
		}

		entry, err := newLimiterEntry(l)
		if err != nil {
			zlog.Fatalf("edge: rate limiter, parse err: %v", err)
			return err
		}

		entries = append(entries, entry)
	}

	m.limiter.Store(s.GetServer().GetName(), entries)

	return nil
}

func (m *MemStore) LoadRulesets(s *edgepb.Setting) error {
	ns := s.GetServer().GetName()

	err := m.rulesets.Store(ns, s.GetSecurity().GetCoreRuleset())
	if err != nil {
		zlog.Errorf("edge: init coraza.WAF error: %v", err)
		return err
	}

	return nil
}

func (m *MemStore) LoadSecRule(s *edgepb.Setting) error {
	rules := make([]*secrulepb.SecRule, 0, len(s.GetSecurity().GetSecRules()))
	for _, rule := range s.GetSecurity().GetSecRules() {
		rules = append(rules, rule.GetIngressRuntime())
	}

	m.secRule.Store(s.GetServer().GetName(), rules)

	return nil
}

func newLimiterEntry(l *edgepb.RateLimit) (ratelimiter.Entry, error) {
	size, err := time.ParseDuration(l.GetTimeWindow())
	if err != nil {
		return ratelimiter.Entry{}, fmt.Errorf("time window: %w", err)
	}

	timeout, err := time.ParseDuration(l.GetTimeout())
	if err != nil {
		return ratelimiter.Entry{}, fmt.Errorf("timeout: %w", err)
	}

	rule := l.GetIngressRuntime()

	return ratelimiter.Entry{
		Eval:    corerule.NewEval(rule.GetExpr()),
		Rule:    rule,
		Limiter: corelimiter.NewRateLimiter(timeout, size, l.GetMaxRequests()),
	}, nil
}
