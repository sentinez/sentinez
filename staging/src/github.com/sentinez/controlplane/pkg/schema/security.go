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
		Table: sztable.SecuritySecRules,
		Columns: []dbx.Column{
			{Name: securitypb.SecRule_Id, Type: dbx.Text},
			{Name: securitypb.SecRule_Name, Type: dbx.Text},
			{Name: securitypb.SecRule_Description, Type: dbx.Text},
			{Name: securitypb.SecRule_Expr, Type: dbx.JSONB},
			{Name: securitypb.SecRule_Status, Type: dbx.Int4},
			{Name: securitypb.SecRule_Priority, Type: dbx.Int4},
			{Name: securitypb.SecRule_Action, Type: dbx.JSONB},
			{Name: securitypb.SecRule_OwnerResourceId, Type: dbx.Text},
			{Name: securitypb.SecRule_OwnerTenantId, Type: dbx.Text},
		},
	}
}

// SecurityRateLimits the schema of the sztable.SecurityRateLimits table.
func SecurityRateLimits() dbx.Schema {
	return dbx.Schema{
		Table: sztable.SecurityRateLimits,
		Columns: []dbx.Column{
			{Name: securitypb.RateLimit_Id, Type: dbx.Text},
			{Name: securitypb.RateLimit_Name, Type: dbx.Text},
			{Name: securitypb.RateLimit_Description, Type: dbx.Text},
			{Name: securitypb.RateLimit_Expr, Type: dbx.JSONB},
			{Name: securitypb.RateLimit_Status, Type: dbx.Int4},
			{Name: securitypb.RateLimit_Priority, Type: dbx.Int4},
			{Name: securitypb.RateLimit_Action, Type: dbx.Text},
			{Name: securitypb.RateLimit_TimeWindow, Type: dbx.Text},
			{Name: securitypb.RateLimit_MaxRequests, Type: dbx.Int8},
			{Name: securitypb.RateLimit_Timeout, Type: dbx.Text},
			{Name: securitypb.RateLimit_OwnerResourceId, Type: dbx.Text},
			{Name: securitypb.RateLimit_OwnerTenantId, Type: dbx.Text},
		},
	}
}

// SecurityCoreRulesets the schema of the sztable.SecurityCoreRulesets table.
func SecurityCoreRulesets() dbx.Schema {
	return dbx.Schema{
		Table: sztable.SecurityCoreRulesets,
		Columns: []dbx.Column{
			{Name: securitypb.CoreRuleset_Id, Type: dbx.Text},
			{Name: securitypb.CoreRuleset_Name, Type: dbx.Text},
			{Name: securitypb.CoreRuleset_Description, Type: dbx.Text},
			{Name: securitypb.CoreRuleset_Content, Type: dbx.JSONB},
			{
				Name: securitypb.CoreRuleset_OwnerResourceId,
				Type: dbx.Text,
			},
			{Name: securitypb.CoreRuleset_OwnerTenantId, Type: dbx.Text},
		},
		// The list filters on content are JSONB containments.
		Indexes: []dbx.Index{
			{Column: securitypb.CoreRuleset_Content, Method: dbx.GIN},
		},
	}
}
