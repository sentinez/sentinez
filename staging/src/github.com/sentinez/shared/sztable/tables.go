// Copyright 2025 Duc-Hung Ho.
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

package sztable

type Table string

func (t Table) String() string {
	return string(t)
}

const (
	IAMUsers    Table = "iam.users"
	IAMAccounts Table = "iam.accounts"
)

const (
	Tenants         Table = "tenants"
	TenantResources Table = "tenant.resources"
	TenantMembers   Table = "tenant.members"
)

const (
	AnalyticActivities   Table = "analytic.activities"
	AnalyticSecRules     Table = "analytic.secrules"
	AnalyticCoreRulesets Table = "analytic.corerulesets"
	AnalyticRateLimiter  Table = "analytic.ratelimiters"
)

const (
	SecuritySecRules     Table = "security.secrules"
	SecurityRateLimits   Table = "security.ratelimits"
	SecurityCoreRulesets Table = "security.corerulesets"
)
