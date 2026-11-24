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
	"crypto/tls"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szlimiter"
	"github.com/sentinez/core/szrule"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szsync"
	"github.com/sentinez/scripting"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/memory/cdnrules"
	"github.com/sentinez/sentinez/internal/memory/certificates"
	"github.com/sentinez/sentinez/internal/memory/routes"
	"github.com/sentinez/sentinez/internal/memory/scripts"
	"github.com/sentinez/sentinez/internal/memory/secratelimits"
	"github.com/sentinez/sentinez/internal/memory/secrules"
	"github.com/sentinez/sentinez/internal/memory/secrulesets"
	"github.com/sentinez/sentinez/internal/memory/settings"
	"github.com/sentinez/sentinez/internal/memory/upstream"
	"github.com/sentinez/sentinez/pkg/cluster"
)

var (
	_store  *MemStore
	_lock   sync.Mutex
	_shared *cluster.Cluster

	errLimiterNoExpr = errors.New("missing ingress expr or script")
	errLimiterBoth   = errors.New("ingress expr and script are exclusive")
)

func NewMemStore(conf *settingpb.Config) (*MemStore, error) {
	_lock.Lock()
	defer _lock.Unlock()

	if _store != nil {
		return _store, nil
	}

	sealer, err := newSealer(conf)
	if err != nil {
		return nil, err
	}

	shared, err := cluster.NewCluster(conf)
	if err != nil {
		return nil, err
	}

	// assigned, not declared: Shutdown stops the global cluster
	_shared = shared

	_shared.Start()

	_store = newMemStore(shared, sealer)

	return _store, nil
}

func newMemStore(shared *cluster.Cluster, sealer *crypto.Sealer,
) *MemStore {
	return &MemStore{
		synced:       newSettingSync(shared, _settingsMap),
		central:      newSettingSync(shared, _centralMap),
		owners:       make(map[string]string),
		locals:       szsync.NewMap[string, *edgepb.Setting](),
		setting:      settings.New(),
		limiter:      secratelimits.New(),
		revProxy:     upstream.New(),
		plainProxy:   szsync.NewMap[string, szhttp.ReverseProxy](),
		passthrough:  szsync.NewMap[string, []string](),
		route:        routes.New(),
		secRule:      secrules.New(),
		scripts:      scripts.New(),
		rulesets:     secrulesets.New(),
		cdnRules:     cdnrules.New(),
		certificates: certificates.New(),
		sealer:       sealer,
	}
}

type MemStore struct {
	// synced holds the settings of every edge of the cluster, by
	// resource id; each change rebuilds the stores below (sync.go)
	synced *settingSync
	// central holds the settings built from centraldata, by resource id;
	// the leader writes it through CentralSink (central.go)
	central *settingSync
	server  szhttp.Server
	// owners maps a namespace to the resource id that holds it
	ownersMu sync.Mutex
	owners   map[string]string
	// locals the settings of this edge's file that are not synced, by
	// namespace: loaded back when centraldata releases the namespace
	locals *szsync.Map[string, *edgepb.Setting]

	setting  *settings.Setting
	limiter  *secratelimits.Limiter
	revProxy *upstream.ReverseProxy
	// plainProxy the proxies without upstream TLS settings, by target:
	// shared by every namespace dialing that target
	plainProxy *szsync.Map[string, szhttp.ReverseProxy]
	// passthrough the upstream addresses of the namespaces in passthrough
	// mode
	passthrough  *szsync.Map[string, []string]
	route        *routes.Router
	secRule      *secrules.SecRule
	scripts      *scripts.Scripts
	rulesets     *secrulesets.Rulesets
	cdnRules     *cdnrules.Rule
	certificates *certificates.Certificates
	// sealer opens the certificates of the settings; nil without
	// SZ_DATA_KEY
	sealer *crypto.Sealer
}

func (m *MemStore) Shutdown(ctx context.Context) error {
	if _shared == nil {
		return nil
	}

	if err := m.synced.Close(); err != nil {
		szlog.Warnf("edge: close settings sync: %v", err)
	}

	if err := m.central.Close(); err != nil {
		szlog.Warnf("edge: close centraldata sync: %v", err)
	}

	return _shared.Shutdown(ctx)
}

