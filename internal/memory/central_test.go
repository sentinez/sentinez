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
	"testing"

	"github.com/sentinez/pkg/sztable"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	cdnpb "github.com/sentinez/sentinez/api/proto/sentinez/types/cdn/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func upsertOf(t *testing.T, table sztable.Table, id string,
	msg proto.Message) *centraldatapb.Change {
	t.Helper()

	packed, err := anypb.New(msg)
	require.NoError(t, err)

	return &centraldatapb.Change{
		Table: table.String(),
		Op:    centraldatapb.Change_OP_UPSERT,
		Id:    id,
		Row:   packed,
	}
}

func resourceRow(t *testing.T, id, tenant,
	domain string) *centraldatapb.Change {
	t.Helper()

	return upsertOf(t, sztable.TenantResources, id, &tenantpb.Resource{
		Id:       id,
		TenantId: tenant,
		Server:   &edgepb.Server{ResourceId: id, ResourceDomain: domain},
	})
}

func ruleRow(t *testing.T, id, resource string,
	status typepb.Status) *centraldatapb.Change {
	t.Helper()

	return upsertOf(t, sztable.SecuritySecRules, id, &securitypb.SecRule{
		Id:     id,
		Owner:  &typepb.Owner{ResourceId: resource},
		Name:   id,
		Action: "ACTION_TYPE_BLOCK",
		Status: status,
		Expr:   &rulepb.Expression{OrCondition: []*rulepb.AndCondition{{}}},
	})
}

func ruleIds(m *MemStore, ns string) []string {
	var ids []string
	for _, e := range m.secRule.Load(ns) {
		ids = append(ids, e.Rule.GetId())
	}

	return ids
}

func TestReplaceLoadsTheResource(t *testing.T) {
	m := newSyncStore()
	sink := m.CentralSink()

	err := sink.Replace([]*centraldatapb.Change{
		resourceRow(t, "rsrc_1", "tnt_1", "central-load.example"),
		ruleRow(t, "srule_1", "rsrc_1", typepb.Status_STATUS_ACTIVE),
		ruleRow(t, "srule_off", "rsrc_1", typepb.Status_STATUS_DISABLE),
		ruleRow(t, "srule_other", "rsrc_2", typepb.Status_STATUS_ACTIVE),
	})

	require.NoError(t, err)
	st, err := m.setting.Load("central-load.example")
	require.NoError(t, err)
	assert.Equal(t, "tnt_1", st.GetMetadata().GetTenantId())
	assert.Equal(t, []string{"srule_1"}, ruleIds(m, "central-load.example"),
		"only the active rule of the resource")
	assert.Equal(t, "rsrc_1", m.owners["central-load.example"])
}

func TestApplyRebuildsOnlyTheChangedResource(t *testing.T) {
	m := newSyncStore()
	sink := m.CentralSink()
	require.NoError(t, sink.Replace([]*centraldatapb.Change{
		resourceRow(t, "rsrc_1", "tnt_1", "central-a.example"),
		resourceRow(t, "rsrc_2", "tnt_1", "central-b.example"),
		ruleRow(t, "srule_1", "rsrc_1", typepb.Status_STATUS_ACTIVE),
	}))
	before, err := m.setting.Load("central-b.example")
	require.NoError(t, err)

	require.NoError(t, sink.Apply(
		ruleRow(t, "srule_2", "rsrc_1", typepb.Status_STATUS_ACTIVE)))

	assert.Equal(t, []string{"srule_1", "srule_2"},
		ruleIds(m, "central-a.example"))
	after, err := m.setting.Load("central-b.example")
	require.NoError(t, err)
	assert.Same(t, before, after, "the other resource is not rebuilt")
}

func TestApplyDeleteOfARule(t *testing.T) {
	m := newSyncStore()
	sink := m.CentralSink()
	require.NoError(t, sink.Replace([]*centraldatapb.Change{
		resourceRow(t, "rsrc_1", "tnt_1", "central-del.example"),
		ruleRow(t, "srule_1", "rsrc_1", typepb.Status_STATUS_ACTIVE),
	}))

	require.NoError(t, sink.Apply(&centraldatapb.Change{
		Table: sztable.SecuritySecRules.String(),
		Op:    centraldatapb.Change_OP_DELETE,
		Id:    "srule_1",
	}))

	assert.Empty(t, ruleIds(m, "central-del.example"))
	assert.True(t, loadedNs(m, "central-del.example"),
		"the resource stays")
}

