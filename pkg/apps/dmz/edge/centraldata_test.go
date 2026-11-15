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

package edge

import (
	"context"
	"testing"

	"github.com/sentinez/sentinez/pkg/cluster/leader"
	"github.com/stretchr/testify/assert"
)

// Once the server shuts down, Lead returns without campaigning: the nil
// elector would panic otherwise.
func TestLeadAfterShutdownDoesNotCampaign(t *testing.T) {
	s := &Server{}
	s.endLead()

	assert.NoError(t, s.Lead(context.Background()))
}

// A term that ends before the edge synced its map follows nothing: the
// nil memstore would panic otherwise.
func TestFollowCentralDataWaitsForTheSync(t *testing.T) {
	s := &Server{synced: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.NoError(t, s.followCentralData(ctx, leader.Term{}))
}
