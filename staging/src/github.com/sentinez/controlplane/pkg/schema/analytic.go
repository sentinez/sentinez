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
	"time"

	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/timescale"
	"github.com/sentinez/pkg/sztable"
	analyticpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	crspb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
)

const (
	// EventRetention how long the events, and their histograms, are kept.
	EventRetention = 30 * 24 * time.Hour

	_eventChunkInterval = 24 * time.Hour
)

// EventHypertable the hypertable of the event table s, partitioned by
// timeColumn, a column of unix milliseconds, kept for EventRetention.
func EventHypertable(s dbx.Schema,
	timeColumn string) timescale.Hypertable {
	return timescale.Hypertable{
		Schema:        s,
		TimeColumn:    timeColumn,
		ChunkInterval: _eventChunkInterval,
		Retention:     EventRetention,
	}
}

// resourceTime the index of the events of a resource over a time range: one
// ordered scan of the range per chunk. The request ids of the http events
// are served by the unique (id, time) key of their hypertable.
func resourceTime(resourceColumn, timeColumn string) dbx.Index {
	return dbx.Index{
		Columns: []string{resourceColumn, timeColumn},
		Method:  dbx.BTree,
	}
}

// AnalyticActivities the schema of the sztable.AnalyticActivities table.
func AnalyticActivities() dbx.Schema {
	return dbx.Schema{
		Table:     sztable.AnalyticActivities,
		CreatedAt: analyticpb.Activity_OwnerCreatedAt,
		UpdatedAt: analyticpb.Activity_OwnerUpdatedAt,
		Columns: []dbx.Column{
			{Name: analyticpb.Activity_Id, Type: dbx.Text},
			{Name: analyticpb.Activity_OwnerResourceId, Type: dbx.Text},
			{Name: analyticpb.Activity_OwnerTenantId, Type: dbx.Text},
			{Name: analyticpb.Activity_UniqueVisitor, Type: dbx.IntArr},
		},
	}
}

// AnalyticSecRulesEvent the schema of the sztable.AnalyticSecRulesEvent
// table: the secrulepb.SecRuleEvent records of the sztopic.EventHTTPSecRule
// topic.
func AnalyticSecRulesEvent() dbx.Schema {
	return secRuleEvents(sztable.AnalyticSecRulesEvent)
}

// AnalyticRateLimiterEvent the schema of the
// sztable.AnalyticRateLimiterEvent table: the secrulepb.SecRuleEvent records
// of the sztopic.EventHTTPRateLimit topic.
func AnalyticRateLimiterEvent() dbx.Schema {
	return secRuleEvents(sztable.AnalyticRateLimiterEvent)
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
			resourceTime(crspb.CoreRuleEvent_ResourceId,
				crspb.CoreRuleEvent_MatchedTime),
			{Column: crspb.CoreRuleEvent_RequestId, Method: dbx.BTree},
		},
	}
}

// AnalyticHttpIngressEvent the schema of the
// sztable.AnalyticHttpIngressEvent table: the httppb.IngressEvent records
// of the sztopic.EventHTTPIngress topic, one per request of the edge.
func AnalyticHttpIngressEvent() dbx.Schema {
	return dbx.Schema{
		Table: sztable.AnalyticHttpIngressEvent,
		Columns: []dbx.Column{
			{Name: httppb.IngressEvent_Id, Type: dbx.Text},
			{Name: httppb.IngressEvent_Method, Type: dbx.Text},
			{Name: httppb.IngressEvent_Scheme, Type: dbx.Text},
			{Name: httppb.IngressEvent_Host, Type: dbx.Text},
			{Name: httppb.IngressEvent_Path, Type: dbx.Text},
			{Name: httppb.IngressEvent_Queries, Type: dbx.Text},
			{Name: httppb.IngressEvent_Headers, Type: dbx.Text},
			{Name: httppb.IngressEvent_Protocol, Type: dbx.Text},
			{Name: httppb.IngressEvent_RemoteAddress, Type: dbx.Text},
			{Name: httppb.IngressEvent_Status, Type: dbx.Int4},
			{Name: httppb.IngressEvent_UserAgent, Type: dbx.Text},
			{Name: httppb.IngressEvent_Asn, Type: dbx.Int8},
			{Name: httppb.IngressEvent_Country, Type: dbx.Text},
			{Name: httppb.IngressEvent_ContentType, Type: dbx.Text},
			{Name: httppb.IngressEvent_Timestamp, Type: dbx.Int8},
			{Name: httppb.IngressEvent_Fingerprint, Type: dbx.Text},
			{Name: httppb.IngressEvent_ClientIp, Type: dbx.Text},
			{Name: httppb.IngressEvent_ResourceId, Type: dbx.Text},
			{Name: httppb.IngressEvent_TenantId, Type: dbx.Text},
		},
		Indexes: []dbx.Index{
			{Column: httppb.IngressEvent_Timestamp, Method: dbx.BTree},
			resourceTime(httppb.IngressEvent_ResourceId,
				httppb.IngressEvent_Timestamp),
		},
	}
}

