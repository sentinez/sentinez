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

package secure

import (
	"strings"

	"github.com/sentinez/core/chains"
	corehttp "github.com/sentinez/core/http"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	httpxcmn "github.com/sentinez/sentinez/pkg/network/httpx/common"
)

var _ chains.Handler = (*DomainBased)(nil)

func NewDomainBased(hostname string) chains.Handler {
	return &DomainBased{
		BaseHandler: chains.New(),
		hostname:    hostname,
	}
}

type DomainBased struct {
	*chains.BaseHandler
	hostname string
}

func (d *DomainBased) Handle(ctx corehttp.Context) error {
	// zlog.Debug("[edge] >>> visit domain")

	ns, ok := d.isValidSingleLevelSubdomain(ctx.Host(), d.hostname)
	if !ok {
		return httpxcmn.Forbidden(ctx)
	}

	ctxValue, ok := corehttp.GetRequestContext(ctx)
	if !ok {
		ctxValue = &edgepb.Context{}
	}

	ctxValue.TenantNs = ns

	ctx = corehttp.SetRequestContext(ctx, ctxValue)

	return d.HandleNext(ctx)
}

func (d *DomainBased) isValidSingleLevelSubdomain(
	subdomain, root string) (string, bool) {

	// Remove port if present
	if colon := strings.IndexByte(subdomain, ':'); colon >= 0 {
		subdomain = subdomain[:colon]
	}

	// Check if it ends with "." + root
	suffix := "." + root
	if !strings.HasSuffix(subdomain, suffix) {
		return "", false
	}

	// Extract subdomain part (before root)
	subPart := subdomain[:len(subdomain)-len(suffix)]

	// Ensure single-level (no extra dots) and non-empty
	if subPart == "" || strings.IndexByte(subPart, '.') >= 0 {
		return "", false
	}

	return subPart, true
}
