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
	analyticpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	crspb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/sentinez/shared/szstorage/dbx"
	"github.com/sentinez/shared/sztable"
)

// AnalyticActivities the schema of the sztable.AnalyticActivities table.
func AnalyticActivities() dbx.Schema {
	return dbx.Schema{
		Table: sztable.AnalyticActivities,
		Columns: []dbx.Column{
			{Name: analyticpb.Activity_Id, Type: dbx.Text},
			{Name: analyticpb.Activity_OwnerResourceId, Type: dbx.Text},
			{Name: analyticpb.Activity_OwnerTenantId, Type: dbx.Text},
			{Name: analyticpb.Activity_UniqueVisitor, Type: dbx.IntArr},
		},
	}
}

// AnalyticSecRulesEvent the schema of the sztable.AnalyticSecRulesEvent
// table: the secrulepb.RuleEvent records of the sztopic.EventHTTPSecRule
// topic.
func AnalyticSecRulesEvent() dbx.Schema {
	return ruleEvents(sztable.AnalyticSecRulesEvent)
}

// AnalyticRateLimiterEvent the schema of the
// sztable.AnalyticRateLimiterEvent table: the secrulepb.RuleEvent records
// of the sztopic.EventHTTPRateLimit topic.
func AnalyticRateLimiterEvent() dbx.Schema {
	return ruleEvents(sztable.AnalyticRateLimiterEvent)
}

// AnalyticCoreRulesetsEvent the schema of the
// sztable.AnalyticCoreRulesetsEvent table: the corerulesetpb.CoreRuleEvent
// records of the sztopic.EventHTTPRulesets topic.
func AnalyticCoreRulesetsEvent() dbx.Schema {
	return dbx.Schema{
		Table: sztable.AnalyticCoreRulesetsEvent,
		Columns: []dbx.Column{
			{Name: crspb.CoreRuleEvent_Id, Type: dbx.Text},
			{Name: crspb.CoreRuleEvent_MatchedRuleIds, Type: dbx.IntArr},
			{Name: crspb.CoreRuleEvent_Severities, Type: dbx.TextArr},
			{Name: crspb.CoreRuleEvent_Messages, Type: dbx.TextArr},
			{Name: crspb.CoreRuleEvent_TotalScore, Type: dbx.Int4},
			{Name: crspb.CoreRuleEvent_MatchedTime, Type: dbx.Int8},
			{Name: crspb.CoreRuleEvent_ResourceId, Type: dbx.Text},
			{Name: crspb.CoreRuleEvent_TenantId, Type: dbx.Text},
			{Name: crspb.CoreRuleEvent_Service, Type: dbx.Int4},
			{Name: crspb.CoreRuleEvent_RequestIp, Type: dbx.Text},
			{Name: crspb.CoreRuleEvent_RequestDomain, Type: dbx.Text},
			{Name: crspb.CoreRuleEvent_RequestScheme, Type: dbx.Text},
			{Name: crspb.CoreRuleEvent_RequestPath, Type: dbx.Text},
			{Name: crspb.CoreRuleEvent_RequestQuery, Type: dbx.Text},
			{Name: crspb.CoreRuleEvent_RequestId, Type: dbx.Text},
			{Name: crspb.CoreRuleEvent_Behavior, Type: dbx.Int4},
		},
		Indexes: []dbx.Index{
			{Column: crspb.CoreRuleEvent_MatchedTime, Method: dbx.BTree},
			{Column: crspb.CoreRuleEvent_ResourceId, Method: dbx.BTree},
		},
	}
}

// ruleEvents the schema of a table of secrulepb.RuleEvent.
func ruleEvents(table sztable.Table) dbx.Schema {
	return dbx.Schema{
		Table: table,
		Columns: []dbx.Column{
			{Name: secrulepb.RuleEvent_Id, Type: dbx.Text},
			{Name: secrulepb.RuleEvent_MatchedSecruleIds, Type: dbx.TextArr},
			{Name: secrulepb.RuleEvent_Messages, Type: dbx.TextArr},
			{Name: secrulepb.RuleEvent_ResourceId, Type: dbx.Text},
			{Name: secrulepb.RuleEvent_TenantId, Type: dbx.Text},
			{Name: secrulepb.RuleEvent_MatchedTime, Type: dbx.Int8},
			{Name: secrulepb.RuleEvent_Action, Type: dbx.JSONB},
			{Name: secrulepb.RuleEvent_Service, Type: dbx.Int4},
			{Name: secrulepb.RuleEvent_RequestDomain, Type: dbx.Text},
			{Name: secrulepb.RuleEvent_RequestScheme, Type: dbx.Text},
			{Name: secrulepb.RuleEvent_RequestPath, Type: dbx.Text},
			{Name: secrulepb.RuleEvent_RequestQuery, Type: dbx.Text},
			{Name: secrulepb.RuleEvent_RequestIp, Type: dbx.Text},
			{Name: secrulepb.RuleEvent_RequestId, Type: dbx.Text},
			{Name: secrulepb.RuleEvent_Tags, Type: dbx.JSONB},
		},
		Indexes: []dbx.Index{
			{Column: secrulepb.RuleEvent_MatchedTime, Method: dbx.BTree},
			{Column: secrulepb.RuleEvent_ResourceId, Method: dbx.BTree},
		},
	}
}
