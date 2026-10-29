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

package szprefix_test

import (
	"regexp"
	"testing"

	"github.com/sentinez/shared/szprefix"
	"github.com/stretchr/testify/assert"
)

var _validPrefix = regexp.MustCompile(`^[a-z]{2,5}$`)

var _all = []szprefix.Prefix{
	szprefix.User, szprefix.Account,
	szprefix.Tenant, szprefix.Member, szprefix.TenantResource,
	szprefix.SecRule, szprefix.Condition, szprefix.RateLimit,
	szprefix.CoreRuleset,
	szprefix.Activity, szprefix.SecRuleEvent, szprefix.CoreRulesetEvent,
	szprefix.Team, szprefix.Role, szprefix.Invitation, szprefix.Session,
	szprefix.APIKey, szprefix.AccessToken, szprefix.SSOConnection,
	szprefix.Project, szprefix.Environment, szprefix.Region, szprefix.Tag,
	szprefix.Instance, szprefix.Cluster, szprefix.Node, szprefix.Service,
	szprefix.Deployment, szprefix.Build, szprefix.Image, szprefix.Function,
	szprefix.Volume, szprefix.Snapshot, szprefix.Bucket, szprefix.Database,
	szprefix.Backup,
	szprefix.VPC, szprefix.Subnet, szprefix.Firewall, szprefix.LoadBalancer,
	szprefix.StaticIP, szprefix.Domain, szprefix.DNSZone,
	szprefix.Certificate,
	szprefix.Secret, szprefix.Policy,
	szprefix.Customer, szprefix.Plan, szprefix.Subscription,
	szprefix.Invoice, szprefix.Payment, szprefix.UsageRecord,
	szprefix.Credit,
	szprefix.Event, szprefix.Job, szprefix.JobRun, szprefix.Webhook,
	szprefix.Notification, szprefix.Alert, szprefix.Incident,
	szprefix.AuditLog, szprefix.Request, szprefix.Ticket,
}

func TestPrefixFormat(t *testing.T) {
	for _, p := range _all {
		assert.Regexp(t, _validPrefix, p.String())
	}
}

func TestPrefixUnique(t *testing.T) {
	seen := make(map[szprefix.Prefix]struct{}, len(_all))
	for _, p := range _all {
		assert.NotContains(t, seen, p)
		seen[p] = struct{}{}
	}
}