// AnalyticHttpEgressEvent the schema of the sztable.AnalyticHttpEgressEvent
// table: the httppb.EgressEvent records of the sztopic.EventHTTPEgress
// topic, one per request the edge proxied to its upstream origin.
func AnalyticHttpEgressEvent() dbx.Schema {
	return dbx.Schema{
		Table: sztable.AnalyticHttpEgressEvent,
		Columns: []dbx.Column{
			{Name: httppb.EgressEvent_Id, Type: dbx.Text},
			{Name: httppb.EgressEvent_Upstream, Type: dbx.Text},
			{Name: httppb.EgressEvent_Method, Type: dbx.Text},
			{Name: httppb.EgressEvent_Scheme, Type: dbx.Text},
			{Name: httppb.EgressEvent_Path, Type: dbx.Text},
			{Name: httppb.EgressEvent_Queries, Type: dbx.Text},
			{Name: httppb.EgressEvent_Status, Type: dbx.Int4},
			{Name: httppb.EgressEvent_ContentType, Type: dbx.Text},
			{Name: httppb.EgressEvent_ContentLength, Type: dbx.Int8},
			{Name: httppb.EgressEvent_LatencyMicros, Type: dbx.Int8},
			{Name: httppb.EgressEvent_Timestamp, Type: dbx.Int8},
			{Name: httppb.EgressEvent_ResourceId, Type: dbx.Text},
			{Name: httppb.EgressEvent_TenantId, Type: dbx.Text},
			{Name: httppb.EgressEvent_Protocol, Type: dbx.Text},
			{Name: httppb.EgressEvent_Headers, Type: dbx.Text},
		},
		Indexes: []dbx.Index{
			{Column: httppb.EgressEvent_Timestamp, Method: dbx.BTree},
			resourceTime(httppb.EgressEvent_ResourceId,
				httppb.EgressEvent_Timestamp),
		},
	}
}

// secRuleEvents the schema of a table of secrulepb.SecRuleEvent.
func secRuleEvents(table sztable.Table) dbx.Schema {
	return dbx.Schema{
		Table: table,
		Columns: []dbx.Column{
			{Name: secrulepb.SecRuleEvent_Id, Type: dbx.Text},
			{Name: secrulepb.SecRuleEvent_MatchedSecruleIds, Type: dbx.TextArr},
			{Name: secrulepb.SecRuleEvent_Messages, Type: dbx.TextArr},
			{Name: secrulepb.SecRuleEvent_ResourceId, Type: dbx.Text},
			{Name: secrulepb.SecRuleEvent_TenantId, Type: dbx.Text},
			{Name: secrulepb.SecRuleEvent_MatchedTime, Type: dbx.Int8},
			{Name: secrulepb.SecRuleEvent_Action, Type: dbx.JSONB},
			{Name: secrulepb.SecRuleEvent_Service, Type: dbx.Int4},
			{Name: secrulepb.SecRuleEvent_RequestDomain, Type: dbx.Text},
			{Name: secrulepb.SecRuleEvent_RequestScheme, Type: dbx.Text},
			{Name: secrulepb.SecRuleEvent_RequestPath, Type: dbx.Text},
			{Name: secrulepb.SecRuleEvent_RequestQuery, Type: dbx.Text},
			{Name: secrulepb.SecRuleEvent_RequestIp, Type: dbx.Text},
			{Name: secrulepb.SecRuleEvent_RequestId, Type: dbx.Text},
			{Name: secrulepb.SecRuleEvent_Tags, Type: dbx.JSONB},
		},
		Indexes: []dbx.Index{
			{Column: secrulepb.SecRuleEvent_MatchedTime, Method: dbx.BTree},
			resourceTime(secrulepb.SecRuleEvent_ResourceId,
				secrulepb.SecRuleEvent_MatchedTime),
			{Column: secrulepb.SecRuleEvent_RequestId, Method: dbx.BTree},
		},
	}
}
