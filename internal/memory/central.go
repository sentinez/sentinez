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
	"cmp"
	"fmt"
	"slices"
	"sync"

	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/sztable"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/protobuf/proto"
)

// rows the rows of the synced tables, by table then id.
type rows map[sztable.Table]map[string]proto.Message

// CentralSink applies the tables centraldata syncs to the stores of the
// edge: it holds the rows, and rebuilds the setting of every resource whose
// rows change. The setting of a resource that is already loaded, by the
// proxy.yaml for instance, is the base the rows are applied to, so what the
// tables do not hold (the CDN) stays.
type CentralSink struct {
	m *MemStore

	mu    sync.Mutex
	rows  rows
	built map[string]string // resource id -> the namespace it loaded
}

// CentralSink returns the sink of the tables of centraldata.
func (m *MemStore) CentralSink() *CentralSink {
	return &CentralSink{m: m, rows: make(rows), built: make(map[string]string)}
}

// Replace implements centralsync.Sink.
func (c *CentralSink) Replace(changes []*centraldatapb.Change) error {
	next := make(rows)
	for _, change := range changes {
		msg, err := change.GetRow().UnmarshalNew()
		if err != nil {
			return fmt.Errorf("%s %s: %w", change.GetTable(),
				change.GetId(), err)
		}

		put(next, sztable.Table(change.GetTable()), change.GetId(), msg)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	dirty := diff(c.rows, next)
	c.rows = next
	c.rebuild(dirty)

	return nil
}

// Apply implements centralsync.Sink.
func (c *CentralSink) Apply(change *centraldatapb.Change) error {
	table, id := sztable.Table(change.GetTable()), change.GetId()

	c.mu.Lock()
	defer c.mu.Unlock()

	dirty := make(map[string]struct{})
	if old, ok := c.rows[table][id]; ok {
		dirty[resourceOf(old)] = struct{}{}
	}

	if change.GetOp() == centraldatapb.Change_OP_DELETE {
		delete(c.rows[table], id)
	} else {
		msg, err := change.GetRow().UnmarshalNew()
		if err != nil {
			return fmt.Errorf("%s %s: %w", table, id, err)
		}

		put(c.rows, table, id, msg)
		dirty[resourceOf(msg)] = struct{}{}
	}

	c.rebuild(dirty)

	return nil
}

// CaCertificates the certificates of the resource, as the table holds them:
// the pem and the private key are still sealed with SENZ_DATA_KEY.
func (c *CentralSink) CaCertificates(resourceId string,
) []*tenantpb.CaCertificate {
	c.mu.Lock()
	defer c.mu.Unlock()

	var certs []*tenantpb.CaCertificate
	for _, msg := range c.rows[sztable.TenantCaCertificates] {
		if cert, ok := msg.(*tenantpb.CaCertificate); ok &&
			cert.GetOwner().GetResourceId() == resourceId {
			certs = append(certs, proto.CloneOf(cert))
		}
	}

	slices.SortFunc(certs, func(a, b *tenantpb.CaCertificate) int {
		return cmp.Compare(a.GetId(), b.GetId())
	})

	return certs
}

func put(r rows, table sztable.Table, id string, msg proto.Message) {
	if r[table] == nil {
		r[table] = make(map[string]proto.Message)
	}

	r[table][id] = msg
}

// diff the resources whose rows differ between old and next.
func diff(old, next rows) map[string]struct{} {
	dirty := make(map[string]struct{})

	for table, byId := range old {
		for id, msg := range byId {
			if cur, ok := next[table][id]; !ok || !proto.Equal(cur, msg) {
				dirty[resourceOf(msg)] = struct{}{}
			}
		}
	}

	for table, byId := range next {
		for id, msg := range byId {
			if _, ok := old[table][id]; !ok {
				dirty[resourceOf(msg)] = struct{}{}
			}
		}
	}

	return dirty
}

// resourceOf the resource a row belongs to.
func resourceOf(msg proto.Message) string {
	switch row := msg.(type) {
	case *tenantpb.Resource:
		return row.GetId()
	case *securitypb.SecRule:
		return row.GetOwner().GetResourceId()
	case *securitypb.RateLimit:
		return row.GetOwner().GetResourceId()
	case *securitypb.CoreRuleset:
		return row.GetOwner().GetResourceId()
	case *tenantpb.CaCertificate:
		return row.GetOwner().GetResourceId()
	}

	return ""
}

func (c *CentralSink) rebuild(dirty map[string]struct{}) {
	for id := range dirty {
		if id == "" {
			continue
		}

		c.rebuildOne(id)
	}
}

// rebuildOne reloads the stores of resource id from its rows, or unloads
// them when the resource is gone.
func (c *CentralSink) rebuildOne(id string) {
	res, ok := c.rows[sztable.TenantResources][id].(*tenantpb.Resource)
	if !ok || res.GetServer().GetResourceDomain() == "" {
		if ns, held := c.built[id]; held {
			c.m.unload(id, ns)
			delete(c.built, id)
		}

		return
	}

	setting := c.setting(res)
	ns := namespace(setting)

	if old, held := c.built[id]; held && old != ns {
		c.m.unload(id, old)
	}

	// the settings of centraldata outrank those of the file
	c.m.take(ns, id)
	c.m.load(id, setting)
	c.built[id] = ns
	szlog.Infof("edge: resource %s loaded from centraldata (%s)", id, ns)
}

// setting the setting of res: the one loaded for its namespace, if any,
// with the server and the security of the rows.
func (c *CentralSink) setting(res *tenantpb.Resource) *edgepb.Setting {
	server := proto.CloneOf(res.GetServer())
	st := &edgepb.Setting{}

	if cur, err := c.m.setting.Load(
		namespace(&edgepb.Setting{Server: server})); err == nil {
		st = proto.CloneOf(cur)
	}

	if st.Metadata == nil {
		st.Metadata = &edgepb.Metadata{}
	}

	st.Metadata.TenantId = cmp.Or(res.GetTenantId(), st.Metadata.TenantId)
	st.Server = server
	st.Security = c.security(res.GetId(), st.GetSecurity())

	return st
}

// security the security of resource id from its rows; the core ruleset of
// base stays when the tables hold none.
func (c *CentralSink) security(id string, base *edgepb.Security,
) *edgepb.Security {
	sec := &edgepb.Security{CoreRuleset: base.GetCoreRuleset()}

	for _, msg := range c.sorted(sztable.SecuritySecRules, id) {
		sec.SecRules = append(sec.SecRules, &edgepb.SecRule{
			IngressRuntime: secRuleOf(msg.(*securitypb.SecRule)),
		})
	}

	for _, msg := range c.sorted(sztable.SecurityRateLimits, id) {
		sec.Limiters = append(sec.Limiters, rateLimitOf(
			msg.(*securitypb.RateLimit)))
	}

	for _, msg := range c.sorted(sztable.SecurityCoreRulesets, id) {
		sec.CoreRuleset = msg.(*securitypb.CoreRuleset).GetContent()

		break
	}

	return sec
}

// sorted the rows of table of resource id, by id.
func (c *CentralSink) sorted(table sztable.Table, id string) []proto.Message {
	var list []proto.Message
	for _, msg := range c.rows[table] {
		if resourceOf(msg) == id {
			list = append(list, msg)
		}
	}

	slices.SortFunc(list, func(a, b proto.Message) int {
		return cmp.Compare(idOf(a), idOf(b))
	})

	return list
}

func idOf(msg proto.Message) string {
	if m, ok := msg.(interface{ GetId() string }); ok {
		return m.GetId()
	}

	return ""
}

// securityRow the fields of a rule the rows of the security tables share.
type securityRow interface {
	GetId() string
	GetName() string
	GetDescription() string
	GetExpr() *rulepb.Expression
	GetAction() string
	GetStatus() typepb.Status
	GetPriority() int32
}

// secRuleOf the runtime rule of a row. Only the type of the action is
// stored, so the parameters of the action in the file are not kept.
func secRuleOf(row securityRow) *secrulepb.SecRule {
	return &secrulepb.SecRule{
		Id:          row.GetId(),
		Name:        row.GetName(),
		Description: row.GetDescription(),
		Expr:        row.GetExpr(),
		Action:      actionOf(row.GetAction()),
		Status:      row.GetStatus(),
		Priority:    row.GetPriority(),
	}
}

func rateLimitOf(row *securitypb.RateLimit) *edgepb.RateLimit {
	return &edgepb.RateLimit{
		IngressRuntime: secRuleOf(row),
		TimeWindow:     row.GetTimeWindow(),
		MaxRequests:    row.GetMaxRequests(),
		Timeout:        row.GetTimeout(),
	}
}

func actionOf(name string) *rulepb.Action {
	if name == "" {
		return nil
	}

	return &rulepb.Action{Type: rulepb.ActionType(
		rulepb.ActionType_value[name])}
}
