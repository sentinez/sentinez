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

package schema

import (
	"testing"

	"github.com/sentinez/pkg/szstorage/dbx"
	analyticpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	crspb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/stretchr/testify/assert"
)

// scannedColumnsTest a table and the columns the generated scanner of its
// model selects.
type scannedColumnsTest struct {
	name        string
	giveSchema  dbx.Schema
	giveColumns []string
}

// TestScannedColumnsInSchema checks that every column the generated
// scanner of a model selects is a column of its table.
func TestScannedColumnsInSchema(t *testing.T) {
	for _, tt := range scannedColumnsTests() {
		t.Run(tt.name, func(t *testing.T) {
			stored := append(tt.giveSchema.ColumnNames(),
				tt.giveSchema.CreatedAtColumn(),
				tt.giveSchema.UpdatedAtColumn())

			assert.Subset(t, stored, tt.giveColumns)
		})
	}
}

func scannedColumnsTests() []scannedColumnsTest {
	return []scannedColumnsTest{
		{"users", IAMUsers(), iampb.UserColumns()},
		{"accounts", IAMAccounts(), iampb.AccountColumns()},
		{"tenants", Tenants(), tenantpb.TenantColumns()},
		{"members", TenantMembers(), tenantpb.MemberColumns()},
		{"resources", TenantResources(), tenantpb.ResourceColumns()},
		{"secrules", SecuritySecRules(), securitypb.SecRuleColumns()},
		{"ratelimits", SecurityRateLimits(), securitypb.RateLimitColumns()},
		{"rulesets", SecurityCoreRulesets(), securitypb.CoreRulesetColumns()},
		{"activities", AnalyticActivities(), analyticpb.ActivityColumns()},
		{"secrule evt", AnalyticSecRulesEvent(), _secRuleEventColumns},
		{"ratelimit evt", AnalyticRateLimiterEvent(), _secRuleEventColumns},
		{"corerule evt", AnalyticCoreRulesetsEvent(), _coreRuleEventColumns},
		{"ingress evt", AnalyticHttpIngressEvent(), _ingressEventColumns},
		{"egress evt", AnalyticHttpEgressEvent(), _egressEventColumns},
	}
}

var (
	_secRuleEventColumns  = secrulepb.SecRuleEventColumns()
	_coreRuleEventColumns = crspb.CoreRuleEventColumns()
	_ingressEventColumns  = httppb.IngressEventColumns()
	_egressEventColumns   = httppb.EgressEventColumns()
)

// eventIndexTest a table of events and the column of its time.
type eventIndexTest struct {
	name       string
	giveSchema dbx.Schema
	giveTime   string
}

// indexedColumns the columns of each index of s.
func indexedColumns(s dbx.Schema) [][]string {
	indexed := make([][]string, 0, len(s.Indexes))
	for _, index := range s.Indexes {
		indexed = append(indexed, index.ColumnNames())
	}

	return indexed
}

// TestEventResourceTimeIndex checks that the events of a resource over a
// time range are served by one index.
func TestEventResourceTimeIndex(t *testing.T) {
	tests := []eventIndexTest{
		{"secrule", AnalyticSecRulesEvent(), "matched_time"},
		{"ratelimit", AnalyticRateLimiterEvent(), "matched_time"},
		{"corerule", AnalyticCoreRulesetsEvent(), "matched_time"},
		{"ingress", AnalyticHttpIngressEvent(), "timestamp"},
		{"egress", AnalyticHttpEgressEvent(), "timestamp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, indexedColumns(tt.giveSchema),
				[]string{"resource_id", tt.giveTime})
		})
	}
}

// TestRuleEventRequestIndex checks that the events of a request are
// served by an index; the http events are by the unique (id, time) key of
// their hypertable.
func TestRuleEventRequestIndex(t *testing.T) {
	tests := []eventIndexTest{
		{name: "secrule", giveSchema: AnalyticSecRulesEvent()},
		{name: "ratelimit", giveSchema: AnalyticRateLimiterEvent()},
		{name: "corerule", giveSchema: AnalyticCoreRulesetsEvent()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, indexedColumns(tt.giveSchema),
				[]string{"request_id"})
		})
	}
}
