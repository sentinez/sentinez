// Copyright 2026 Duc-Hung Ho.
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

package memory

import (
	"context"
	"errors"
	"fmt"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/sznormalize"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/sentinez/pkg/cluster"
	"google.golang.org/protobuf/proto"
)

const (
	// _settingsMap names the synced map of the edge settings.
	_settingsMap = "sentinez.dmz.edge.settings"
	// _centralMap names the synced map of the settings built from the
	// tables of centraldata, by resource id: only the leader of the zone
	// follows centraldata and writes it (CentralSink), every edge loads it.
	_centralMap = "sentinez.dmz.edge.central"
)

// settingSync holds the raw settings of every edge, by tenant id.
type settingSync = cluster.SyncMap[edgepb.Setting, *edgepb.Setting]

// newSettingSync the synced map of settings name of c.
func newSettingSync(c *cluster.Cluster, name string) *settingSync {
	return cluster.NewSyncMap[edgepb.Setting](c, name)
}

// Sync loads the settings of the cluster and shares local with it, then
// keeps the hot-path stores in step: every setting a node stores, here
// or on another edge, is rebuilt here. A setting is synced by its
// metadata.tenant_id; one without, or with the default one, stays on
// this edge only. The settings of centraldata are loaded the same way,
// from the map the leader writes.
func (m *MemStore) Sync(
	ctx context.Context, server szhttp.Server, local ...*edgepb.Setting,
) error {
	m.server = server
	m.synced.OnChange(m.applySetting)
	m.central.OnChange(m.applyCentral)

	if err := m.synced.Start(ctx); err != nil {
		return err
	}

	if err := m.central.Start(ctx); err != nil {
		return fmt.Errorf("centraldata: %w", err)
	}

	for _, s := range local {
		m.share(ctx, s)
	}

	return nil
}

// share stores s in the synced map; the map skips a setting it holds
// unchanged, so a restart does not make every edge rebuild it.
func (m *MemStore) share(ctx context.Context, s *edgepb.Setting) {
	id := s.GetMetadata().GetTenantId()
	if id == "" || id == defaults.TenantId {
		szlog.Warnf("edge: setting %q has no tenant id of its own, "+
			"not synced", namespace(s))
		m.locals.Store(namespace(s), s)
		m.load("", s)

		return
	}

	if err := m.synced.Store(ctx, id, s); err != nil {
		szlog.Errorf("edge: share setting %s: %v", id, err)
	}
}

// applySetting rebuilds the stores of the setting of resource id when
// it changes: a removed or renamed namespace is unloaded first.
func (m *MemStore) applySetting(id string, prev, next *edgepb.Setting) {
	m.unloadMoved(id, prev, next)

	if next != nil {
		m.load(id, next)
	}
}

// applyCentral rebuilds the stores of the setting centraldata holds for
// resource id: like applySetting, but it outranks the setting of the
// file that holds the namespace, and gives the namespace back to it once
// centraldata releases it.
func (m *MemStore) applyCentral(id string, prev, next *edgepb.Setting) {
	if szlog.Enabled(szlog.LevelDebug) {
		m.debugCentral(id, prev, next)
	}

	if ns := m.unloadMoved(id, prev, next); ns != "" {
		m.restore(ns)
	}

	if next == nil {
		return
	}

	ns := namespace(next)
	m.take(ns, id)
	m.load(id, next)
	szlog.Infof("edge: resource %s loaded from centraldata (%s)", id, ns)
}

// debugCentral logs a change of the map of centraldata and the node that
// wrote it: this edge when it leads, else the leader it synced it from.
func (m *MemStore) debugCentral(id string, prev, next *edgepb.Setting) {
	op, ns := "updated", namespace(next)
	switch {
	case prev == nil:
		op = "added"
	case next == nil:
		op, ns = "removed", namespace(prev)
	}

	node, local := m.central.Origin(id)
	if local {
		szlog.Debugf("edge: centraldata: %s resource %s (%s), written "+
			"here as the leader", op, id, ns)

		return
	}

	szlog.Debugf("edge: centraldata: %s resource %s (%s), synced from "+
		"the leader %s", op, id, ns, node)
}

// unloadMoved unloads the namespace of prev when next removes or renames
// it, and returns it if id held it.
func (m *MemStore) unloadMoved(id string, prev, next *edgepb.Setting,
) string {
	if prev == nil {
		return ""
	}

	old := namespace(prev)
	if (next != nil && namespace(next) == old) || !m.unload(id, old) {
		return ""
	}

	return old
}

// restore loads the setting of the file that holds namespace ns, this
// edge's own or a synced one, if any.
func (m *MemStore) restore(ns string) {
	if s, ok := m.locals.Load(ns); ok {
		m.load("", s)
		szlog.Infof("edge: %s loaded back from the local setting", ns)

		return
	}

	m.synced.Range(func(id string, s *edgepb.Setting) bool {
		if namespace(s) != ns {
			return true
		}

		m.load(id, s)
		szlog.Infof("edge: %s loaded back from the setting of %s", ns, id)

		return false
	})
}

// load normalizes a copy of s and rebuilds the stores of its namespace,
// unless another resource holds the namespace.
func (m *MemStore) load(id string, s *edgepb.Setting) {
	st, _ := proto.Clone(s).(*edgepb.Setting)
	sznormalize.NormalizeEdgeSetting(st)

	// the stores key st by it too
	ns := st.GetServer().GetResourceDomain()
	if ns == "" {
		szlog.Errorf("edge: setting of %s has no server.resourceDomain, "+
			"not loaded", id)

		return
	}

	if owner, ok := m.claim(ns, id); !ok {
		szlog.Errorf("edge: namespace %q of %s is held by %s",
			ns, id, owner)

		return
	}

	m.setting.Store(st)

	err := errors.Join(
		m.LoadRouter(st),
		m.LoadReverseProxy(m.server, st),
		m.LoadCDNRule(st),
		m.LoadSecRule(st),
		m.LoadRateLimiter(st),
		m.LoadRulesets(st),
		m.LoadCertificate(st),
	)
	if err != nil {
		szlog.Errorf("edge: load '%s': %v", ns, err)
	}
}

// unload removes namespace from the stores if resource id holds it, and
// reports whether it did. The reverse proxies stay: they are keyed by
// upstream, shared between namespaces.
func (m *MemStore) unload(id, ns string) bool {
	m.ownersMu.Lock()
	if m.owners[ns] != id {
		m.ownersMu.Unlock()
		return false
	}

	delete(m.owners, ns)
	m.ownersMu.Unlock()

	m.setting.Delete(ns)
	m.route.Delete(ns)
	m.cdnRules.Delete(ns)
	m.secRule.Delete(ns)
	m.limiter.Delete(ns)
	m.rulesets.Delete(ns)
	m.certificates.Delete(ns)

	return true
}

// namespace of the raw (not normalized) s: the resource domain of its
// server, as sznormalize.NormalizeEdgeSetting normalizes it.
func namespace(s *edgepb.Setting) string {
	return sznormalize.ResourceDomain(s.GetServer().GetResourceDomain())
}

// claim gives namespace to resource id, unless another one holds it.
func (m *MemStore) claim(ns, id string) (string, bool) {
	m.ownersMu.Lock()
	defer m.ownersMu.Unlock()

	if owner, ok := m.owners[ns]; ok && owner != id {
		return owner, false
	}

	m.owners[ns] = id

	return id, true
}

// take gives namespace to resource id, whoever held it.
func (m *MemStore) take(ns, id string) {
	m.ownersMu.Lock()
	defer m.ownersMu.Unlock()

	m.owners[ns] = id
}
