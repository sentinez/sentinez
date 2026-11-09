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
	"testing"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/sentinez/internal/memory/cdnrules"
	"github.com/sentinez/sentinez/internal/memory/ratelimiter"
	"github.com/sentinez/sentinez/internal/memory/routes"
	"github.com/sentinez/sentinez/internal/memory/secrules"
	"github.com/sentinez/sentinez/internal/memory/settings"
	"github.com/sentinez/sentinez/pkg/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newSyncStore is a MemStore whose synced map is not started: its writes
// apply locally, which runs the same change hook as a synced write.
func newSyncStore() *MemStore {
	m := &MemStore{
		synced:   newSettingSync(&cluster.Cluster{}),
		owners:   make(map[string]string),
		setting:  settings.New(),
		limiter:  ratelimiter.New(),
		route:    routes.New(),
		secRule:  secrules.New(),
		cdnRules: cdnrules.New(),
	}
	m.synced.OnChange(m.applySetting)

	return m
}

func newSetting(id, ns string) *edgepb.Setting {
	return &edgepb.Setting{
		Metadata: &edgepb.Metadata{TenantId: id},
		Server:   &edgepb.Server{ResourceDomain: ns},
	}
}

func loadedNs(m *MemStore, ns string) bool {
	_, err := m.setting.Load(ns)
	return err == nil
}

type shareTest struct {
	name       string
	giveShares []*edgepb.Setting
	giveDelete string // a resource id, or none
	wantLoaded []string
	wantGone   []string
	wantOwners map[string]string
}

func shareTests() []shareTest {
	return []shareTest{
		{
			name:       "loads",
			giveShares: []*edgepb.Setting{newSetting("res-1", "ns-load")},
			wantLoaded: []string{"ns-load"},
			wantOwners: map[string]string{"ns-load": "res-1"},
		},
		{
			name: "rename unloads the old namespace",
			giveShares: []*edgepb.Setting{
				newSetting("res-1", "ns-old"), newSetting("res-1", "ns-new"),
			},
			wantLoaded: []string{"ns-new"},
			wantGone:   []string{"ns-old"},
			wantOwners: map[string]string{"ns-new": "res-1"},
		},
		{
			name:       "delete unloads",
			giveShares: []*edgepb.Setting{newSetting("res-1", "ns-del")},
			giveDelete: "res-1",
			wantGone:   []string{"ns-del"},
			wantOwners: map[string]string{},
		},
		{
			name: "namespace held by another resource",
			giveShares: []*edgepb.Setting{
				newSetting("res-1", "ns-dup"), newSetting("res-2", "ns-dup"),
			},
			giveDelete: "res-2",
			wantLoaded: []string{"ns-dup"},
			wantOwners: map[string]string{"ns-dup": "res-1"},
		},
	}
}

func TestShareAppliesChanges(t *testing.T) {
	for _, tt := range shareTests() {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := newSyncStore()
			for _, s := range tt.giveShares {
				m.share(ctx, s)
			}

			// deleting a missing key changes nothing
			require.NoError(t, m.synced.Delete(ctx, tt.giveDelete))

			for _, ns := range tt.wantLoaded {
				assert.True(t, loadedNs(m, ns), ns)
			}

			for _, ns := range tt.wantGone {
				assert.False(t, loadedNs(m, ns), ns)
			}

			assert.Equal(t, tt.wantOwners, m.owners)
		})
	}
}

func TestShareSkipsUnchangedSetting(t *testing.T) {
	ctx := context.Background()
	m := newSyncStore()

	var changes int
	m.synced.OnChange(func(id string, prev, next *edgepb.Setting) {
		changes++
		m.applySetting(id, prev, next)
	})

	m.share(ctx, newSetting("res-1", "sync-same"))
	m.share(ctx, newSetting("res-1", "sync-same"))

	assert.Equal(t, 1, changes)
}

func TestShareWithoutTenantIdStaysLocal(t *testing.T) {
	m := newSyncStore()

	m.share(context.Background(), newSetting("", "sync-local"))

	assert.True(t, loadedNs(m, "sync-local"))
	_, ok := m.synced.Load("")
	assert.False(t, ok)
}

func TestShareDefaultTenantIdStaysLocal(t *testing.T) {
	m := newSyncStore()

	m.share(context.Background(),
		newSetting(defaults.TenantId, "sync-default"))

	assert.True(t, loadedNs(m, "sync-default"))
	_, ok := m.synced.Load(defaults.TenantId)
	assert.False(t, ok)
}
