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

package centraldatasvc

import (
	"context"
	"testing"

	memrepo "github.com/sentinez/controlplane/apps/centraldata/v1/repos/memory"
	"github.com/sentinez/pkg/sztable"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// collect runs Sync and returns what it sent.
func collect(srv *CentralDataService, req *centraldatapb.SyncRequest,
) ([]*centraldatapb.SyncResponse, error) {
	var sent []*centraldatapb.SyncResponse
	err := srv.Sync(context.Background(), req,
		func(r *centraldatapb.SyncResponse) error {
			sent = append(sent, r)

			return nil
		})

	return sent, err
}

func newSyncService(t *testing.T) *CentralDataService {
	t.Helper()

	memory, err := memrepo.New()
	require.NoError(t, err)
	require.NoError(t, memory.Upsert(sztable.SecuritySecRules,
		rule("b", "two"), rule("a", "one")))
	require.NoError(t, memory.Upsert(sztable.TenantResources,
		&tenantpb.Resource{Id: "r"}))

	return New(nil, memory, sztable.SecuritySecRules, sztable.TenantResources)
}

func TestSyncStreamsRowsThenSnapshot(t *testing.T) {
	srv := newSyncService(t)

	sent, err := collect(srv, &centraldatapb.SyncRequest{})

	require.NoError(t, err)
	require.Len(t, sent, 4)

	var ids []string
	for _, r := range sent[:3] {
		ids = append(ids, r.GetRow().GetTable()+"/"+r.GetRow().GetId())
		assert.Equal(t, centraldatapb.Change_OP_UPSERT, r.GetRow().GetOp())
	}
	assert.Equal(t, []string{
		"security.secrules/a", "security.secrules/b", "tenant.resources/r",
	}, ids)

	done := sent[3].GetDone()
	require.NotNil(t, done, "the last message is the snapshot")
	assert.Equal(t, srv.memory.Version(), done.GetVersion())
	assert.Equal(t, srv.memory.Epoch(), done.GetEpoch())
}

func TestSyncRowCarriesTheModel(t *testing.T) {
	srv := newSyncService(t)

	sent, err := collect(srv, &centraldatapb.SyncRequest{
		Tables: []string{"security.secrules"},
	})
	require.NoError(t, err)
	require.Len(t, sent, 3)

	var got securitypb.SecRule
	require.NoError(t, sent[0].GetRow().GetRow().UnmarshalTo(&got))
	assert.Equal(t, "one", got.GetName())
}

func TestSyncOneTable(t *testing.T) {
	srv := newSyncService(t)

	sent, err := collect(srv, &centraldatapb.SyncRequest{
		Tables: []string{"tenant.resources"},
	})

	require.NoError(t, err)
	require.Len(t, sent, 2)
	assert.Equal(t, "r", sent[0].GetRow().GetId())
}

func TestSyncRefusesATableNotSynced(t *testing.T) {
	srv := newSyncService(t)

	for _, name := range []string{"iam.users", "nope"} {
		t.Run(name, func(t *testing.T) {
			sent, err := collect(srv,
				&centraldatapb.SyncRequest{Tables: []string{name}})

			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Empty(t, sent, "nothing may be streamed")
		})
	}
}

func TestSyncStopsWhenTheCallerLeaves(t *testing.T) {
	srv := newSyncService(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := srv.Sync(ctx, &centraldatapb.SyncRequest{},
		func(*centraldatapb.SyncResponse) error { return nil })

	assert.ErrorIs(t, err, context.Canceled)
}
