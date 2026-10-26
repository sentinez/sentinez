// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cdnrules

import (
	"testing"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	cdnpb "github.com/sentinez/sentinez/api/proto/sentinez/types/cdn/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/szsync"
	"github.com/stretchr/testify/assert"
)

func TestStoreOrdersByPriority(t *testing.T) {
	rc := newTestRule()

	rc.Store("ns", delivery(
		cdn("low", typepb.Status_STATUS_ACTIVE, 1),
		cdn("off", typepb.Status_STATUS_DISABLE, 99),
		cdn("high", typepb.Status_STATUS_ACTIVE, 10),
		cdn("low-2", typepb.Status_STATUS_ACTIVE, 1),
	))

	assert.Equal(t, []string{"high", "low", "low-2"}, idsOf(rc.Load("ns")))
}

func TestStoreEmptyDeletes(t *testing.T) {
	rc := newTestRule()
	rc.Store("ns", delivery(cdn("a", typepb.Status_STATUS_ACTIVE, 0)))

	rc.Store("ns", delivery(cdn("off", typepb.Status_STATUS_DISABLE, 0)))

	assert.Nil(t, rc.Load("ns"))
}

func newTestRule() *Rule {
	return &Rule{chains: szsync.NewMap[string, []Entry]()}
}

func delivery(cdns ...*cdnpb.CDN) *edgepb.Delivery {
	return &edgepb.Delivery{Cdn: cdns}
}

func cdn(id string, status typepb.Status, priority int32) *cdnpb.CDN {
	return &cdnpb.CDN{RuleRuntime: &cdnpb.Rule{
		Id: id, Status: status, Priority: priority,
	}}
}

func idsOf(chain []Entry) []string {
	var ids []string
	for _, e := range chain {
		ids = append(ids, e.Rule.GetId())
	}

	return ids
}
