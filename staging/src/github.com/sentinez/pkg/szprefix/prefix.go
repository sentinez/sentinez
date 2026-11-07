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

// Package szprefix holds the id prefix of every entity, TypeID style:
// lowercase a-z only, 2 to 5 characters, one unique prefix per table.
// Prefixes are short, guessable abbreviations (truncated, or with vowels
// dropped). Once exposed through the API a prefix must never change.
//
// An id is the prefix, Separator, then the random part: "user_<uuid>".
package szprefix

// Separator joins a prefix to the random part of an id.
const Separator = "_"

type Prefix string

func (p Prefix) String() string {
	return string(p)
}

// Key returns the leading part of every p id: p followed by Separator
// ("user_"), or "" for None.
func (p Prefix) Key() string {
	if p == None {
		return ""
	}

	return string(p) + Separator
}

const (
	None Prefix = ""

	// ── IAM (iam.users, iam.accounts) ──
	User    Prefix = "user" // user
	Account Prefix = "acct" // login account of a user

	// ── Tenant (tenants, tenant.members, tenant.resources) ──
	Tenant         Prefix = "tnt"  // tenant / organization
	Member         Prefix = "mem"  // membership (user ↔ tenant)
	TenantResource Prefix = "rsrc" // resource owned by a tenant

	// ── Security (security.secrules, .ratelimits, .corerulesets) ──
	SecRule     Prefix = "srule" // WAF SecRule
	Condition   Prefix = "cond"  // SecRule expression condition
	RateLimit   Prefix = "rlim"  // rate limit
	CoreRuleset Prefix = "crs"   // OWASP core ruleset

	// ── Analytic (analytic.activities, .secrules, .corerulesets) ──
	Activity         Prefix = "actv"  // tenant activity
	SecRuleEvent     Prefix = "srevt" // SecRule match event
	CoreRulesetEvent Prefix = "crevt" // core ruleset match event

	// ── Identity & access ──
	Team          Prefix = "team" // team
	Role          Prefix = "role" // role
	Invitation    Prefix = "invt" // invitation
	Session       Prefix = "sess" // login session
	APIKey        Prefix = "key"  // API key
	AccessToken   Prefix = "tok"  // access token
	SSOConnection Prefix = "sso"  // SSO / SAML connection

	// ── Resource organization ──
	Project     Prefix = "proj" // project
	Environment Prefix = "env"  // environment (prod, staging…)
	Region      Prefix = "rgn"  // deployment region
	Tag         Prefix = "tag"  // resource tag

	// ── Compute ──
	Instance   Prefix = "inst"  // virtual machine / instance
	Cluster    Prefix = "clstr" // cluster (k8s…)
	Node       Prefix = "node"  // cluster node
	Service    Prefix = "svc"   // service / application
	Deployment Prefix = "depl"  // deployment
	Build      Prefix = "build" // build
	Image      Prefix = "img"   // container / VM image
	Function   Prefix = "fn"    // serverless function

	// ── Storage & data ──
	Volume   Prefix = "vol"  // block storage volume
	Snapshot Prefix = "snap" // snapshot
	Bucket   Prefix = "bkt"  // object storage bucket
	Database Prefix = "db"   // database instance
	Backup   Prefix = "bkp"  // backup

	// ── Networking ──
	VPC          Prefix = "vpc"  // virtual private cloud
	Subnet       Prefix = "snet" // subnet
	Firewall     Prefix = "fw"   // firewall / security group
	LoadBalancer Prefix = "lb"   // load balancer
	StaticIP     Prefix = "ip"   // static IP address
	Domain       Prefix = "dom"  // domain name
	DNSZone      Prefix = "zone" // DNS zone
	Certificate  Prefix = "cert" // TLS certificate

	// ── Secrets & policies ──
	Secret Prefix = "scrt" // secret value
	Policy Prefix = "pol"  // access policy

	// ── Billing ──
	Customer     Prefix = "cus"   // billing customer
	Plan         Prefix = "plan"  // pricing plan
	Subscription Prefix = "sub"   // subscription
	Invoice      Prefix = "invc"  // invoice
	Payment      Prefix = "pay"   // payment
	UsageRecord  Prefix = "usage" // usage / metering record
	Credit       Prefix = "crdt"  // credit / voucher

	// ── Operations & system ──
	Event        Prefix = "evt"   // event
	Job          Prefix = "job"   // background job
	JobRun       Prefix = "run"   // job run
	Webhook      Prefix = "whk"   // webhook endpoint
	Notification Prefix = "ntf"   // notification
	Alert        Prefix = "alert" // alert
	Incident     Prefix = "inc"   // incident
	AuditLog     Prefix = "aud"   // audit log entry
	Request      Prefix = "req"   // request id (tracing)
	Ticket       Prefix = "tkt"   // support ticket
)
