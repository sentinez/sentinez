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

// Package schema declares the tables of the controlplane repositories: the
// repositories sync them in postgres, centraldata keeps them in go-memdb.
package schema

import "github.com/sentinez/shared/szstorage/dbx"

// All the schemas of every controlplane table.
func All() []dbx.Schema {
	return []dbx.Schema{
		IAMUsers(),
		IAMAccounts(),
		Tenants(),
		TenantMembers(),
		TenantResources(),
		SecuritySecRules(),
		SecurityRateLimits(),
		SecurityCoreRulesets(),
		AnalyticActivities(),
	}
}