func (m *MemStore) Rulesets() *secrulesets.Rulesets {
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

func (m *MemStore) Scripts() *scripts.Scripts {
	if m == nil {
		return nil
	}

	return m.scripts
}

func (m *MemStore) Route() *routes.Router {
	if m == nil {
		return nil
	}

	return m.route
}

func (m *MemStore) ReverseProxy() *upstream.ReverseProxy {
	if m == nil {
		return nil
	}

	return m.revProxy
}

// Cluster the Olric cluster this edge belongs to.
func (m *MemStore) Cluster() *cluster.Cluster {
	if m == nil {
		return nil
	}

	return _shared
}

func (m *MemStore) Limiter() *secratelimits.Limiter {
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

// Certificates the certificates the edge serves, by domain.
func (m *MemStore) Certificates() *certificates.Certificates {
	if m == nil {
		return nil
	}

	return m.certificates
}

func (m *MemStore) CDNRules() *cdnrules.Rule {
	if m == nil {
		return nil
	}

	return m.cdnRules
}

func (m *MemStore) LoadCDNRule(st ...*edgepb.Setting) error {
	for _, s := range st {
		ns := s.GetIngress().GetResource().GetDomain()
		m.cdnRules.Store(ns, s.GetDelivery())
	}

	return nil
}

func (m *MemStore) LoadRouter(s *edgepb.Setting) error {
	m.route.Store(s.GetIngress().GetResource())
	return nil
}

// LoadReverseProxy builds the proxies of the upstreams of s, dialed with
// its upstream TLS settings, in place of the previous ones of its
// namespace. When those settings fail the namespace has no proxy: its
// upstreams are never dialed without the CAs or the client certificate it
// asks for. A namespace in passthrough mode has its upstream addresses
// instead, for the listener (see Passthrough).
func (m *MemStore) LoadReverseProxy(
	server szhttp.Server, s *edgepb.Setting) error {

	res := s.GetIngress().GetResource()
	ns := res.GetDomain()
	if res.GetTlsMode() == _passthrough {
		// the listener forwards its connections, no request reaches here
		m.revProxy.Delete(ns)
		m.passthrough.Store(ns, passthroughAddrs(res))

		return nil
	}

	m.passthrough.Delete(ns)

	conf, err := m.upstreamTLS(s)
	if err != nil {
		m.revProxy.Delete(ns)
		return fmt.Errorf("reverse proxy: upstream tls: %w", err)
	}

	proxies, err := m.newReverseProxies(server, res, conf)
	m.revProxy.Store(ns, proxies)

	return err
}

// newReverseProxies the proxies of the upstreams of res, by server,
// dialing with conf; an upstream that fails is left out.
func (m *MemStore) newReverseProxies(server szhttp.Server,
	res *edgepb.Resource, conf *tls.Config,
) (map[string]szhttp.ReverseProxy, error) {

	servers := upstreamServers(res)
	proxies := make(map[string]szhttp.ReverseProxy, len(servers))

	var errs []error
	for _, upstream := range servers {
		rproxy, err := m.newReverseProxy(server, upstream,
			res.GetTlsMode(), conf)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		proxies[upstream.GetServer()] = rproxy
	}

	return proxies, errors.Join(errs...)
}

// newReverseProxy the proxy to upstream. Without conf it is the one every
// namespace shares for that target, so its connections are pooled once
// and kept across loads; with conf it is the namespace's own.
func (m *MemStore) newReverseProxy(server szhttp.Server,
	upstream *edgepb.Upstream, mode edgepb.TlsMode, conf *tls.Config,
) (szhttp.ReverseProxy, error) {

	target, err := targetOf(mode, upstream)
	if err != nil {
		return nil, fmt.Errorf("reverse proxy: target: %w", err)
	}

	if conf == nil {
		if rproxy, ok := m.plainProxy.Load(target); ok {
			return rproxy, nil
		}
	}

	rproxy, err := server.AcceptReverse(target, conf)
	if err != nil {
		return nil, fmt.Errorf("reverse proxy: create: %w", err)
	}

	if conf == nil {
		m.plainProxy.Store(target, rproxy)
	}

	return rproxy, nil
}

// _modeProtocol the protocol a TLS mode dials every upstream with; a mode
// left out follows the protocol of each upstream.
var _modeProtocol = map[edgepb.TlsMode]edgepb.ProxyProtocol{
	edgepb.TlsMode_TLS_MODE_TERMINATE: _protocolHTTP,
	edgepb.TlsMode_TLS_MODE_REENCRYPT: _protocolHTTPS,
}

const (
	_protocolHTTP  = edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTP
	_protocolHTTPS = edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTPS
)

// targetOf the target of the proxy to upstream, of the protocol mode asks
// for.
func targetOf(mode edgepb.TlsMode, upstream *edgepb.Upstream,
) (string, error) {
	if protocol, ok := _modeProtocol[mode]; ok {
		upstream = &edgepb.Upstream{
			Server: upstream.GetServer(), Protocol: protocol,
		}
	}

	return szhttp.Upstream2Target(upstream)
}

// upstreamServers the upstreams of res with a server, each server once, in
// the order of its locations.
func upstreamServers(res *edgepb.Resource) []*edgepb.Upstream {
	var (
		upstreams []*edgepb.Upstream
		seen      = make(map[string]struct{})
	)

	for _, loc := range res.GetLocations() {
		for _, upstream := range loc.GetProxyPass() {
			server := upstream.GetServer()
			if _, ok := seen[server]; ok || server == "" {
				continue
			}

			seen[server] = struct{}{}
			upstreams = append(upstreams, upstream)
		}
	}

	return upstreams
}

func (m *MemStore) LoadRateLimiter(s *edgepb.Setting) error {
	ns := s.GetIngress().GetResource().GetDomain()
	limiters := s.GetSecurity().GetLimiters()
	entries := make([]secratelimits.Entry, 0, len(limiters))
	held := m.limiter.Load(ns)

	for i, l := range limiters {
		if err := checkLimiter(l); err != nil {
			return fmt.Errorf("rate limiter %d: %w", i, err)
		}

		if l.GetIngress().GetStatus() != typepb.Status_STATUS_ACTIVE {
			szlog.Infof("edge:limiter: ignore '%s'", ns)
			continue
		}

		entry, err := newLimiterEntry(held, l)
		if err != nil {
			return fmt.Errorf("rate limiter: parse: %w", err)
		}

		entries = append(entries, entry)
	}

	m.limiter.Store(ns, entries)

	return nil
}

// checkLimiter a limiter must say what it counts: an expr, a nil one would
// match every request, or a script, not both.
func checkLimiter(l *edgepb.RateLimit) error {
	rule := l.GetIngress()
	hasExpr := len(rule.GetExpr().GetOrCondition()) > 0

	switch {
	case hasExpr && rule.GetScript() != "":
		return errLimiterBoth
	case !hasExpr && rule.GetScript() == "":
		return errLimiterNoExpr
	}

	return nil
}

func (m *MemStore) LoadRulesets(s *edgepb.Setting) error {
	ns := s.GetIngress().GetResource().GetDomain()

	err := m.rulesets.Store(ns, s.GetSecurity().GetRuleset())
	if err != nil {
		return fmt.Errorf("rulesets: init coraza waf: %w", err)
	}

	return nil
}

func (m *MemStore) LoadSecRule(s *edgepb.Setting) error {
	rules := make([]*secrulepb.SecRule, 0, len(s.GetSecurity().GetSecrules()))
	for _, rule := range s.GetSecurity().GetSecrules() {
		rules = append(rules, rule.GetIngress())
	}

	return m.secRule.Store(s.GetIngress().GetResource().GetDomain(), rules)
}

// LoadScripts compiles the Lua scripts of s.
func (m *MemStore) LoadScripts(s *edgepb.Setting) error {
	return m.scripts.Store(
		s.GetIngress().GetResource().GetDomain(),
		s.GetScripting().GetScripts(),
	)
}

func newLimiterEntry(held []secratelimits.Entry, l *edgepb.RateLimit,
) (secratelimits.Entry, error) {
	size, err := time.ParseDuration(l.GetTimeWindow())
	if err != nil {
		return secratelimits.Entry{}, fmt.Errorf("time window: %w", err)
	}

	timeout, err := time.ParseDuration(l.GetTimeout())
	if err != nil {
		return secratelimits.Entry{}, fmt.Errorf("timeout: %w", err)
	}

	rule := l.GetIngress()
	entry := secratelimits.Entry{
		Rule:    rule,
		Limiter: szlimiter.NewRateLimiter(timeout, size, l.GetMaxRequests()),
	}

	if rule.GetScript() == "" {
		entry.Eval = szrule.NewEval(rule.GetId(), rule.GetExpr())
		return entry, nil
	}

	entry.Script, err = limiterScript(held, rule)

	return entry, err
}

// limiterScript the program of the script of rule, the held one when it
// did not change.
func limiterScript(held []secratelimits.Entry, rule *secrulepb.SecRule,
) (*scripting.Program, error) {
	programs := make([]*scripting.Program, 0, len(held))
	for _, e := range held {
		if e.Script != nil {
			programs = append(programs, e.Script)
		}
	}

	return scripts.Rule(programs, rule, scripting.KeyFunc)
}
