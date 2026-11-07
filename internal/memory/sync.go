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

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/sznormalize"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/sentinez/pkg/cluster"
	"google.golang.org/protobuf/proto"
)

// _settingsMap names the synced map of the edge settings.
const _settingsMap = "sentinez.dmz.edge.settings"

// settingSync holds the raw settings of every edge, by tenant id.
type settingSync = cluster.SyncMap[edgepb.Setting, *edgepb.Setting]

func newSettingSync(c *cluster.Cluster) *settingSync {
	return cluster.NewSyncMap[edgepb.Setting](c, _settingsMap)
}

// Sync loads the settings of the cluster and shares local with it, then
// keeps the hot-path stores in step: every setting a node stores, here
// or on another edge, is rebuilt here. A setting is synced by its
// metadata.tenant_id; one without, or with the default one, stays on
// this edge only.
func (m *MemStore) Sync(
	ctx context.Context, server szhttp.Server, local ...*edgepb.Setting,
) error {
	m.server = server
	m.synced.OnChange(m.applySetting)

	if err := m.synced.Start(ctx); err != nil {
		return err
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
			"not synced", s.GetServer().GetName())
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
	ns := next.GetServer().GetName()
	if prev != nil && (next == nil || prev.GetServer().GetName() != ns) {
		m.unload(id, prev.GetServer().GetName())
	}

	if next != nil {
		m.load(id, next)
	}
}

// load normalizes a copy of s and rebuilds the stores of its namespace,
// unless another resource holds the namespace.
func (m *MemStore) load(id string, s *edgepb.Setting) {
	ns := s.GetServer().GetName()
	if owner, ok := m.claim(ns, id); !ok {
		szlog.Errorf("edge: namespace %q of %s is held by %s",
			ns, id, owner)

		return
	}

	st, _ := proto.Clone(s).(*edgepb.Setting)
	sznormalize.NormalizeEdgeSetting(st)
	m.setting.Store(st)

	err := errors.Join(
		m.LoadRouter(st),
		m.LoadReverseProxy(m.server, st),
		m.LoadCDNRule(st),
		m.LoadSecRule(st),
		m.LoadRateLimiter(st),
		m.LoadRulesets(st),
	)
	if err != nil {
		szlog.Errorf("edge: load '%s': %v", ns, err)
	}
}

// unload removes namespace from the stores if resource id holds it. The
// reverse proxies stay: they are keyed by upstream, shared between
// namespaces.
func (m *MemStore) unload(id, ns string) {
	m.ownersMu.Lock()
	if m.owners[ns] != id {
		m.ownersMu.Unlock()
		return
	}

	delete(m.owners, ns)
	m.ownersMu.Unlock()

	m.setting.Delete(ns)
	m.route.Delete(ns)
	m.cdnRules.Delete(ns)
	m.secRule.Delete(ns)
	m.limiter.Delete(ns)
	m.rulesets.Delete(ns)
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
