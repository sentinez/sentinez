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
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/sentinez/pkg/szstorage/dbx"

	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/sznormalize"
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
type rows map[dbx.Table]map[string]proto.Message

// _centralWriteTimeout bounds the writes of one rebuild to the cluster:
// past it the rest apply here only, and the reconcile of the map shares
// them.
const _centralWriteTimeout = 30 * time.Second

// CentralSink applies the tables centraldata syncs to the edges: it holds
// the rows, and rebuilds the setting of every resource whose rows change
// into the synced map of centraldata, which every edge loads
// (MemStore.applyCentral). Only the leader of the zone runs one. The
// setting of a resource that is already loaded, by the proxy.yaml for
// instance, is the base the rows are applied to, so what the tables do not
// hold (the CDN) stays.
type CentralSink struct {
	m *MemStore

	mu   sync.Mutex
	rows rows
}

// CentralSink returns the sink of the tables of centraldata.
func (m *MemStore) CentralSink() *CentralSink {
	return &CentralSink{m: m, rows: make(rows)}
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

		put(next, dbx.Table(change.GetTable()), change.GetId(), msg)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	dirty := diff(c.rows, next)
	// the map may hold resources this sink never saw: those a former
	// leader wrote and that were deleted while the zone had none
	c.m.central.Range(func(id string, _ *edgepb.Setting) bool {
		if _, live := liveResource(next, id); !live {
			dirty[id] = struct{}{}
		}

		return true
	})

	szlog.Debugf("edge: centraldata: loaded %d rows, %d resources to "+
		"rebuild", len(changes), len(dirty))

	c.rows = next
	c.rebuild(dirty)

	return nil
}

// Apply implements centralsync.Sink.
func (c *CentralSink) Apply(change *centraldatapb.Change) error {
	table, id := dbx.Table(change.GetTable()), change.GetId()
	szlog.Debugf("edge: centraldata: received %s %s %s, version %d",
		change.GetOp(), change.GetTable(), change.GetId(),
		change.GetVersion())

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

func put(r rows, table dbx.Table, id string, msg proto.Message) {
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
	case *securitypb.SecRateLimit:
		return row.GetOwner().GetResourceId()
	case *securitypb.SecRuleset:
		return row.GetOwner().GetResourceId()
	case *tenantpb.CaCertificate:
		return row.GetOwner().GetResourceId()
	}

	return ""
}

// rebuild never fails: a write the cluster missed is applied here, and
// the next reconcile of the map shares it.
func (c *CentralSink) rebuild(dirty map[string]struct{}) {
	ctx, cancel := context.WithTimeout(context.Background(),
		_centralWriteTimeout)
	defer cancel()

	for id := range dirty {
		if id == "" {
			continue
		}

		if err := c.rebuildOne(ctx, id); err != nil {
			szlog.Warnf("edge: centraldata: resource %s: %v", id, err)
		}
	}
}

// rebuildOne writes the setting of resource id from its rows, or removes
// it when the resource is gone. The map skips a setting it holds
// unchanged, so the edges rebuild only what changed.
func (c *CentralSink) rebuildOne(ctx context.Context, id string) error {
	if res, live := liveResource(c.rows, id); live {
		// the map skips it when unchanged: no edge logs it then
		szlog.Debugf("edge: centraldata: rebuild resource %s (%s) for the "+
			"cluster", id, res.GetDomain())

		return c.m.central.Store(ctx, id, c.setting(res))
	}

	if _, held := c.m.central.Load(id); held {
		szlog.Debugf("edge: centraldata: remove resource %s from the "+
			"cluster", id)

		return c.m.central.Delete(ctx, id)
	}

	return nil
}

// liveResource the resource id of r, when it has a domain to serve.
func liveResource(r rows, id string) (*tenantpb.Resource, bool) {
	res, ok := r[sztable.TenantResources][id].(*tenantpb.Resource)

	return res, ok && res.GetDomain() != ""
}

// setting the setting of res: the one loaded for its namespace, if any,
// with the resource, the certificates and the security of the rows.
func (c *CentralSink) setting(res *tenantpb.Resource) *edgepb.Setting {
	// none loaded: the setting is res alone
	cur, _ := c.m.setting.Load(sznormalize.ResourceDomain(res.GetDomain()))

	st := sznormalize.EdgeSetting(res, cur)
	st.Ingress.Tls = c.tls(res.GetId())
	st.Security = c.security(res.GetId(), st.GetSecurity())

	return st
}

// tls the certificates of resource id, still sealed, by id, sorted by
// usage: the ones the edge serves (SERVER), the CAs of the upstreams
// (UPSTREAM) and the client certificate it presents to them
// (UPSTREAM_CLIENT). Among SERVER and UPSTREAM_CLIENT, the ones without a
// private key are the chain of the others.
func (c *CentralSink) tls(id string) *edgepb.Tls {
	t := &edgepb.Tls{}
	found := false
	for _, msg := range c.sorted(sztable.TenantCaCertificates, id) {
		row, ok := msg.(*tenantpb.CaCertificate)
		if !ok {
			continue
		}

		if dst := tlsSlot(t, row.GetUsage()); dst != nil {
			found = true
			*dst = append(*dst, &edgepb.Certificate{
				Id:         row.GetId(),
				Pem:        row.GetPem(),
				PrivateKey: row.GetPrivateKey(),
			})
		}
	}

	if !found {
		return nil
	}

	return t
}

// tlsSlot the list of t a certificate of usage goes in, nil for none.
func tlsSlot(t *edgepb.Tls, usage tenantpb.CaCertificateUsage,
) *[]*edgepb.Certificate {
	switch usage {
	case tenantpb.CaCertificateUsage_CA_CERTIFICATE_USAGE_SERVER:
		return &t.Certificates
	case tenantpb.CaCertificateUsage_CA_CERTIFICATE_USAGE_UPSTREAM:
		return &t.UpstreamCas
	case tenantpb.CaCertificateUsage_CA_CERTIFICATE_USAGE_UPSTREAM_CLIENT:
		return &t.UpstreamCertificates
	default:
		return nil
	}
}

// security the security of resource id from its rows; the ruleset of
// base stays when the tables hold none.
func (c *CentralSink) security(id string, base *edgepb.Security,
) *edgepb.Security {
	sec := &edgepb.Security{Ruleset: base.GetRuleset()}

	for _, msg := range c.sorted(sztable.SecuritySecRules, id) {
		sec.Secrules = append(sec.Secrules, &edgepb.SecRule{
			Ingress: secRuleOf(msg.(*securitypb.SecRule)),
		})
	}

	for _, msg := range c.sorted(sztable.SecuritySecRateLimits, id) {
		sec.Limiters = append(sec.Limiters, rateLimitOf(
			msg.(*securitypb.SecRateLimit)))
	}

	for _, msg := range c.sorted(sztable.SecuritySecRulesets, id) {
		sec.Ruleset = msg.(*securitypb.SecRuleset).GetSetting()

		break
	}

	return sec
}

// sorted the rows of table of resource id, by id.
func (c *CentralSink) sorted(table dbx.Table, id string) []proto.Message {
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

func rateLimitOf(row *securitypb.SecRateLimit) *edgepb.RateLimit {
	return &edgepb.RateLimit{
		Ingress:     secRuleOf(row),
		TimeWindow:  row.GetTimeWindow(),
		MaxRequests: row.GetMaxRequests(),
		Timeout:     row.GetTimeout(),
	}
}

func actionOf(name string) *rulepb.Action {
	if name == "" {
		return nil
	}

	return &rulepb.Action{Type: rulepb.ActionType(
		rulepb.ActionType_value[name])}
}
