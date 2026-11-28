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

// Package identity names the services of the internal mTLS. A name is the
// CN of the certificate of the service (make sz.<service>.cert), the name
// its peers verify that certificate by, and the name of its service token
// (szservicetoken -name).
package identity

const (
	// Analytic is the analytic service.
	Analytic = "sentinez.apps.analytic"
	// Communication is the communication service.
	Communication = "sentinez.apps.communication"
	// Greeter is the greeter service.
	Greeter = "sentinez.apps.greeter"
	// IAM is the identity access management service.
	IAM = "sentinez.apps.iam"
	// Security is the security service.
	Security = "sentinez.apps.security"
	// Tenant is the tenant service.
	Tenant = "sentinez.apps.tenant"
	// CentralData is centraldata, the server the edges follow.
	CentralData = "sentinez.apps.centraldata"
	// Edge is an edge; the edges of a cluster share it.
	Edge = "sentinez.dmz.edge"
)