func TestApplyDeleteOfTheResourceUnloadsIt(t *testing.T) {
	m := newSyncStore()
	sink := m.CentralSink()
	require.NoError(t, sink.Replace([]*centraldatapb.Change{
		resourceRow(t, "rsrc_1", "tnt_1", "central-gone.example"),
	}))
	require.True(t, loadedNs(m, "central-gone.example"))

	require.NoError(t, sink.Apply(&centraldatapb.Change{
		Table: sztable.TenantResources.String(),
		Op:    centraldatapb.Change_OP_DELETE,
		Id:    "rsrc_1",
	}))

	assert.False(t, loadedNs(m, "central-gone.example"))
	assert.NotContains(t, m.owners, "central-gone.example")
}

func TestReplaceRemovesWhatTheNewSnapshotLacks(t *testing.T) {
	m := newSyncStore()
	sink := m.CentralSink()
	require.NoError(t, sink.Replace([]*centraldatapb.Change{
		resourceRow(t, "rsrc_1", "tnt_1", "central-lack.example"),
		ruleRow(t, "srule_1", "rsrc_1", typepb.Status_STATUS_ACTIVE),
	}))

	require.NoError(t, sink.Replace([]*centraldatapb.Change{
		resourceRow(t, "rsrc_1", "tnt_1", "central-lack.example"),
	}))

	assert.Empty(t, ruleIds(m, "central-lack.example"))
}

func TestMovingTheDomainUnloadsTheOldOne(t *testing.T) {
	m := newSyncStore()
	sink := m.CentralSink()
	require.NoError(t, sink.Replace([]*centraldatapb.Change{
		resourceRow(t, "rsrc_1", "tnt_1", "central-old.example"),
	}))

	require.NoError(t, sink.Apply(
		resourceRow(t, "rsrc_1", "tnt_1", "central-new.example")))

	assert.False(t, loadedNs(m, "central-old.example"))
	assert.True(t, loadedNs(m, "central-new.example"))
}

// The tables outrank the file, but keep what they do not hold: the CDN.
func TestTakesOverTheFileAndKeepsTheCDN(t *testing.T) {
	m := newSyncStore()
	file := newSetting("tnt_1", "central-file.example")
	file.Delivery = &edgepb.Delivery{Cdn: []*cdnpb.CDN{{}}}
	m.load("tnt_1", file)
	require.Equal(t, "tnt_1", m.owners["central-file.example"])
	sink := m.CentralSink()

	require.NoError(t, sink.Replace([]*centraldatapb.Change{
		resourceRow(t, "rsrc_1", "tnt_1", "central-file.example"),
		ruleRow(t, "srule_1", "rsrc_1", typepb.Status_STATUS_ACTIVE),
	}))

	assert.Equal(t, "rsrc_1", m.owners["central-file.example"])
	assert.Equal(t, []string{"srule_1"}, ruleIds(m, "central-file.example"))
	st, err := m.setting.Load("central-file.example")
	require.NoError(t, err)
	assert.Len(t, st.GetDelivery().GetCdn(), 1)
}

func TestCaCertificatesStaySealed(t *testing.T) {
	m := newSyncStore()
	sink := m.CentralSink()
	cert := func(id, resource string) *centraldatapb.Change {
		return upsertOf(t, sztable.TenantCaCertificates, id,
			&tenantpb.CaCertificate{
				Id: id, Owner: &typepb.Owner{ResourceId: resource},
				Pem: "v1:sealed", PrivateKey: "v1:sealed-key",
			})
	}

	require.NoError(t, sink.Replace([]*centraldatapb.Change{
		cert("cacrt_b", "rsrc_1"), cert("cacrt_a", "rsrc_1"),
		cert("cacrt_c", "rsrc_2"),
	}))
	got := sink.CaCertificates("rsrc_1")

	require.Len(t, got, 2)
	assert.Equal(t, "cacrt_a", got[0].GetId())
	assert.Equal(t, "v1:sealed-key", got[0].GetPrivateKey())
}
