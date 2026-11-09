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
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
)

// Tenants the schema of the sztable.Tenants table.
func Tenants() dbx.Schema {
	return dbx.Schema{
		Table: sztable.Tenants,
		Columns: []dbx.Column{
			{Name: tenantpb.Tenant_Id, Type: dbx.Text},
			{Name: tenantpb.Tenant_Name, Type: dbx.Text},
			{Name: tenantpb.Tenant_Description, Type: dbx.Text},
			{Name: tenantpb.Tenant_Plan, Type: dbx.Int4},
		},
	}
}

// TenantMembers the schema of the sztable.TenantMembers table.
func TenantMembers() dbx.Schema {
	return dbx.Schema{
		Table: sztable.TenantMembers,
		Columns: []dbx.Column{
			{Name: tenantpb.Member_Id, Type: dbx.Text},
			{Name: tenantpb.Member_UserId, Type: dbx.Text},
			{Name: tenantpb.Member_TenantId, Type: dbx.Text},
			{Name: tenantpb.Member_Roles, Type: dbx.IntArr},
		},
	}
}

// Columns of the sztable.TenantResources table derived from the stored
// server (tenantpb.Resource_Server), to filter the resources by: they are
// written with it, never on their own.
const (
	ResourceServerName   = "server_name"
	ResourceServerDomain = "server_resource_domain"
)

// TenantResources the schema of the sztable.TenantResources table.
func TenantResources() dbx.Schema {
	return dbx.Schema{
		Table: sztable.TenantResources,
		Columns: []dbx.Column{
			{Name: tenantpb.Resource_Id, Type: dbx.Text},
			{Name: tenantpb.Resource_Server, Type: dbx.ByteA},
			{Name: ResourceServerName, Type: dbx.Text},
			{Name: ResourceServerDomain, Type: dbx.Text},
			{Name: tenantpb.Resource_TenantId, Type: dbx.Text},
			{Name: tenantpb.Resource_Status, Type: dbx.Int4},
		},
	}
}
