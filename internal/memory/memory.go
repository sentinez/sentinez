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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szlimiter"
	"github.com/sentinez/core/szrule"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/memory/cdnrules"
	"github.com/sentinez/sentinez/internal/memory/ratelimiter"
	"github.com/sentinez/sentinez/internal/memory/reverseproxy"
	"github.com/sentinez/sentinez/internal/memory/routes"
	"github.com/sentinez/sentinez/internal/memory/rulesets"
	"github.com/sentinez/sentinez/internal/memory/secrules"
	"github.com/sentinez/sentinez/internal/memory/settings"
	"github.com/sentinez/sentinez/pkg/cluster"
	"github.com/sentinez/sentinez/pkg/protocol"
	"github.com/sentinez/shared/szlog"
)

var (
	_store  *MemStore
	_lock   sync.Mutex
	_shared *cluster.Cluster
)

func NewMemStore(conf *settingpb.Config) (*MemStore, error) {
	_lock.Lock()
	defer _lock.Unlock()

	if _store != nil {
		return _store, nil
	}

	_shared, err := cluster.NewCluster(conf)
	if err != nil {
		return nil, err
	}

	_shared.Start()

	_store = &MemStore{
		setting:      settings.New(_shared),
		limiter:      ratelimiter.New(),
		reverseProxy: reverseproxy.New(),
		route:        routes.New(),
		secRule:      secrules.New(),
		rulesets:     rulesets.New(),
		cdnRules:     cdnrules.New(),
	}

	return _store, nil
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

func (m *MemStore) Shutdown(ctx context.Context) error {
	if _shared == nil {
		return nil
	}

	return _shared.Shutdown(ctx)
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

func (m *MemStore) LoadServer(server szhttp.Server) {
	m.setting.Visit(func(s *edgepb.Setting) bool {
		err := errors.Join(
			m.LoadRouter(s),
			m.LoadReverseProxy(server, s),
			m.LoadCDNRule(s),
			m.LoadSecRule(s),
			m.LoadRateLimiter(s),
			m.LoadRulesets(s),
		)
		if err != nil {
			szlog.Errorf("edge: load '%s': %v", s.GetServer().GetName(), err)
		}

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
			szlog.Errorf("edge: %v", err)
		}
	}
}

func (m *MemStore) LoadRouter(s *edgepb.Setting) error {
	m.route.Store(s.GetServer())
	return nil
}

func (m *MemStore) LoadReverseProxy(
	server szhttp.Server, s *edgepb.Setting) error {

	for _, routeConfig := range s.GetServer().GetLocations() {
		for _, upstream := range routeConfig.GetProxyPass() {
			if _, ok := m.reverseProxy.Load(upstream.GetServer()); ok {
				continue
			}

			target, err := protocol.Upstream2Target(upstream)
			if err != nil {
				return fmt.Errorf("reverse proxy: target: %w", err)
			}

			rproxy, err := server.AcceptReverse(target)
			if err != nil {
				return fmt.Errorf("reverse proxy: create: %w", err)
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
			szlog.Infof("edge:limiter: ignore '%s'", s.GetServer().GetName())
			continue
		}

		entry, err := newLimiterEntry(l)
		if err != nil {
			return fmt.Errorf("rate limiter: parse: %w", err)
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
		return fmt.Errorf("rulesets: init coraza waf: %w", err)
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
		Eval:    szrule.NewEval(rule.GetExpr()),
		Rule:    rule,
		Limiter: szlimiter.NewRateLimiter(timeout, size, l.GetMaxRequests()),
	}, nil
}
