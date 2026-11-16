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

import "github.com/sentinez/pkg/szstorage/dbx"

const (
	IAMUsers    dbx.Table = "iam.users"
	IAMAccounts dbx.Table = "iam.accounts"
)

const (
	Tenants              dbx.Table = "tenants"
	TenantResources      dbx.Table = "tenant.resources"
	TenantMembers        dbx.Table = "tenant.members"
	TenantCaCertificates dbx.Table = "tenant.cacertificates"
)

const (
	AnalyticActivities   dbx.Table = "analytic.activities"
	AnalyticSecRules     dbx.Table = "analytic.secrules"
	AnalyticCoreRulesets dbx.Table = "analytic.corerulesets"
	AnalyticRateLimiter  dbx.Table = "analytic.ratelimiters"

	AnalyticSecRulesEvent     dbx.Table = "analytic.secrules.events"
	AnalyticCoreRulesetsEvent dbx.Table = "analytic.corerulesets.events"
	AnalyticRateLimiterEvent  dbx.Table = "analytic.ratelimiters.events"
	AnalyticHttpIngressEvent  dbx.Table = "analytic.httpingress.events"
	AnalyticHttpEgressEvent   dbx.Table = "analytic.httpegress.events"
)

const (
	SecuritySecRules     dbx.Table = "security.secrules"
	SecurityRateLimits   dbx.Table = "security.ratelimits"
	SecurityCoreRulesets dbx.Table = "security.corerulesets"
)
