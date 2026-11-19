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
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/sztable"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
)

// SecuritySecRules the schema of the sztable.SecuritySecRules table.
func SecuritySecRules() dbx.Schema {
	return dbx.Schema{
		Table:     sztable.SecuritySecRules,
		CreatedAt: securitypb.SecRule_OwnerCreatedAt,
		UpdatedAt: securitypb.SecRule_OwnerUpdatedAt,
		Columns: []dbx.Column{
			{Name: securitypb.SecRule_Id, Type: dbx.Text},
			{Name: securitypb.SecRule_Name, Type: dbx.Text},
			{Name: securitypb.SecRule_Description, Type: dbx.Text},
			{Name: securitypb.SecRule_Expr, Type: dbx.JSONB},
			{Name: securitypb.SecRule_Status, Type: dbx.Int4},
			{Name: securitypb.SecRule_Priority, Type: dbx.Int4},
			{Name: securitypb.SecRule_Action, Type: dbx.Text},
			{Name: securitypb.SecRule_OwnerResourceId, Type: dbx.Text},
			{Name: securitypb.SecRule_OwnerTenantId, Type: dbx.Text},
		},
	}
}

// SecuritySecRateLimits the schema of the sztable.SecuritySecRateLimits table.
func SecuritySecRateLimits() dbx.Schema {
	return dbx.Schema{
		Table:     sztable.SecuritySecRateLimits,
		CreatedAt: securitypb.SecRateLimit_OwnerCreatedAt,
		UpdatedAt: securitypb.SecRateLimit_OwnerUpdatedAt,
		Columns: []dbx.Column{
			{Name: securitypb.SecRateLimit_Id, Type: dbx.Text},
			{Name: securitypb.SecRateLimit_Name, Type: dbx.Text},
			{Name: securitypb.SecRateLimit_Description, Type: dbx.Text},
			{Name: securitypb.SecRateLimit_Expr, Type: dbx.JSONB},
			{Name: securitypb.SecRateLimit_Status, Type: dbx.Int4},
			{Name: securitypb.SecRateLimit_Priority, Type: dbx.Int4},
			{Name: securitypb.SecRateLimit_Action, Type: dbx.Text},
			{Name: securitypb.SecRateLimit_TimeWindow, Type: dbx.Text},
			{Name: securitypb.SecRateLimit_MaxRequests, Type: dbx.Int8},
			{Name: securitypb.SecRateLimit_Timeout, Type: dbx.Text},
			{Name: securitypb.SecRateLimit_OwnerResourceId, Type: dbx.Text},
			{Name: securitypb.SecRateLimit_OwnerTenantId, Type: dbx.Text},
		},
	}
}

// SecuritySecRulesets the schema of the sztable.SecuritySecRulesets table.
func SecuritySecRulesets() dbx.Schema {
	return dbx.Schema{
		Table:     sztable.SecuritySecRulesets,
		CreatedAt: securitypb.SecRuleset_OwnerCreatedAt,
		UpdatedAt: securitypb.SecRuleset_OwnerUpdatedAt,
		Columns: []dbx.Column{
			{Name: securitypb.SecRuleset_Id, Type: dbx.Text},
			{Name: securitypb.SecRuleset_Name, Type: dbx.Text},
			{Name: securitypb.SecRuleset_Description, Type: dbx.Text},
			{Name: securitypb.SecRuleset_Setting, Type: dbx.JSONB},
			{
				Name: securitypb.SecRuleset_OwnerResourceId,
				Type: dbx.Text,
			},
			{Name: securitypb.SecRuleset_OwnerTenantId, Type: dbx.Text},
		},
		// The list filters on content are JSONB containments.
		Indexes: []dbx.Index{
			{Column: securitypb.SecRuleset_Setting, Method: dbx.GIN},
		},
	}
}
